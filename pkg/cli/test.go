package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/miea04/mygo/pkg/build"
	"github.com/miea04/mygo/pkg/compiler"
	"github.com/spf13/cobra"
)

var testCmd = &cobra.Command{
	Use:                "test [flags] [packages]",
	Short:              "Test packages (experimental)",
	DisableFlagParsing: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Setup builder to prepare workspace
		cwd, _ := os.Getwd()
		opts := &build.Options{
			RootDir:     cwd,
			KeepWorkDir: false,
		}
		builder := build.NewBuilder(opts)
		workDir, err := builder.PrepareWorkspace()
		if err != nil {
			return err
		}
		defer os.RemoveAll(workDir)
		loader := compiler.NewPackageLoader(cwd)

		// Walk the source directory to find packages and transpile them
		err = filepath.Walk(cwd, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				// Skip hidden dirs and workDir if it's inside cwd (unlikely for TempDir but possible if user sets it)
				if strings.HasPrefix(info.Name(), ".") && info.Name() != "." {
					return filepath.SkipDir
				}

				// Check for .mygo files
				hasMyGo := false
				entries, _ := os.ReadDir(path)
				for _, e := range entries {
					if strings.HasSuffix(e.Name(), ".mygo") {
						hasMyGo = true
						break
					}
				}

				if hasMyGo {
					rel, _ := filepath.Rel(cwd, path)
					importPath := filepath.ToSlash(rel)
					if importPath == "." {
						importPath = ""
					}

					pkg, err := loader.LoadPackage(importPath)
					if err != nil {
						// Log error but continue? Or fail?
						// For now, skip failed packages
						return nil
					}

					// Split files
					var normalFiles []*compiler.SourceFile
					var testFiles []*compiler.SourceFile

					for _, f := range pkg.Files {
						if strings.HasSuffix(f.Path, "_test.mygo") {
							testFiles = append(testFiles, f)
						} else {
							normalFiles = append(normalFiles, f)
						}
					}

					// Transpile normal
					hasHelpers := false
					if len(normalFiles) > 0 {
						pkg.Files = normalFiles
						code, err := compiler.TranspilePackage(pkg, loader, builder.ModName, true)
						if err == nil {
							outDir := filepath.Join(workDir, rel)
							os.MkdirAll(outDir, 0755)
							os.WriteFile(filepath.Join(outDir, pkg.Name+".go"), []byte(code), 0644)
							hasHelpers = true
						}
					}

					// Transpile test
					if len(testFiles) > 0 {
						pkg.Files = testFiles
						code, err := compiler.TranspilePackage(pkg, loader, builder.ModName, !hasHelpers)
						if err == nil {
							outDir := filepath.Join(workDir, rel)
							os.MkdirAll(outDir, 0755)
							os.WriteFile(filepath.Join(outDir, pkg.Name+"_test.go"), []byte(code), 0644)
						}
					}
				}
			}
			return nil
		})

		if err != nil {
			return err
		}

		// Run go test
		c := exec.Command("go", append([]string{"test"}, args...)...)
		c.Dir = workDir
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		return c.Run()
	},
}

func init() {
	rootCmd.AddCommand(testCmd)
}
