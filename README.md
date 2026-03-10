# MyGo Compiler

[English](README.md) | [简体中文](README-zh.md)

MyGo is a modern statically typed programming language designed to combine the simplicity of Go with advanced type system features (such as enhanced generics, Trait system, enums, etc.). This project is the reference compiler implementation for the MyGo language.

## Table of Contents

- [MyGo Compiler](#mygo-compiler)
  - [Table of Contents](#table-of-contents)
  - [Introduction](#introduction)
    - [Significance of MyGo](#significance-of-mygo)
    - [Current Stage](#current-stage)
    - [Key Features](#key-features)
    - [Upcoming Features](#upcoming-features)
    - [Future Goals](#future-goals)
  - [Language Features](#language-features)
    - [Variable \& Function](#variable--function)
    - [Control Flow](#control-flow)
    - [Traits \& Generics](#traits--generics)
  - [Interoperability with Go](#interoperability-with-go)
    - [Importing Go Packages](#importing-go-packages)
    - [Mixed Compilation](#mixed-compilation)
  - [Build the Compiler](#build-the-compiler)
    - [Prerequisites](#prerequisites)
    - [Build Steps](#build-steps)
  - [Quick Start](#quick-start)
    - [Write Your First MyGo Program](#write-your-first-mygo-program)
    - [Transpile Mode (Recommended)](#transpile-mode-recommended)
    - [Build Mode (Experimental)](#build-mode-experimental)
  - [Usage Guide](#usage-guide)
    - [Command Line Arguments](#command-line-arguments)
      - [`transpile` - Transpile Code](#transpile---transpile-code)
      - [`build` - Build Project](#build---build-project)
  - [Project Structure](#project-structure)
  - [Documentation](#documentation)
  - [Contribution](#contribution)

## Introduction

The MyGo compiler currently adopts a **Source-to-Source** compilation strategy, transpiling MyGo code into Go code, and then using the Go compiler to generate the final executable. This allows MyGo to seamlessly leverage the existing Go ecosystem and maintain high performance.

### Significance of MyGo

To rapidly validate new Trait system syntax and explore a statically typed programming language that aligns with modern development experiences.

### Current Stage

**MVP (Minimum Viable Product) stage**.

### Key Features

- **Enhanced Generics**: Supports `where` clause constraints.
- **Trait System**: Flexible behavior composition, supporting `trait bind`.
- **Algebraic Data Types**: Supports `enum` with data (Tagged Unions) and pattern matching (`match`).
- **Modern Syntax**: Removes some Go boilerplate and introduces cleaner control flow.
- **Annotations & Metaprogramming**: Supports `@Derive`, `@macro` and compile-time code generation.

### Upcoming Features

- **Static Reflection**: RFC-008
- **OS Thread Package**: RFC-009
- **Integration with C Ecosystem**: RFC-010

### Future Goals

To rewrite the entire language framework after absorbing sufficient value from the Go ecosystem, transition to full self-bootstrapping while maintaining syntax stability, and pave the way for the future `capy` language.

## Language Features

### Variable & Function

MyGo uses `let` and `const` for variable declarations, and `fn` for functions.

```mygo
fn add(a: int, b: int): int {
    let result = a + b;
    return result;
}
```

### Control Flow

MyGo provides powerful control flow structures like `match`.

```mygo
match x {
    1 => fmt.Println("One");
    is int => fmt.Println("Is Integer");
    other => fmt.Println("Other");
}
```

### Traits & Generics

Traits define behavior, and generics support `where` clauses.

```mygo
trait Show {
    fn String(): string;
}

where T: Show
fn printShow<T>(item: T) {
    fmt.Println(item.String());
}
```

### Metaprogramming

MyGo supports macros and annotations for compile-time code generation.

```mygo
@Derive(Json)
struct User {
    name: string,
    age: int
}

@macro log_exec {
    // macro implementation...
}

@log_exec
fn do_work() {
    println("Working...");
}
```

## Interoperability with Go

### Importing Go Packages

MyGo is fully compatible with the Go ecosystem. You can import and use any Go package directly.

```mygo
import "fmt";
import "net/http";

fn main() {
    fmt.Println("Hello from Go package!");
}
```

### Mixed Compilation

Since MyGo transpiles to Go, you can mix `.mygo` and `.go` files in the same project. They will be compiled together into a single Go binary.

### Go Dependency Support

MyGo is designed to be fully compatible with the Go ecosystem:
- **Direct Import**: You can import any Go standard library or third-party package directly in your `.mygo` files (e.g., `import "encoding/json";`).
- **Go Modules**: MyGo leverages the existing `go.mod` and `go.sum` files for dependency management. Just run `go get` to add dependencies and use them in MyGo.
- **Single File Compilation**: You can transpile and run a single `.mygo` file without any configuration files or complex project structures.

## Build the Compiler

### Prerequisites

- **Go 1.20+**: Required as the compiler itself uses Go generics.
- **Make** (Optional): For running build scripts (if available).

### Build Steps

1. Clone the repository:
   ```bash
   git clone https://github.com/miea04/mygo.git
   cd mygo
   ```

2. Build the MyGo compiler:
   ```bash
   # Windows
   go build -o mygo.exe ./cmd/mygo

   # Linux / macOS
   go build -o mygo ./cmd/mygo
   ```

3. Verify installation:
   ```bash
   ./mygo.exe --help
   # Or run directly to see default demo
   ./mygo.exe
   ```

## Quick Start

### Write Your First MyGo Program

Create a file named `hello.mygo`:

```mygo
package main
import "fmt";

fn main() {
    fmt.Println("Hello, MyGo!");
}
```

### Transpile Mode (Recommended)

The most stable way currently is to transpile MyGo code to Go code first, then run it.

1. **Transpile**:
   ```bash
   ./mygo.exe transpile -o hello.go hello.mygo
   ```

2. **Run**:
   ```bash
   go run hello.go
   # Output: Hello, MyGo!
   ```

### Build Mode (Experimental)

The compiler also attempts to encapsulate the build process (currently under development, might not support single-file compilation perfectly, recommended for package compilation).

```bash
./mygo.exe build -o hello.exe .
```

## CLI Reference

The MyGo compiler (`mygo`) provides several commands for managing the development lifecycle.

### Core Commands

#### `run`
Compile and run a MyGo program immediately.
```bash
# Run a single file
mygo run main.mygo

# Run a package in the current directory
mygo run .

# Pass arguments to the program
mygo run main.mygo -- arg1 arg2
```

#### `build`
Compile a MyGo program into an executable binary.
```bash
# Build a single file
mygo build -o app.exe main.mygo

# Build a package
mygo build -o app.exe .
```

#### `transpile`
Transpile MyGo source code to Go source code without compiling to binary.
```bash
# Transpile a single file
mygo transpile -o main.go main.mygo

# Transpile a directory
mygo transpile .
```

### Development Tools

#### `fmt`
Format MyGo source code.
```bash
# Format all .mygo files in the current directory
mygo fmt

# Format specific files
mygo fmt main.mygo utils.mygo
```

#### `vet`
Run static analysis to catch potential errors.
```bash
# Check the current package
mygo vet .
```

#### `test`
Run tests (files ending in `_test.mygo`).
```bash
# Run all tests in the current directory
mygo test
```

#### `doc`
Show documentation for a package or symbol.
```bash
# Show docs for the current directory
mygo doc .
```

### Dependency Management
MyGo wraps Go's dependency management tools for convenience.

#### `mod`
Module maintenance (wrapper around `go mod`).
```bash
# Initialize a new module
mygo mod init myproject

# Tidy dependencies
mygo mod tidy
```

#### `get`
Add dependencies (wrapper around `go get`).
```bash
# Add a dependency
mygo get github.com/gin-gonic/gin
```

#### `clean`
Remove temporary build artifacts.
```bash
mygo clean
```

## Usage Guide

### Command Line Arguments

The MyGo compiler supports the following subcommands:

#### `transpile` - Transpile Code

Transpiles MyGo source code to Go source code.

```bash
mygo transpile [options] <source.mygo|directory>
```

- `-o <file>`: Specify the output Go file path.
- `-root <dir>`: Specify the root directory for package resolution (default is current directory).

#### `build` - Build Project

Compiles a MyGo project into an executable.

```bash
mygo build [options] <source.mygo|directory>
```

- `-o <file>`: Specify the output executable file path.
- `-root <dir>`: Specify the project root directory.
- `-keep-work-dir`: Keep temporary directories during compilation (for debugging).

## Project Structure

- `cmd/mygo/`: Compiler entry point (`main.go`).
- `pkg/`: Core library code.
  - `ast/`: ANTLR4 generated syntax tree nodes and parser.
  - `compiler/`: Compiler core logic (Loader, Semantic Analysis, Transpiler).
  - `build/`: Build system logic.
- `MyGo.g4`: ANTLR4 grammar definition file.

## Documentation

For detailed syntax and design documentation, please refer to the `docs/` directory:

- **Basic Syntax**: [English](docs/BasicSyntax-en.md) | [简体中文](docs/BasicSyntax-zh.md)
- **Control Flow**: [English](docs/ControlFlow-en.md) | [简体中文](docs/ControlFlow-zh.md)
- **Functions & Closures**: [English](docs/FunctionsAndClosures-en.md) | [简体中文](docs/FunctionsAndClosures-zh.md)
- **Data Structures**: [English](docs/DataStructures-en.md) | [简体中文](docs/DataStructures-zh.md)
- **OOP & Generics**: [English](docs/OOPAndGenerics-en.md) | [简体中文](docs/OOPAndGenerics-zh.md)
- **Error Handling & Concurrency**: [English](docs/ErrorHandlingAndConcurrency-en.md) | [简体中文](docs/ErrorHandlingAndConcurrency-zh.md)
- **Packages & Visibility**: [English](docs/PackageManagementAndVisibility-en.md) | [简体中文](docs/PackageManagementAndVisibility-zh.md)

## Contribution

Issues and Pull Requests are welcome!
