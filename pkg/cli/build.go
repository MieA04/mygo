package cli

import (
	"path/filepath"

	"github.com/miea04/mygo/pkg/build"
	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build [flags] <source.mygo|directory>",
	Short: "Build MyGo source code",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		sourcePath := args[0]
		outputFile, _ := cmd.Flags().GetString("output")
		rootDir, _ := cmd.Flags().GetString("root")
		keepWorkDir, _ := cmd.Flags().GetBool("keep-work-dir")
		goos, _ := cmd.Flags().GetString("os")
		goarch, _ := cmd.Flags().GetString("arch")

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
			OutputPath:  outputFile,
			KeepWorkDir: keepWorkDir,
			RootDir:     absRoot,
			GOOS:        goos,
			GOARCH:      goarch,
		}

		builder := build.NewBuilder(opts)
		return builder.Build()
	},
}

func init() {
	buildCmd.Flags().StringP("output", "o", "", "Output executable file path")
	buildCmd.Flags().String("root", ".", "Root directory for package resolution")
	buildCmd.Flags().Bool("keep-work-dir", false, "Keep temporary workspace directory")
	buildCmd.Flags().String("os", "", "Target operating system (GOOS)")
	buildCmd.Flags().String("arch", "", "Target architecture (GOARCH)")
	rootCmd.AddCommand(buildCmd)
}
