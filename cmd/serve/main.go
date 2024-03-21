package serve

import (
	"errors"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-webauthn/webauthn/metadata"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gitlab.com/daedaluz/gindb"
	"golang.org/x/net/context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path"
	"runtime/debug"
	"syscall"
	"time"
	"uyulala/internal/api/v1"
	"uyulala/internal/db/migrations"
	"uyulala/internal/trust"
	wellknown "uyulala/internal/well-known"
)

func logger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		end := time.Now()
		latency := end.Sub(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		path := c.Request.URL.Path
		logger.Info("Request", slog.String("path", path),
			slog.String("method", method),
			slog.String("clientIP", clientIP),
			slog.Int("statusCode", statusCode),
			slog.Duration("latency", latency))
	}
}

//func customCORS() gin.HandlerFunc {
//	return func(c *gin.Context) {
//		if c.Request.Method == "OPTIONS" {
//			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
//			c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
//			//nolint: lll
//			c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Err, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, Cookie")
//			c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length, Cookie")
//			c.Writer.Header().Set("Access-Control-Max-Age", "0")
//		}
//		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
//		c.Next()
//	}
//}

func versionHandler(c *gin.Context) {
	binfo, ok := debug.ReadBuildInfo()
	if !ok {
		c.JSON(http.StatusOK, gin.H{
			"revision": "unknown",
		})
		return
	}

	res := gin.H{}
	for _, v := range binfo.Settings {
		switch v.Key {
		case "vcs.revision":
			res["revision"] = v.Value
		case "vcs.time":
			res["time"] = v.Value
		case "vcs.modified":
			res["dirty"] = v.Value
		}
	}
	res["version"] = binfo.Main.Version

	c.JSON(http.StatusOK, res)
}

func populateMDS() {
	slog.Info("Populating fido alliance metadata...")
	if err := metadata.PopulateMetadata(viper.GetString("webauthn.mds3")); err != nil {
		slog.Error("Failed to populate metadata", "error", err)
		os.Exit(1)
	}
	slog.Info("Populated!")
}

func prepareDatabase() *sqlx.DB {
	db, err := gindb.Open("mysql", viper.GetString("database.dsn"))
	if err != nil {
		slog.Error("Couldn't open the database", "error", err)
		os.Exit(1)
	}
	slog.Info("Waiting for database to be ready")
	if err := gindb.WaitForDB(db, 60*time.Second); err != nil {
		slog.Error("Database did not become available within 60 seconds", "error", err)
		os.Exit(1)
	}
	slog.Info("Database is ready")
	slog.Info("Migrating database")
	if err := gindb.Migrate(db, migrations.Migrations); err != nil {
		slog.Error("Migration error", "error", err)
		os.Exit(1)
	}
	slog.Info("Database migrated")
	return db
}

func setupGinEngine(db *sqlx.DB) *gin.Engine {
	engine := gin.New()
	engine.Use(gin.Recovery(),
		logger(slog.Default()),
		static.Serve("/", static.LocalFile(viper.GetString("http.staticPath"), true)),
		gindb.MiddlewareDB(db),
		gindb.MiddlewareTX(nil),
	)

	engine.NoRoute(func(c *gin.Context) {
		staticPath := fmt.Sprintf("%s/index.html", path.Clean(viper.GetString("http.staticPath")))
		slog.Info("NoRoute", slog.String("path", c.Request.URL.Path))
		slog.Info("StaticPath", slog.String("path", staticPath))
		c.File(staticPath)
	})

	root := engine.Group("/")
	wknown := root.Group("/", cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowHeaders:     []string{"Authorization", "*"},
		AllowCredentials: true,
	}))
	wellknown.AddRoutes(wknown)
	wknown.GET("/api/version", versionHandler)
	wknown.OPTIONS("/api/version", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	r := engine.Group("/api/v1")
	v1.AddRoutes(r)
	return engine
}

func Main(cmd *cobra.Command, args []string) {
	// Migrate database
	gin.SetMode(gin.ReleaseMode)

	db := prepareDatabase()
	populateMDS()

	engine := setupGinEngine(db)

	if issuer := viper.GetString("userApi.trustedIssuer"); issuer != "" {
		if err := trust.Configure(issuer); err != nil {
			slog.Error("Failed to configure trust", "error", err)
		} else {
			slog.Info("Trust configured")
		}
	}

	server := &http.Server{
		Addr:              viper.GetString("http.addr"),
		Handler:           engine,
		ReadTimeout:       viper.GetDuration("http.readTimeout"),
		ReadHeaderTimeout: viper.GetDuration("http.readHeaderTimeout"),
		WriteTimeout:      viper.GetDuration("http.writeTimeout"),
		IdleTimeout:       viper.GetDuration("http.idleTimeout"),
		MaxHeaderBytes:    viper.GetInt("http.maxHeaderBytes"),
		ErrorLog:          slog.NewLogLogger(slog.Default().Handler(), slog.LevelError),
	}

	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	slog.Info("Starting server")
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Server returned error", "error", err)
		} else {
			slog.Info("Bye!")
		}
	}()
	slog.Info("Server started", "addr", viper.GetString("http.addr"))
	<-sigch
	slog.Info("Shutting down server")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Server shutdown error", "error", err)
	} else {
		slog.Info("Server shutdown")
	}
}
