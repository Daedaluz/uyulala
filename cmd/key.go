package cmd

import (
	"uyulala/cmd/create/key"

	"github.com/spf13/cobra"
)

// keyCmd represents the key command
var keyCmd = &cobra.Command{
	Use:   "key",
	Short: "Create a server key",
	Long:  `Create a server key that will be used to sign the id_token JWTs.`,
	Run:   key.Main,
}

func init() {
	createCmd.AddCommand(keyCmd)
	key.KeyAlg = keyCmd.Flags().StringP("alg", "a", "RS256", "Key algorithm")
}
