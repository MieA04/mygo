# MyGo 基础语法 (Basic Syntax)

MyGo 是一门现代化的静态类型编程语言，深度兼容 Go 语言生态，同时引入了大量源自 Rust、Swift 等现代语言的特性。本章将详细介绍 MyGo 的基础语法、类型系统、变量声明以及运算符。

## 1. 核心设计哲学

- **显式优于隐式**：无论是错误处理还是可见性控制，MyGo 都强调显式声明。
- **安全性**：通过空安全操作符和更严格的类型检查减少运行时错误。
- **组合优于继承**：摒弃类继承，通过 Struct 和 Trait 的组合实现多态。

---

## 2. 程序结构 (Program Structure)

每一个 MyGo 源文件都由以下部分按顺序组成：
1. **包声明 (`package`)**：定义文件所属的逻辑包。
2. **导入语句 (`import`)**：引入外部依赖。
3. **顶层声明**：包括函数、结构体、枚举、特征等。

### 语法定义
```antlr
program: packageDecl? importStmt* statement+ EOF;
```

---

## 3. 变量与常量 (Variables & Constants)

MyGo 使用 `let` 声明变量，使用 `const` 声明常量。

### 3.1 变量声明 (`let`)
变量声明支持显式类型标注和自动类型推导。

#### 语法结构
```antlr
modifier? 'let' ID (':' typeType)? ('=' expr)? ';'
```

#### 完整示例
- **基础声明**：`let x: int = 10;`
- **类型推导**：`let y = 20.5;` (自动推导为 `float`)
- **延迟赋值**：`let z: string; z = "hello";`

### 3.2 元组解构声明 (Tuple Destructuring)
MyGo 支持一次性声明并赋值多个变量。

#### 语法结构
```antlr
modifier? 'let' '(' ID (',' ID)* ')' '=' expr ';'
```

#### 示例
```mygo
let (a, b) = (1, 2);
let (status, message) = getResult();
```

### 3.3 常量声明 (`const`)
常量在声明时必须赋值，且其值在运行期间不可修改。

#### 示例
```mygo
const PI = 3.1415926;
const APP_NAME: string = "MyGoApp";
```

---

## 4. 基础数据类型 (Basic Types)

MyGo 的类型系统设计兼顾了底层性能和高层抽象。

| 类型名称 | 说明 | 示例 |
| :--- | :--- | :--- |
| **int** | 平台相关的有符号整数 | `let a: int = 100;` |
| **float** | 64位浮点数 (等同于 Go 的 float64) | `let b: float = 3.14;` |
| **bool** | 布尔值 | `let c: bool = true;` |
| **string** | UTF-8 编码的字符串 | `let d: string = "你好";` |
| **nil** | 空值字面量 | `let p: *User = nil;` |

### 指针类型 (Pointers)
MyGo 支持指针，但不支持指针运算（安全性考虑）。
- **定义**：使用 `*` 前缀，如 `*int`。
- **取地址**：使用 `&` 操作符。
- **解引用**：使用 `*` 操作符。

---

## 5. 运算符 (Operators)

MyGo 提供了一套丰富的运算符，涵盖了算术、逻辑、位运算等。

### 5.1 算术运算符
- `+`, `-`, `*`, `/`, `%`
- **自增/自减**：`++`, `--` (仅支持作为后缀，如 `i++`)

### 5.2 比较运算符
- `==`, `!=`, `>`, `<`, `>=`, `<=`
- **类型检查**：`is`, `!is` (如 `x is string`)

### 5.3 逻辑运算符
- `&&` (与), `||` (或), `!` (非)

### 5.4 特殊运算符
- **三元运算符**：`condition ? expr1 : expr2`
- **空安全操作符**：
  - `?!`：尝试解包并处理错误。
  - `?!!`：强制解包，失败则 Panic。
- **类型转换**：`to` (如 `10 to float`)

---

## 6. 编译器指令 (Compiler Directives)

MyGo 在 Trait 系统中引入了特殊的指令来控制行为：
- **ban**：显式禁止某些方法的实现或调用。
- **flip**：反转 `ban` 的逻辑，通常用于"白名单"模式。
- **repeat**：用于处理重复定义的策略。

这些指令将在 [TraitSystem-zh.md](./TraitSystem-zh.md) 中详细说明。
