package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Remove temporary build artifacts",
	RunE: func(cmd *cobra.Command, args []string) error {
		tempPattern := filepath.Join(os.TempDir(), "mygo-build-*")
		matches, err := filepath.Glob(tempPattern)
		if err != nil {
			return err
		}

		for _, match := range matches {
			fmt.Printf("Removing %s\n", match)
			os.RemoveAll(match)
		}
		fmt.Println("Clean complete.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(cleanCmd)
}
