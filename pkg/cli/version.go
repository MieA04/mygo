package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of MyGo",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("MyGo v0.1.0-alpha")
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
