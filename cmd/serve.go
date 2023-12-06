package cmd

import (
	"uyulala/cmd/serve"

	"github.com/spf13/cobra"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start server",
	Long:  ``,
	Run:   serve.Main,
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
