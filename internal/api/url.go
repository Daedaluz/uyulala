package api

import (
	"log/slog"
	"net/url"
	"uyulala/internal/db/appdb"
)

func AllowedRedirect(app *appdb.Application, redirect string) bool {
	for _, r := range app.RedirectURI {
		appURL, err := url.Parse(r)
		if err != nil {
			slog.Error("allowedRedirect", "error", err, "app", app.ID, "redirect", redirect, "allowed", r)
			continue
		}
		redirectURL, err := url.Parse(redirect)
		if err != nil {
			slog.Error("allowedRedirect", "error", err, "app", app.ID, "redirect", redirect, "allowed", r)
			continue
		}
		appURL.RawQuery = ""
		redirectURL.RawQuery = ""
		appURL.RawFragment = ""
		redirectURL.RawFragment = ""

		if appURL.String() == redirectURL.String() {
			return true
		}
	}
	return false
}
