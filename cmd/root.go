package cmd

import (
	"log/slog"
	"net/http"
	"os"
	"uyulala/internal/mysqlslog"

	"github.com/go-sql-driver/mysql"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "uyulala",
	Short: "",
	Long:  ``,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	_ = mysql.SetLogger(&mysqlslog.Logger{Logger: logger})
	slog.SetDefault(logger)

	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is /etc/uyulala/uyulala.yaml)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Search config in current directory.
		viper.AddConfigPath(".")
		// Search config in /etc
		viper.AddConfigPath("/etc/uyulala")
		viper.SetConfigType("yaml")
		viper.SetConfigName("uyulala")
	}

	// Set defaults
	viper.SetDefault("log.level", 0)
	viper.SetDefault("log.add_source", false)

	viper.SetDefault("http.addr", ":8080")
	viper.SetDefault("http.staticPath", "frontend/dist")
	viper.SetDefault("http.readTimeout", "5s")
	viper.SetDefault("http.readHeaderTimeout", "5s")
	viper.SetDefault("http.writeTimeout", "5s")
	viper.SetDefault("http.idleTimeout", "5s")
	viper.SetDefault("http.maxHeaderBytes", http.DefaultMaxHeaderBytes)
	viper.SetDefault("http.cache_control", "no-cache, no-store, must-revalidate")
	viper.SetDefault("http.referer_policy", "origin")

	viper.SetDefault("challenge.max_time_diff", "5s")

	viper.AutomaticEnv() // read in environment variables that match
	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.Level(viper.GetInt("log.level")),
		}))
		_ = mysql.SetLogger(&mysqlslog.Logger{Logger: logger})
		slog.SetDefault(logger)
	} else {
		slog.Error("Unable to read configuration", "error", err)
	}
}
