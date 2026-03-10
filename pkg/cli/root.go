package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "mygo",
	Short: "MyGo Compiler",
	Long:  `MyGo is a modern statically typed programming language designed to combine the simplicity of Go with advanced type system features.`,
	Run: func(cmd *cobra.Command, args []string) {
		// If no arguments, print help
		if len(args) == 0 {
			cmd.Help()
			return
		}
		// Fallback to transpile if arguments are provided (backward compatibility)
		// But ideally we should encourage using subcommands.
		// For now, let's just print help to encourage standard usage.
		cmd.Help()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// Global flags can be defined here
}
