# MyGo 语言语法大全 (Syntax Reference)

**版本**: 1.0
**状态**: 根据 `MyGo.g4` 生成

本文档详细列出了 MyGo 编程语言的所有关键字、语法结构和规则。它是理解和编写 MyGo 代码的权威参考。

---

## 1. 词法结构 (Lexical Structure)

### 1.1 注释 (Comments)
MyGo 支持两种风格的注释：
*   **单行注释**: 以 `//` 开头，直到行尾。
    ```mygo
    // 这是一个单行注释
    let x = 1; //这也是注释
    ```
*   **多行注释**: 以 `/*` 开始，以 `*/` 结束，支持跨行。
    ```mygo
    /* 这是一个
       多行注释 */
    ```

### 1.2 标识符 (Identifiers)
标识符用于命名变量、函数、类型等。
*   **规则**: `[a-zA-Z_][a-zA-Z_0-9]*`
*   **示例**: `myVar`, `MyStruct`, `_unused`, `calculate_sum`

### 1.3 关键字 (Keywords)
以下单词是保留的关键字，不能用作标识符（部分基础类型目前暂未作为硬性关键字，但在规范中视为预定义类型）：

| 关键字 | 用途 |
| :--- | :--- |
| `pub`, `pri` | 访问修饰符 (Public, Private) |
| `let`, `const` | 变量与常量声明 |
| `fn` | 函数声明 |
| `return` | 函数返回 |
| `struct` | 结构体定义 |
| `enum` | 枚举定义 |
| `trait` | 特质/接口定义 |
| `bind` | Trait 绑定扩展 |
| `where` | 泛型约束子句 |
| `if`, `else` | 条件分支 |
| `match`, `other` | 模式匹配 |
| `while`, `loop`, `for` | 循环控制 |
| `break`, `continue` | 循环跳转 |
| `is`, `!is` | 类型检查 |
| `to` | 类型转换 |
| `macro` | 宏定义 |
| `this` | 当前实例引用 |
| `nil` | 空指针/空值 |
| `ban`, `flip`, `repeat` | Trait 冲突解决指令 |

### 1.4 字面量 (Literals)
*   **整数 (`INT`)**: 连续的数字序列。例: `123`, `0`, `9999`。
*   **浮点数 (`FLOAT`)**: 包含小数点的数字。例: `3.14`, `0.01`。
*   **字符串 (`STRING`)**: 双引号包裹的字符序列。例: `"Hello World"`, `""`。

---

## Import 语句

MyGo 支持单行导入和块状导入，并支持包别名。

```mygo
// 单行导入
import "fmt";

// 块状导入与别名
import {
    "net/http",
    "encoding/json" as json,
    "myproject/utils" as u
}
```

> **注意**：包路径必须使用双引号包裹。别名使用 `as` 关键字指定。

---

## 2. 类型系统 (Types)

MyGo 是静态强类型语言。类型 (`typeType`) 可以在声明变量或函数参数时显式指定。

### 2.1 基础类型 (Primitive Types)
*虽在语法中解析为标识符，但在语义上预定义：*
*   `int`: 整数
*   `float`: 浮点数
*   `bool`: 布尔值
*   `string`: 字符串
*   `byte`: 字节

### 2.2 复合类型 (Composite Types)
*   **指针**: `*T`。例: `*int`。
*   **数组/切片**: `T[N]` 或 `T[]`。例: `int[5]`, `string[]`。
*   **元组**: `(T1, T2, ...)`。例: `(int, string)`。
*   **泛型实例化**: `Type<Args>`。例: `List<int>`, `Map<string, int>`。
*   **函数类型**: `fn(Args): Ret`。例: `fn(int, int): bool`。

---

## 3. 声明 (Declarations)

### 3.1 变量声明 (`let` / `const`)
*   **基本声明**:
    ```mygo
    let x: int = 10;
    let y = 20; // 类型推导
    const PI: float = 3.14;
    ```
*   **元组解构**:
    ```mygo
    let (a, b) = some_tuple;
    ```
*   **可见性**: `pub let global_var = 1;`

### 3.2 函数声明 (`fn`)
```mygo
fn add(a: int, b: int): int {
    return a + b;
}

// 泛型函数与约束
where T: Comparable
fn max<T>(a: T, b: T): T  {
    // ...
}
```

### 3.3 结构体 (`struct`)
```mygo
struct Point {
    x: int,
    y: int
}

// 泛型结构体
struct Box<T> {
    value: T
}
```

### 3.4 枚举 (`enum`)
```mygo
enum Color {
    Red,
    Green,
    Blue
}

// 带值枚举 (Tagged Union)
enum Result<T, E> {
    Ok(T),
    Err(E)
}
```

### 3.5 Trait 定义 (`trait`)
```mygo
// 纯 Trait
trait Printable {
    fn print();
}

// 绑定 Trait (扩展方法)
trait bind(p: Point) {
    fn distance(): float {
        // ...
    }
}
```

### 3.6 注解 (`@Annotation`)

注解用于为声明添加元数据。

```mygo
// 标记结构体自动生成 JSON 序列化方法
@Derive(Json)
struct User {
    id: int,
    name: string
}

// 标记初始化函数
@Init
fn init_system() {
    // ...
}
```

### 3.7 宏定义 (`@macro`)

宏允许在编译期操作 AST 并生成代码。

```mygo
@macro log_exec {
    let name = target.name;
    let body = target.body;
    return #quote {
        println("Enter: " + name);
        body;
        println("Exit: " + name);
    };
}

@log_exec
fn my_func() {
    // ...
}
```

---

## 4. 语句 (Statements)

### 4.1 赋值
```mygo
x = x + 1;
```

### 4.2 条件 (`if`)
```mygo
if x > 0 {
    // ...
} else if x == 0 {
    // ...
} else {
    // ...
}
```

### 4.3 模式匹配 (`match`)
```mygo
match value {
    1 => print("One"),
    2, 3 => print("Two or Three"), // 多值匹配
    is string => print("It's a string"), // 类型匹配
    other => print("Default") // 默认分支
}
```

### 4.4 循环
*   **While 循环**:
    ```mygo
    while x < 10 { x = x + 1; }
    ```
*   **无限循环**:
    ```mygo
    loop { if cond { break; } }
    ```
*   **For 循环**:
    ```mygo
    // 1. 范围
    for (i : 0..10) { ... }
    
    // 2. 迭代器
    for (item : list) { ... }
    for (key, val : map) { ... }
    
    // 3. C 风格
    for (let i = 0; i < 10; i++) { ... }
    ```

---

## 5. 表达式 (Expressions)

### 5.1 运算符
*   **算术**: `+`, `-`, `*`, `/`
*   **比较**: `==`, `!=`, `>`, `<`, `>=`, `<=`
*   **逻辑**: `&&`, `||`, `!` (非)
*   **指针**: `&` (取地址), `*` (解引用)
*   **自增/减**: `++`, `--` (后缀)

### 5.2 构造与转换
*   **结构体实例化**:
    ```mygo
    let p = Point{ x: 1, y: 2 };
    ```
*   **数组字面量**: `[1, 2, 3]`
*   **类型转换**: `expr to Type` (例: `1.5 to int`)
*   **类型检查**: `expr is Type`, `expr !is Type`

### 5.3 特殊表达式
*   **Lambda**: `(x) => { return x + 1; }`
*   **Try Unwrap (`?!`)**: 尝试解包 Result/Option，失败则返回错误。
*   **Panic Unwrap (`?!!`)**: 强制解包，失败则 Panic。
*   **内置函数 (Built-in Functions)**:
    *   `print(...)`: 变长参数。在转译阶段自动映射为 Go 的 `fmt.Println`。
    *   `_mygo_must(v, err)`: 内部解包助手，处理带错误返回的调用。
    *   `_mygo_ternary(cond, a, b)`: 三元表达式助手。

---

## 6. 编译器工具链 (Toolchain)

### 6.1 命令行子命令
MyGo 编译器支持以下操作模式：
*   **`build`**: 编译单文件到可执行文件。
    ```bash
    mygo build -o output.exe source.mygo
    ```
*   **`transpile`**: 仅转译为 Go 源代码。
    ```bash
    mygo transpile -o output.go source.mygo
    ```

---

## 7. 语法详细定义 (EBNF 摘要)

以下是核心语法的 EBNF 摘要（简化版）：

```ebnf
program      ::= statement+ EOF
statement    ::= varDecl | fnDecl | structDecl | enumDecl | traitDecl 
               | ifStmt | matchStmt | loopStmt | whileStmt | forStmt 
               | returnStmt | breakStmt | continueStmt | assignmentStmt | exprStmt

varDecl      ::= modifier? ('let' | 'const') ID (':' type)? ('=' expr)? ';'
fnDecl       ::= modifier? 'fn' ID typeParams? '(' paramList? ')' (':' type)? block

structDecl   ::= 'struct' ID typeParams? '{' fieldList? '}'
enumDecl     ::= 'enum' ID typeParams? '{' variantList? '}'

ifStmt       ::= 'if' expr block ('else' 'if' expr block)* ('else' block)?
matchStmt    ::= 'match' expr '{' matchCase+ '}'

expr         ::= literal
               | ID
               | expr op expr
               | expr '(' args? ')'
               | expr '.' ID
               | structLiteral
               | lambda
               | ...
```
