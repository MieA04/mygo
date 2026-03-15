# MyGo 包管理与可见性 (Packages & Visibility)

MyGo 采用模块化设计，使用包 (Package) 来组织代码。它与 Go 语言的包系统高度兼容，但提供了更精细、显式的可见性控制。

##### 1. 包声明 (Package Declaration)

每个源文件都必须隶属于一个包。包声明位于文件的第一行（注释除外）。

#### 底层原理：模块系统 (Under the Hood: Module System)
*   **Go Modules 集成**：MyGo 完全复用 Go Modules。项目根目录下的 `go.mod` 定义了模块路径。
*   **目录即包**：与 Go 一样，同一目录下的所有文件必须声明为同一个包名。
*   **编译**：编译器基于 `GOPATH` 或 `go.mod` 解析导入路径，确保与现有 Go 库的兼容性。

### 语法定义
```antlr
packageDecl: 'package' ID ';'? ;
```

### 示例
```mygo
package main

// 或者定义库包
package math_utils
```

- `package main`：定义可执行程序的入口包，必须包含 `main` 函数。
- 其他包名：通常与目录名保持一致，用于库代码。

---

## 2. 导入 (Import)

使用 `import` 关键字导入依赖包。MyGo 支持单行导入和块导入。

### 语法定义
```antlr
importStmt
    : 'import' '{' importSpec (',' importSpec)* ','? '}' ';'? # BlockImport
    | 'import' importSpec ';'?                                # SingleImport
    ;
importSpec: STRING ('as' ID)? ;
```

### 示例
```mygo
// 单行导入
import "fmt";

// 块导入
import {
    "net/http",
    "os" as std_os // 别名导入
}
```

---

## 3. 可见性修饰符 (Visibility Modifiers)

MyGo 摒弃了 Go 语言"首字母大小写决定可见性"的隐式规则，转而采用显式的关键字修饰符。这使得代码意图更加清晰，且不受命名风格限制。

### 语法定义
```antlr
modifier: 'pub' | 'pkg' | 'pri' ;
```

### 3.1 级别说明

| 关键字 | 级别 | 说明 | 适用范围 |
| :--- | :--- | :--- | :--- |
| **pub** | Public | 公开可见 | 任何导入该包的代码均可访问 |
| **pkg** | Package | 包内可见 | 同一个包下的所有文件可访问 (默认级别) |
| **pri** | Private | 文件私有 | 仅当前源文件内可见 |

#### 底层原理：符号重整 (Under the Hood: Symbol Mangling)
为了在 Go 简单的"首字母大写即公开"模型之上实现这三层可见性，MyGo 在转译时使用名称重整技术：
*   `pub fn add` -> `func Add` (导出)
*   `pkg fn helper` -> `func helper` (包内私有)
*   `pri fn internal` -> `func internal_Hash123` (文件私有，增加文件哈希后缀以防冲突)

### 3.2 示例详解

假设我们有一个数学库包 `math_lib`。

```mygo
package math_lib

// pub: 对外公开的结构体
pub struct Vector {
    // pub: 对外公开的字段
    pub X: float,
    pub Y: float,
    
    // pkg: 仅包内可见的字段 (默认，不写修饰符即为 pkg)
    cachedLength: float,
    
    // pri: 仅当前文件可见的私有字段
    pri secretKey: string
}

// pub: 对外公开的函数
pub fn NewVector(x: float, y: float): Vector {
    return Vector{X: x, Y: y, cachedLength: 0.0, secretKey: "hidden"};
}
```

### 3.3 编译映射 (Transpilation Mapping)

当 MyGo 代码被转译为 Go 代码时，编译器会自动处理命名转换以符合 Go 的可见性规则：

- `pub` 符号 -> 转译为 **大写首字母** (例如 `NewVector`).
- `pkg` / `pri` 符号 -> 转译为 **小写首字母** (例如 `newVector` 或 `newVector_suffix`).

这种设计让你可以在 MyGo 中自由使用小写字母开头的公开函数（如 `pub fn add()`），而编译器会负责将其转换为 Go 能够导出的 `Add()`。

---

## 4. 混合编译 (Mixed Compilation)

MyGo 旨在与现有的 Go 生态系统无缝集成。

1.  **直接导入 Go 包**：你可以在 MyGo 中直接 `import "encoding/json"`，编译器会自动识别标准库。
2.  **项目共存**：MyGo 项目可以包含 `.go` 文件。MyGo 编译器在构建时会处理这些文件，允许 MyGo 代码调用 Go 代码，反之亦然（需遵循可见性规则）。
3.  **Go Modules**：MyGo 复用 `go.mod` 和 `go.sum` 进行依赖管理，无需额外的包管理工具。

```bash
# 初始化项目
go mod init myproject

# 编译 MyGo 项目
mygo build
```
