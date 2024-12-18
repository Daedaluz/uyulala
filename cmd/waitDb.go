package cmd

import (
	"time"
	wait_db "uyulala/cmd/wait-db"

	"github.com/spf13/cobra"
)

// waitDbCmd represents the waitDb command
var waitDbCmd = &cobra.Command{
	Use:   "wait-db",
	Short: "Wait for the database to become ready",
	Run:   wait_db.Main,
}

func init() {
	rootCmd.AddCommand(waitDbCmd)
	waitDbCmd.Flags().DurationP("timeout", "t", time.Minute, "Time to wait for db before exiting with error")
}
