package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"runtime/debug"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		binfo, ok := debug.ReadBuildInfo()
		if !ok {
			fmt.Println("No build info available")
			return
		}
		fmt.Println("Version: ", binfo.Main.Version)
		for _, v := range binfo.Settings {
			switch v.Key {
			case "vcs.revision":
				fmt.Println("Revision: ", v.Value)
			case "vcs.time":
				fmt.Println("Build time: ", v.Value)
			case "vcs.modified":
				fmt.Println("Dirty: ", v.Value)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
