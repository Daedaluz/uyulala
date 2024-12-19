package cmd

import (
	"time"
	"uyulala/cmd/wait-db"

	"github.com/spf13/cobra"
)

// waitDBCmd represents the waitDb command
var waitDBCmd = &cobra.Command{
	Use:   "wait-db",
	Short: "Wait for the database to become ready",
	Run:   waitdb.Main,
}

func init() {
	rootCmd.AddCommand(waitDBCmd)
	waitDBCmd.Flags().DurationP("timeout", "t", time.Minute, "Time to wait for db before exiting with error")
}
