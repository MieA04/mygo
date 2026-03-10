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
  - [构建编译器](#构建编译器)
    - [前提条件](#前提条件)
    - [编译步骤](#编译步骤)
  - [快速开始](#快速开始)
    - [编写第一个 MyGo 程序](#编写第一个-mygo-程序)
    - [转译模式 (推荐)](#转译模式-推荐)
    - [构建模式 (实验性)](#构建模式-实验性)
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
- **代数数据类型**: 支持带数据的 `enum` (Tagged Unions) 和模式匹配 (`match`)。
- **现代化语法**: 去除了部分 Go 的样板代码，引入更简洁的控制流。

### 待加入特性

- **注解 (Annotations)**: RFC-007
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

fn printShow<T>(item: T) where T: Show {
    fmt.Println(item.String());
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

## 快速开始

### 编写第一个 MyGo 程序

创建一个名为 `hello.mygo` 的文件：

```mygo
package main

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

- [语法参考](docs/语法/)
  - [01 基础语法](docs/语法/01_基础语法.md)
  - [02 流程控制](docs/语法/02_流程控制.md)
  - [03 函数与闭包](docs/语法/03_函数与闭包.md)
  - [04 数据结构](docs/语法/04_数据结构.md)
  - [05 面向对象与泛型](docs/语法/05_面向对象与泛型.md)
  - [06 错误处理与并发](docs/语法/06_错误处理与并发.md)
  - [07 包管理与可见性](docs/语法/07_包管理与可见性.md)

## 贡献

欢迎提交 Issue 和 Pull Request！
