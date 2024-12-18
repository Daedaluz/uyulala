package cmd

import (
	"time"
	"uyulala/cmd/available"

	"github.com/spf13/cobra"
)

// availableCmd represents the available command
var availableCmd = &cobra.Command{
	Use:   "available",
	Short: "check if the server is available and serving",
	Run:   available.Main,
}

func init() {
	rootCmd.AddCommand(availableCmd)
	availableCmd.Flags().BoolP("wait", "w", false, "Wait for the service to be available")
	availableCmd.Flags().DurationP("timeout", "t", time.Minute, "Time to wait for service before exiting with error")
}
