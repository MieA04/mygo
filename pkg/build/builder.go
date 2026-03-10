package build

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/miea04/mygo/pkg/compiler"
)

// Options represents build configuration
type Options struct {
	SourcePath  string
	OutputPath  string
	KeepWorkDir bool
	WorkDir     string
	RootDir     string
	GOOS        string
	GOARCH      string
}

// Builder orchestrates the build process
type Builder struct {
	Options *Options
	Loader  *compiler.PackageLoader
	ModName string
}

func NewBuilder(opts *Options) *Builder {
	return &Builder{
		Options: opts,
		Loader:  compiler.NewPackageLoader(opts.RootDir),
	}
}

// Build executes the build process
func (b *Builder) Build() error {
	// 1. Prepare Workspace
	workDir, err := b.PrepareWorkspace()
	if err != nil {
		return err
	}
	if !b.Options.KeepWorkDir {
		defer os.RemoveAll(workDir)
	} else {
		fmt.Printf("Build workspace kept at: %s\n", workDir)
	}

	// 2. Load Package
	mainPkg, err := b.loadPackage()
	if err != nil {
		return err
	}

	// 3. Transpile
	if err := b.transpile(workDir); err != nil {
		return err
	}

	// 4. Go Build
	if err := b.goBuild(workDir, mainPkg.Name, b.Options.OutputPath); err != nil {
		return err
	}

	return nil
}

// Run executes the build process and runs the binary
func (b *Builder) Run(args []string) error {
	// 1. Prepare Workspace
	workDir, err := b.PrepareWorkspace()
	if err != nil {
		return err
	}
	if !b.Options.KeepWorkDir {
		defer os.RemoveAll(workDir)
	}

	// 2. Load Package
	mainPkg, err := b.loadPackage()
	if err != nil {
		return err
	}

	// 3. Transpile
	if err := b.transpile(workDir); err != nil {
		return err
	}

	// 4. Go Build (to temp executable)
	tempExe := filepath.Join(workDir, "runner")
	if runtime.GOOS == "windows" {
		tempExe += ".exe"
	}

	if err := b.goBuild(workDir, mainPkg.Name, tempExe); err != nil {
		return err
	}

	// 5. Execute
	cmd := exec.Command(tempExe, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func (b *Builder) PrepareWorkspace() (string, error) {
	// Create temp dir
	workDir, err := os.MkdirTemp("", "mygo-build-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}

	// Check for go.mod in RootDir
	goModPath := filepath.Join(b.Options.RootDir, "go.mod")
	if _, err := os.Stat(goModPath); err == nil {
		// Copy go.mod and go.sum
		if err := copyFile(goModPath, filepath.Join(workDir, "go.mod")); err != nil {
			return "", err
		}
		goSumPath := filepath.Join(b.Options.RootDir, "go.sum")
		if _, err := os.Stat(goSumPath); err == nil {
			if err := copyFile(goSumPath, filepath.Join(workDir, "go.sum")); err != nil {
				return "", err
			}
		}

		// Read module name
		modName, err := readModuleName(goModPath)
		if err != nil {
			return "", err
		}
		b.ModName = modName
	} else {
		// Init new go.mod
		b.ModName = "mygo_project"
		if err := initGoMod(workDir, b.ModName); err != nil {
			return "", err
		}
	}

	// Copy all .go files from source to workDir, maintaining structure
	// We need to walk RootDir and copy relevant files
	err = filepath.Walk(b.Options.RootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip .git, node_modules, temp dirs, etc.
		if info.IsDir() {
			if strings.HasPrefix(info.Name(), ".") && info.Name() != "." {
				return filepath.SkipDir
			}
			if info.Name() == "node_modules" || info.Name() == "vendor" { // Maybe keep vendor?
				return filepath.SkipDir
			}
			return nil
		}

		// We care about .go, .c, .h, .s files primarily
		ext := filepath.Ext(path)
		if ext == ".go" || ext == ".c" || ext == ".h" || ext == ".s" {
			relPath, err := filepath.Rel(b.Options.RootDir, path)
			if err != nil {
				return err
			}
			destPath := filepath.Join(workDir, relPath)
			if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
				return err
			}
			if err := copyFile(path, destPath); err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return "", fmt.Errorf("failed to copy project files: %w", err)
	}

	return workDir, nil
}

func (b *Builder) loadPackage() (*compiler.Package, error) {
	// Logic similar to main.go runBuild
	stat, err := os.Stat(b.Options.SourcePath)
	if err != nil {
		return nil, err
	}

	var pkg *compiler.Package
	// Adjust import path relative to RootDir
	relPath, err := filepath.Rel(b.Options.RootDir, b.Options.SourcePath)
	if err != nil {
		return nil, err
	}
	// On Windows, Rel returns backslashes, convert to slashes for import path
	importPath := filepath.ToSlash(relPath)
	if importPath == "." {
		importPath = "" // Root package? Or use "."?
	}

	if stat.IsDir() {
		pkg, err = b.Loader.LoadPackage(importPath)
	} else {
		// TODO: Handle single file build if needed, for now assuming dir build or adapting logic
		// If it's a file, we treat parent dir as package but filter for that file?
		// Or create ad-hoc package like in main.go
		// For RFC-005, let's focus on directory/package build as primary
		return nil, fmt.Errorf("single file build not fully supported in new builder yet")
	}

	if err != nil {
		return nil, err
	}
	return pkg, nil
}

func (b *Builder) transpile(workDir string) error {
	// Transpile all loaded MyGo packages
	for _, pkg := range b.Loader.LoadedPackages {
		// Skip if it's a Go package (placeholder logic for now)
		if pkg.IsGoPackage {
			continue
		}

		goCode, err := compiler.TranspilePackage(pkg, b.Loader, b.ModName, true)
		if err != nil {
			return err
		}

		// Determine output path in workDir

		// Safer approach: Use pkg.DirPath relative to Loader.RootPath
		relDir, err := filepath.Rel(b.Options.RootDir, pkg.DirPath)
		if err != nil {
			return err
		}

		outDir := filepath.Join(workDir, relDir)
		if err := os.MkdirAll(outDir, 0755); err != nil {
			return err
		}

		// Write file
		// Main package -> main.go? Or package_name.go?
		// To avoid conflict with existing .go files, maybe use .mygo.go?
		// But existing .go files are already copied.
		// If main package, main.go is standard.
		outFile := filepath.Join(outDir, pkg.Name+".go")
		if pkg.Name == "main" {
			outFile = filepath.Join(outDir, "main.go")
		}

		if err := os.WriteFile(outFile, []byte(goCode), 0644); err != nil {
			return err
		}
	}
	return nil
}

func (b *Builder) goBuild(workDir, mainPkgName, outputPath string) error {
	exePath := outputPath
	if exePath == "" {
		// Default name based on source dir
		base := filepath.Base(b.Options.SourcePath)
		if base == "." || base == string(filepath.Separator) {
			base = filepath.Base(b.Options.RootDir)
		}
		exePath = base
		if runtime.GOOS == "windows" {
			exePath += ".exe"
		}
	}

	absOutput, err := filepath.Abs(exePath)
	if err != nil {
		return err
	}

	// Calculate build target package
	// If SourcePath is RootDir, then "."
	// Else relative path
	relPath, err := filepath.Rel(b.Options.RootDir, b.Options.SourcePath)
	if err != nil {
		return err
	}
	buildTarget := "./" + filepath.ToSlash(relPath)
	if relPath == "." {
		buildTarget = "."
	}

	cmd := exec.Command("go", "build", "-o", absOutput, buildTarget)
	cmd.Dir = workDir
	// Inherit environment variables
	cmd.Env = os.Environ()
	if b.Options.GOOS != "" {
		cmd.Env = append(cmd.Env, "GOOS="+b.Options.GOOS)
	}
	if b.Options.GOARCH != "" {
		cmd.Env = append(cmd.Env, "GOARCH="+b.Options.GOARCH)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("go build failed: %w\n%s", err, string(output))
	}

	// Only print success if OutputPath was specified in Options (meaning user requested build)
	// If it's a temp run, maybe skip?
	// But here we don't know if it's run or build easily without checking b.Options.OutputPath == outputPath
	if b.Options.OutputPath == outputPath && outputPath != "" {
		fmt.Printf("Build successful: %s\n", absOutput)
	}
	return nil
}

// Helpers

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func initGoMod(dir, moduleName string) error {
	cmd := exec.Command("go", "mod", "init", moduleName)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("go mod init failed: %w\n%s", err, string(out))
	}
	return nil
}

func readModuleName(goModPath string) (string, error) {
	content, err := os.ReadFile(goModPath)
	if err != nil {
		return "", err
	}
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module ")), nil
		}
	}
	return "", fmt.Errorf("module name not found in go.mod")
}
