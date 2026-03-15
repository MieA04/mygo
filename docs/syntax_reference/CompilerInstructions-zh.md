# MyGo 编译器指令文档

本文件详细介绍了如何构建 MyGo 编译器以及如何使用其提供的各项指令。

## 1. 构建 MyGo 编译器

MyGo 编译器本身使用 Go 语言开发。要从源码构建编译器，请确保您的系统中已安装 Go (建议版本 1.24.1+)。

在项目根目录下运行以下命令：

```bash
# 构建编译器并生成可执行文件
go build -o mygo.exe ./cmd/mygo/main.go
```

构建完成后，您将得到一个名为 `mygo` (Windows 下为 `mygo.exe`) 的可执行文件，这就是 MyGo 编译器的核心工具。

---

## 2. 编译器指令概览

MyGo 编译器采用了类似于 `go` 命令行工具的子命令结构。基本用法如下：

```bash
mygo <command> [arguments] [flags]
```

### 2.1 `build` - 编译代码
将 MyGo 源代码编译为可执行文件。

*   **用法**: `mygo build [flags] <source.mygo|directory>`
*   **参数**:
    *   `<source.mygo|directory>`: 指定要编译的源文件或包含项目的目录。
*   **常用选项 (Flags)**:
    *   `-o, --output <path>`: 指定输出的可执行文件路径。
    *   `--root <dir>`: 指定包解析的根目录（默认为当前目录 `.`）。
    *   `--os <os>`: 目标操作系统（如 `windows`, `linux`, `darwin`）。
    *   `--arch <arch>`: 目标指令集架构（如 `amd64`, `arm64`）。
    *   `--keep-work-dir`: 保留编译过程中生成的临时工作目录（用于调试）。

### 2.2 `run` - 编译并运行
直接编译并执行 MyGo 程序，不保留生成的可执行文件。

*   **用法**: `mygo run [flags] <source.mygo|directory> [args...]`
*   **参数**:
    *   `[args...]`: 传递给运行中程序的命令行参数。
*   **常用选项**:
    *   `--root <dir>`: 指定包解析根目录。
    *   `--keep-work-dir`: 保留临时编译文件。

### 2.3 `transpile` - 转译代码
将 MyGo 代码转译为等效的 Go 源代码。

*   **用法**: `mygo transpile [flags] <source.mygo|directory>`
*   **常用选项**:
    *   `-o, --output <path>`: 指定生成的 Go 代码输出路径。
    *   `--root <dir>`: 包解析根目录。

### 2.4 `fmt` - 格式化代码
自动调整 MyGo 源代码的格式，使其符合标准规范。

*   **用法**: `mygo fmt [files]`
*   **说明**: 如果不指定文件，将递归格式化当前目录下的所有 `.mygo` 文件。

### 2.5 `test` - 运行测试
执行 MyGo 项目中的测试用例（实验性功能）。

*   **用法**: `mygo test [flags] [packages]`
*   **说明**: 该命令会识别以 `_test.mygo` 结尾的文件并运行其中的测试。

### 2.6 `mod` - 模块管理
MyGo 模块管理的包装指令，底层调用 `go mod`。

*   **用法**: `mygo mod <arguments>`
*   **常见用法**:
    *   `mygo mod init <module_name>`: 初始化新模块。
    *   `mygo mod tidy`: 整理依赖关系。

### 2.7 `get` - 依赖获取
获取并安装第三方依赖，底层调用 `go get`。

*   **用法**: `mygo get [packages]`

### 2.8 `vet` - 静态分析
对 MyGo 代码进行静态语义检查，识别潜在的编程错误。

*   **用法**: `mygo vet [packages]`

### 2.9 `doc` - 查看文档
提取并显示 MyGo 包或符号的文档注释。

*   **用法**: `mygo doc [path]`

### 2.10 `clean` - 清理缓存
删除编译过程中产生的临时文件和构件。

*   **用法**: `mygo clean`

### 2.11 `version` - 版本信息
显示当前 MyGo 编译器的版本。

*   **用法**: `mygo version`

---

## 3. 开发建议

在开发 MyGo 项目时，建议将编译出的 `mygo` 可执行文件所在路径添加到系统的 `PATH` 环境变量中，以便在任何目录下都能直接使用 `mygo` 命令。
