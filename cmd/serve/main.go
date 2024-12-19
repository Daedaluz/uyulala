package serve

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path"
	"runtime/debug"
	"syscall"
	"time"
	"uyulala/internal/api/v1"
	"uyulala/internal/db/migrations"
	"uyulala/internal/mds"
	"uyulala/internal/trust"
	wellknown "uyulala/internal/well-known"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gitlab.com/daedaluz/gindb"
	"golang.org/x/net/context"
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
		urlPath := c.Request.URL.Path
		logger.Info("Request", slog.String("path", urlPath),
			slog.String("method", method),
			slog.String("clientIP", clientIP),
			slog.Int("statusCode", statusCode),
			slog.Duration("latency", latency))
	}
}

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
		c.Header("Cache-Control", viper.GetString("http.cacheControl"))
		c.Header("Referer-Policy", viper.GetString("http.refererPolicy"))
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

	engine := setupGinEngine(db)

	if issuer := viper.GetString("userApi.trustedIssuer"); issuer != "" {
		if err := trust.Configure(issuer); err != nil {
			slog.Error("Failed to configure trust", "error", err)
		} else {
			slog.Info("Trust configured")
		}
	}

	go mds.Init()

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
		if viper.GetBool("tls.enable") {
			loadCerts(server)
			ln, err := net.Listen("tcp", viper.GetString("http.addr"))
			if err != nil {
				slog.Error("Couldn't start server", "error", err)
				return
			}
			if err := server.ServeTLS(ln, "", ""); err != nil {
				slog.Error("TLS server returned error", "error", err)
			} else {
				slog.Info("Byte TLS.")
			}
		} else {
			if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				slog.Error("Server returned error", "error", err)
			} else {
				slog.Info("Bye!")
			}
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

func loadCerts(server *http.Server) {
	var cert tls.Certificate
	var err error
	if viper.GetBool("tls.generate") {
		cert, err = generateCerts()
		if err != nil {
			slog.Error("Couldn't generate certs", "error", err)
			os.Exit(1)
		}
		server.TLSConfig = &tls.Config{
			MinVersion:   tls.VersionTLS12,
			Certificates: []tls.Certificate{cert},
		}
	} else {
		cert, err = tls.LoadX509KeyPair(viper.GetString("tls.cert"), viper.GetString("tls.key"))
		if err != nil {
			slog.Error("Couldn't load certs", "error", err)
			os.Exit(1)
		}
		server.TLSConfig = &tls.Config{
			MinVersion:   tls.VersionTLS12,
			Certificates: []tls.Certificate{cert},
		}
	}
}

func generateCerts() (tls.Certificate, error) {
	key, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		return tls.Certificate{}, err
	}
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return tls.Certificate{}, err
	}
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:         "TestCertificate",
			Organization:       []string{"Uyulala"},
			Country:            []string{"SE"},
			OrganizationalUnit: []string{"IDP"},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().AddDate(1, 0, 0),
		KeyUsage:  x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
		},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	for _, domain := range viper.GetStringSlice("webauthn.origins") {
		x, err := url.Parse(domain)
		if err != nil {
			continue
		}
		template.DNSNames = append(template.DNSNames, x.Hostname())
	}

	ipAddresses, err := net.InterfaceAddrs()
	if err != nil {
		return tls.Certificate{}, err
	}
	for _, addr := range ipAddresses {
		ip, _, err := net.ParseCIDR(addr.String())
		if err != nil {
			return tls.Certificate{}, err
		}
		template.IPAddresses = append(template.IPAddresses, ip)
	}
	crt, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		return tls.Certificate{}, err
	}
	return tls.Certificate{
		Certificate: [][]byte{crt},
		PrivateKey:  key,
	}, nil
}
