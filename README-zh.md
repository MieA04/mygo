# MyGo 语言编译器

[English](README.md) | [简体中文](README-zh.md)

MyGo 是一种现代化的静态类型编程语言，旨在结合 Go 语言的简洁性与更高级的类型系统特性（如增强的泛型、Trait 系统、枚举等）。本项目是 MyGo 语言的参考编译器实现。

## 目录

- [MyGo 语言编译器](#mygo-语言编译器)
  - [目录](#目录)
  - [简介](#简介)
    - [MyGo 的意义](#mygo-的意义)
    - [当前阶段](#当前阶段)
    - [主要特性](#主要特性)
    - [待加入特性](#待加入特性)
    - [未来目标](#未来目标)
  - [语言特性](#语言特性)
    - [变量与函数](#变量与函数)
    - [流程控制](#流程控制)
    - [Trait 与泛型](#trait-与泛型)
    - [元编程 (Metaprogramming)](#元编程-metaprogramming)
  - [与 Go 的互操作性](#与-go-的互操作性)
    - [引用 Go 包](#引用-go-包)
    - [混合编译](#混合编译)
    - [Go 依赖支持](#go-依赖支持)
  - [构建编译器](#构建编译器)
    - [前提条件](#前提条件)
    - [编译步骤](#编译步骤)
  - [快速开始](#快速开始)
    - [编写第一个 MyGo 程序](#编写第一个-mygo-程序)
    - [转译模式 (推荐)](#转译模式-推荐)
    - [构建模式 (实验性)](#构建模式-实验性)
  - [命令行参考](#命令行参考)
    - [核心命令](#核心命令)
      - [`run`](#run)
      - [`build`](#build)
      - [`transpile`](#transpile)
    - [开发工具](#开发工具)
      - [`fmt`](#fmt)
      - [`vet`](#vet)
      - [`test`](#test)
      - [`doc`](#doc)
    - [依赖管理](#依赖管理)
      - [`mod`](#mod)
      - [`get`](#get)
      - [`clean`](#clean)
  - [使用指南](#使用指南)
    - [命令行参数](#命令行参数)
      - [`transpile` - 转译代码](#transpile---转译代码)
      - [`build` - 构建项目](#build---构建项目)
  - [项目结构](#项目结构)
  - [文档](#文档)
  - [贡献](#贡献)

## 简介

MyGo 编译器目前采用 **Source-to-Source (源码到源码)** 的编译策略，将 MyGo 代码转译为 Go 代码，然后利用 Go 编译器生成最终的可执行文件。这使得 MyGo 能够无缝利用现有的 Go 生态系统，并保持高性能。

### MyGo 的意义

快速验证新的 Trait 系统语法，并尝试探索更符合现代化开发体验的静态类型编程语言。

### 当前阶段

**MVP (Minimum Viable Product) 阶段**。

### 主要特性

- **增强的泛型**: 支持 `where` 子句约束。
- **Trait 系统**: 灵活的行为组合，支持 `trait bind`。
- **现代化错误处理**: 引入 `?!` (错误传播) 和 `?!!` (Panic-Unwrap) 操作符，基于 `Result<T, E>` 类型实现显式且高效的错误管理。
- **可选类型**: 支持 `T?` 语法并归一化为 `Option<T>`，支持 Option 场景的 `?!` / `?!!`。
- **代数数据类型**: 支持带数据的 `enum` (Tagged Unions) 和模式匹配 (`match`)。
- **现代化语法**: 去除了部分 Go 的样板代码，引入更简洁的控制流。
- **注解与元编程**: 支持 `@Derive`, `@macro` 等高级特性，实现编译期代码生成。

### 待加入特性

- **静态反射 (Static Reflection)**: RFC-008
- **OS 线程包**: RFC-009
- **与 C 生态的融合**: RFC-010

### 未来目标

从 Go 生态中汲取足够的营养后重写整个语言框架，在保持语法不变的前提下过渡到完全自举，并为未来的 `capy` 语言做出探索和尝试。

## 语言特性

### 变量与函数

MyGo 使用 `let` 和 `const` 声明变量，使用 `fn` 定义函数。

```mygo
fn add(a: int, b: int): int {
    let result = a + b;
    return result;
}
```

### 错误处理

MyGo 采用类似 Rust 的现代化错误处理机制，使用 `?!` 操作符进行错误传播。

```mygo
fn readFile(path: string): Result<string, error> {
    // 如果 openFile 失败，立即返回该错误
    let f = os.Open(path) ?!;
    let content = io.ReadAll(f) ?!;
    return Result.Ok(content);
}
```

### 流程控制

MyGo 提供了强大的流程控制结构，如 `match` 模式匹配。

```mygo
match x {
    1 => fmt.Println("One");
    is int => fmt.Println("Is Integer");
    other => fmt.Println("Other");
}
```

### Trait 与泛型

Trait 定义行为，泛型支持 `where` 约束子句。

```mygo
trait Show {
    fn String(): string;
}

where T: Show
fn printShow<T>(item: T) {
    fmt.Println(item.String());
}
```

### 元编程 (Metaprogramming)

MyGo 支持宏和注解，允许在编译期生成代码。

```mygo
@Derive(Json)
struct User {
    name: string,
    age: int
}

@macro log_exec {
    // 宏实现...
}

@log_exec
fn do_work() {
    println("Working...");
}
```

## 与 Go 的互操作性

### 引用 Go 包

MyGo 与 Go 生态完全兼容。你可以直接导入并使用任何 Go 标准库或第三方包。

```mygo
import "fmt";
import "net/http";

fn main() {
    fmt.Println("Hello from Go package!");
}
```

### 混合编译

由于 MyGo 会被转译为 Go 代码，你可以在同一个项目中混合使用 `.mygo` 和 `.go` 文件。它们将被一起编译成最终的 Go 二进制文件。

### Go 依赖支持

MyGo 设计为与 Go 生态系统完全兼容：
- **直接导入**: 你可以在 `.mygo` 文件中直接导入任何 Go 标准库或第三方包（例如 `import "encoding/json";`）。
- **Go Modules**: MyGo 利用现有的 `go.mod` 和 `go.sum` 文件进行依赖管理。只需运行 `go get` 添加依赖，然后在 MyGo 中直接使用。
- **单文件编译**: 你可以在没有任何配置文件或复杂项目结构的情况下转译并运行单个 `.mygo` 文件。

## 构建编译器

### 前提条件

- **Go 1.20+**: 由于编译器本身使用了 Go 泛型，需要较新的 Go 版本。
- **Make** (可选): 用于运行构建脚本（如果有）。

### 编译步骤

1. 克隆仓库：
   ```bash
   git clone https://github.com/miea04/mygo.git
   cd mygo
   ```

2. 编译 MyGo 编译器：
   ```bash
   # Windows
   go build -o mygo.exe ./cmd/mygo

   # Linux / macOS
   go build -o mygo ./cmd/mygo
   ```

3. 验证安装：
   ```bash
   ./mygo.exe --help
   # 或者直接运行查看默认演示
   ./mygo.exe
   ```

4. 交叉编译（发布常用）：
   ```bash
   # Windows
   go build -o mygo.exe ./cmd/mygo

   # Linux amd64
   GOOS=linux GOARCH=amd64 go build -o mygo-linux-amd64 ./cmd/mygo
   ```

## 快速开始

### 编写第一个 MyGo 程序

创建一个名为 `hello.mygo` 的文件：

```mygo
package main
import "fmt";

fn main() {
    fmt.Println("Hello, MyGo!");
}
```

### 转译模式 (推荐)

目前最稳定的使用方式是先将 MyGo 代码转译为 Go 代码，然后运行。

1. **转译**:
   ```bash
   ./mygo.exe transpile -o hello.go hello.mygo
   ```

2. **运行**:
   ```bash
   go run hello.go
   # 输出: Hello, MyGo!
   ```

### 构建模式 (实验性)

编译器也尝试直接封装构建过程（目前仍在完善中，可能不支持单文件编译，建议用于包编译）。

```bash
./mygo.exe build -o hello.exe .
```

## 命令行参考

MyGo 编译器 (`mygo`) 提供了一系列用于管理开发生命周期的命令。

### 核心命令

#### `run`
立即编译并运行 MyGo 程序。
```bash
# 运行单个文件
mygo run main.mygo

# 运行当前目录下的包
mygo run .

# 传递参数给程序
mygo run main.mygo -- arg1 arg2
```

#### `build`
将 MyGo 程序编译为可执行二进制文件。
```bash
# 编译单个文件
mygo build -o app.exe main.mygo

# 编译包
mygo build -o app.exe .
```

#### `transpile`
将 MyGo 源代码转译为 Go 源代码，而不编译为二进制文件。
```bash
# 转译单个文件
mygo transpile -o main.go main.mygo

# 转译目录
mygo transpile .
```

### 开发工具

#### `fmt`
格式化 MyGo 源代码。
```bash
# 格式化当前目录下的所有 .mygo 文件
mygo fmt

# 格式化指定文件
mygo fmt main.mygo utils.mygo
```

#### `vet`
运行静态分析以捕获潜在错误。
```bash
# 检查当前包
mygo vet .
```

#### `test`
运行测试（以 `_test.mygo` 结尾的文件）。
```bash
# 运行当前目录下的所有测试
mygo test
```

#### `doc`
显示包或符号的文档。
```bash
# 显示当前目录的文档
mygo doc .
```

### 依赖管理
MyGo 封装了 Go 的依赖管理工具以方便使用。

#### `mod`
模块维护（`go mod` 的包装器）。
```bash
# 初始化新模块
mygo mod init myproject

# 整理依赖
mygo mod tidy
```

#### `get`
添加依赖（`go get` 的包装器）。
```bash
# 添加依赖
mygo get github.com/gin-gonic/gin
```

#### `clean`
清理临时构建产物。
```bash
mygo clean
```

## 使用指南

### 命令行参数

MyGo 编译器支持以下子命令：

#### `transpile` - 转译代码

将 MyGo 源代码转译为 Go 源代码。

```bash
mygo transpile [options] <source.mygo|directory>
```

- `-o <file>`: 指定输出的 Go 文件路径。
- `-root <dir>`: 指定包解析的根目录（默认为当前目录）。

#### `build` - 构建项目

编译 MyGo 项目为可执行文件。

```bash
mygo build [options] <source.mygo|directory>
```

- `-o <file>`: 指定输出的可执行文件路径。
- `-root <dir>`: 指定项目根目录。
- `-keep-work-dir`: 保留编译过程中的临时目录（用于调试）。

## 项目结构

- `cmd/mygo/`: 编译器入口点 (`main.go`)。
- `pkg/`: 核心库代码。
  - `ast/`: ANTLR4 生成的语法树节点和解析器。
  - `compiler/`: 编译器核心逻辑 (Loader, Semantic Analysis, Transpiler)。
  - `build/`: 构建系统逻辑。
- `MyGo.g4`: ANTLR4 语法定义文件。

## 文档

详细的语法和设计文档请参考 `docs/` 目录：

- **基础语法**: [English](docs/BasicSyntax-en.md) | [简体中文](docs/BasicSyntax-zh.md)
- **流程控制**: [English](docs/ControlFlow-en.md) | [简体中文](docs/ControlFlow-zh.md)
- **函数与闭包**: [English](docs/FunctionsAndClosures-en.md) | [简体中文](docs/FunctionsAndClosures-zh.md)
- **数据结构**: [English](docs/DataStructures-en.md) | [简体中文](docs/DataStructures-zh.md)
- **面向对象与泛型**: [English](docs/OOPAndGenerics-en.md) | [简体中文](docs/OOPAndGenerics-zh.md)
- **错误处理与并发**: [English](docs/ErrorHandlingAndConcurrency-en.md) | [简体中文](docs/ErrorHandlingAndConcurrency-zh.md)
- **包管理与可见性**: [English](docs/PackageManagementAndVisibility-en.md) | [简体中文](docs/PackageManagementAndVisibility-zh.md)

## 贡献

欢迎提交 Issue 和 Pull Request！
