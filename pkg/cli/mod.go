package cli

import (
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var modCmd = &cobra.Command{
	Use:   "mod [arguments]",
	Short: "Module maintenance (wrapper around go mod)",
	DisableFlagParsing: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Pass-through to go mod
		c := exec.Command("go", append([]string{"mod"}, args...)...)
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		c.Stdin = os.Stdin
		return c.Run()
	},
}

func init() {
	rootCmd.AddCommand(modCmd)
}
