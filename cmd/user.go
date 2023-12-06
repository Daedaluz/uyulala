package cmd

import (
	"uyulala/cmd/create/user"

	"github.com/spf13/cobra"
)

// userCmd represents the user command
var userCmd = &cobra.Command{
	Use:   "user",
	Short: "Create a new user",
	Long:  `Create user`,
	Run:   user.Main,
}

func init() {
	createCmd.AddCommand(userCmd)
}
