package waitdb

import (
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gitlab.com/daedaluz/gindb"
)

func Main(cmd *cobra.Command, args []string) {
	db, err := gindb.Open("mysql", viper.GetString("database.dsn"))
	if err != nil {
		slog.Error("Couldn't open the database", "error", err)
		os.Exit(1)
	}
	slog.Info("Waiting for database to be ready")
	d, err := cmd.Flags().GetDuration("timeout")
	if err != nil {
		slog.Error("Parse duration error", "err", err)
		os.Exit(1)
	}
	if err := gindb.WaitForDB(db, d); err != nil {
		slog.Error("Database did not become available", "error", err, "duration", d)
		os.Exit(1)
	}
	slog.Info("Database is ready")
	os.Exit(0)
}
