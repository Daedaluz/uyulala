package app

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gitlab.com/daedaluz/gindb"
	"log/slog"
	"os"
	"uyulala/internal/db/keydb"
)

var (
	Urls        *[]string
	Description *string
	Icon        *string
	AppID       *string
	Secret      *string
	Demo        *bool
	Alg         *string
	KeyID       *string
	Admin       *bool
)

func Main(_ *cobra.Command, args []string) {
	db, err := gindb.Connect("mysql", viper.GetString("database.dsn"))
	if err != nil {
		slog.Error("Couldn't connect to database", "error", err)
		os.Exit(1)
	}
	slog.Info("Mysql", "dsn", viper.GetString("database.dsn"))

	name := args[0]

	tx, err := db.Beginx()
	if err != nil {
		slog.Error("Create app begin", "error", err)
		os.Exit(1)
	}

	if *Demo {
		*AppID = "demo"
		*Secret = "demo"
		*Description = "Demo application"
		*Icon = "https://www.svgrepo.com/download/341627/auth0.svg"
		*Urls = []string{}
		*Alg = "RS256"
		*Admin = true
		*Urls = append(*Urls,
			"http://localhost:5173/demo",
			"https://localhost:5173/demo",
			"http://localhost/demo",
			"http://localhost:3000/login/generic_oauth",
			"https://localhost/demo",
			"https://localhost:8080/demo",
			"https://oauthdebugger.com/debug",
			"https://oauth.tools/callback/code",
		)
	}

	var kid = *KeyID
	if kid != "" {
		srvKey := &keydb.ServerKey{}
		if err := tx.Select(srvKey, `call get_server_key(?)`, kid); err != nil {
			slog.Error("Couldn't get server key", "error", err, "kid", kid)
			_ = tx.Rollback()
			os.Exit(1)
		}
		*Alg = srvKey.Algorithm
	}
	if kid == "" {
		srvKey := &keydb.ServerKey{}
		if err := tx.Get(srvKey, `call get_server_key_with_alg(?)`, *Alg); err != nil {
			slog.Error("Couldn't get server key", "error", err, "alg", *Alg)
			_ = tx.Rollback()
			os.Exit(1)
		}
		kid = srvKey.ID
	}

	res, err := tx.Queryx(`call create_app(?, ?, ?, ?, ?, ?, ?, ?)`, *AppID, *Secret, name, *Description, *Icon, *Alg, kid, *Admin)
	if err != nil {
		slog.Error("Create app query", "error", err)
		_ = tx.Rollback()
		os.Exit(1)
	}
	var appID, appSecret string
	res.Next()
	if err := res.Scan(&appID, &appSecret); err != nil {
		slog.Error("Create app scan", "error", err)
		_ = tx.Rollback()
		os.Exit(1)
	}
	_ = res.Close()

	for _, url := range *Urls {
		if _, err := tx.Exec(`call create_app_redirect_url(?, ?)`, appID, url); err != nil {
			slog.Error("Add redirect url to app", "error", err)
			_ = tx.Rollback()
			os.Exit(1)
		}
	}
	err = tx.Commit()
	if err != nil {
		slog.Error("Create app error", "error", err)
	}
	slog.Info("Created app", "appId", appID, "appSecret", appSecret)
}
