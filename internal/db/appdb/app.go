package appdb

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	"gitlab.com/daedaluz/gindb"
	"time"
)

type Application struct {
	ID                   string    `json:"id" db:"id"`
	Name                 string    `json:"name" db:"name"`
	Created              time.Time `json:"created" db:"created"`
	Secret               string    `json:"-" db:"secret"`
	Description          string    `json:"description" db:"description"`
	Icon                 string    `json:"icon" db:"icon"`
	IDTokenAlg           string    `json:"idTokenAlg" db:"alg"`
	KeyID                string    `json:"keyId" db:"kid"`
	Admin                bool      `json:"admin" db:"is_admin"`
	RedirectURI          []string  `json:"-"`
	CIBAMode             string    `json:"-" db:"ciba_mode"`
	NotificationEndpoint string    `json:"-" db:"notification_endpoint"`
}

func GetApplication(ctx *gin.Context, appID string) (*Application, error) {
	app := &Application{}
	tx := gindb.GetTX(ctx)
	res, err := tx.Queryx(`call get_app(?)`, appID)
	if err != nil {
		return nil, err
	}
	if !res.Next() {
		_ = res.Close()
		return nil, sql.ErrNoRows
	}
	if err := res.StructScan(app); err != nil {
		_ = res.Close()
		return nil, err
	}
	res.Close()

	res, err = tx.Queryx(`call get_app_redirect_urls(?)`, appID)
	if err != nil {
		return nil, err
	}
	defer res.Close()
	for res.Next() {
		var url string
		if err := res.Scan(&url); err != nil {
			return nil, err
		}
		app.RedirectURI = append(app.RedirectURI, url)
	}
	return app, nil
}
