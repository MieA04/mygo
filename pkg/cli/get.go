package cli

import (
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get [packages]",
	Short: "Add dependencies to current module and install them (wrapper around go get)",
	DisableFlagParsing: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Pass-through to go get
		c := exec.Command("go", append([]string{"get"}, args...)...)
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		c.Stdin = os.Stdin
		return c.Run()
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
}
