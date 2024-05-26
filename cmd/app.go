package cmd

import (
	"github.com/spf13/cobra"
	"uyulala/cmd/create/app"
	"uyulala/internal/db"
)

// appCmd represents the app command
var appCmd = &cobra.Command{
	Use:   "app",
	Short: "Create a new application",
	Long:  ``,
	Args:  cobra.ExactArgs(1),
	Run:   app.Main,
}

func init() {
	createCmd.AddCommand(appCmd)
	app.AppID = appCmd.Flags().StringP("id", "i", db.GenerateUUID(), "Application ID (Default is a randomized UUID)")
	app.Secret = appCmd.Flags().StringP("secret", "s", db.GenerateUUID(), "Application secret (Default is a randomized UUID)")
	app.Description = appCmd.Flags().StringP("desc", "d", "", "Application description")
	app.Icon = appCmd.Flags().StringP("icon", "c", "", "Application icon")
	app.Urls = appCmd.PersistentFlags().StringSliceP("url", "u", []string{}, "Accepted Redirect urls for this client")
	app.Demo = appCmd.Flags().Bool("demo", false, "Create a demo application")
	app.Alg = appCmd.Flags().StringP("alg", "l", "RS256", "Algorithm to use for signing tokens")
	app.KeyID = appCmd.Flags().StringP("kid", "k", "", "Key ID to use for signing tokens")
	app.Admin = appCmd.Flags().Bool("admin", false, "Make this application an admin application")
	app.CIBAMode = appCmd.Flags().String("ciba", "poll", "CIBA mode for this client (poll, push, ping)")
	app.CIBANotificationEndpoint = appCmd.Flags().String("notification", "", "Endpoint to send CIBA notifications")
}
