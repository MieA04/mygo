package cli

import (
	"path/filepath"

	"github.com/miea04/mygo/pkg/build"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run [flags] <source.mygo|directory> [args...]",
	Short: "Compile and run MyGo program",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		sourcePath := args[0]
		programArgs := args[1:]
		
		rootDir, _ := cmd.Flags().GetString("root")
		keepWorkDir, _ := cmd.Flags().GetBool("keep-work-dir")

		// Resolve absolute paths
		absRoot, err := filepath.Abs(rootDir)
		if err != nil {
			return err
		}
		absSource, err := filepath.Abs(sourcePath)
		if err != nil {
			return err
		}

		opts := &build.Options{
			SourcePath:  absSource,
			KeepWorkDir: keepWorkDir,
			RootDir:     absRoot,
		}

		builder := build.NewBuilder(opts)
		return builder.Run(programArgs)
	},
}

func init() {
	runCmd.Flags().String("root", ".", "Root directory for package resolution")
	runCmd.Flags().Bool("keep-work-dir", false, "Keep temporary workspace directory")
	rootCmd.AddCommand(runCmd)
}
