package core

import (
	"github.com/miea04/mygo/pkg/ast"
	"github.com/miea04/mygo/pkg/compiler/symbols"
)

// Package represents a MyGo package (a directory of .mygo files)
type Package struct {
	Name        string
	ImportPath  string // e.g. "myproject/utils" or "./utils"
	DirPath     string
	Files       []*SourceFile
	Scope       *symbols.Scope // Public symbols exported by this package
	IsGoPackage bool
}

type SourceFile struct {
	Path    string
	Code    string
	AST     ast.IProgramContext
	Imports []string // Import paths extracted from AST
}
