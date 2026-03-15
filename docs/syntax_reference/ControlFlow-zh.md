# MyGo 流程控制 (Control Flow)

MyGo 提供了丰富且现代的流程控制语句，包括条件分支、强大的 `match` 模式匹配以及多种循环结构。

## 1. 条件语句 (If-Else)

`if` 语句用于基于条件执行代码块。MyGo 的 `if` 语法简洁，条件表达式不需要小括号包裹，但执行体必须使用大括号。

### 语法定义
```antlr
ifStmt: 'if' expr block ('else' 'if' expr block)* ('else' block)? ;
```

### 示例
```mygo
let score = 85;

if score >= 90 {
    fmt.Println("Excellent");
} else if score >= 60 {
    fmt.Println("Pass");
} else {
    fmt.Println("Fail");
}
```

### 嵌套使用
`if` 语句可以任意嵌套，但为了代码可读性，建议优先使用 `match` 处理复杂的分支逻辑。

---

## 2. 模式匹配 (Match)

`match` 是 MyGo 中最强大的控制流结构，它替代了传统的 `switch` 语句，并提供了更强的表达能力。`match` 支持值匹配、类型匹配以及解构匹配。

### 语法定义
```antlr
matchStmt: 'match' expr '{' matchCase+ '}' ;
matchCase
    : expr (',' expr)* '=>' (block | statement)    // 值匹配/解构匹配
    | 'is' typeType '=>' (block | statement)       // 类型匹配
    | 'other' '=>' (block | statement)             // 默认分支
    ;
```

> **注意**：如果分支右侧是单行语句，必须以分号 `;` 结尾。如果是代码块 `{ ... }`，则不需要分号。

### 2.1 值匹配 (Value Matching)
可以匹配字面量、变量或常量。支持使用逗号分隔多个匹配项。

```mygo
let status = 200;

match status {
    200, 201 => {
        fmt.Println("Success");
    }
    400 => fmt.Println("Bad Request"); // 单行语句需要分号
    404 => fmt.Println("Not Found");
    500 => fmt.Println("Server Error");
    other => fmt.Println("Unknown status: ", status);
}
```

### 2.2 变量绑定 (Variable Binding)
如果匹配模式是一个未绑定的标识符（且不是 `true`/`false`），它会将匹配的值绑定到该变量，并在分支内可用。这通常用于替代 `other` 分支以获取值。

```mygo
let x = 42;

match x {
    0 => fmt.Println("Zero");
    val => fmt.Println("Value is:", val); // val 被绑定为 42
}
```

### 2.3 类型匹配 (Type Matching)
`match` 可以根据运行时类型进行分支处理（类似于 Go 的 Type Switch）。

```mygo
fn printType(v: any) {
    match v {
        is int => fmt.Println("It's an integer");
        is string => fmt.Println("It's a string");
        is bool => fmt.Println("It's a boolean");
        is Point => fmt.Println("It's a Point struct");
        other => fmt.Println("Unknown type");
    }
}
```

### 2.4 枚举解构匹配 (Enum Destructuring)
当匹配枚举（ADT）时，可以解构提取内部关联的值。

```mygo
enum Result {
    Ok(int),
    Err(string)
}

fn handleResult(res: Result) {
    match res {
        // 匹配 Result.Ok 并将内部值绑定到 val
        Result.Ok(val) => fmt.Printf("Got value: %d\n", val);
        
        // 匹配 Result.Err 并将内部值绑定到 msg
        Result.Err(msg) => fmt.Printf("Error occurred: %s\n", msg);
    }
}
```

#### 嵌套泛型枚举匹配
MyGo 也支持泛型枚举的匹配。

```mygo
// 假设 Option<T> 定义为 enum Option<T> { Some(T), None }
fn checkOption(opt: Option<int>) {
    match opt {
        Option.Some(v) => fmt.Println("Got integer:", v);
        Option.None => fmt.Println("Got nothing");
    }
}
```

### 2.5 底层原理 (Under the Hood)

#### 编译策略
MyGo 编译器会将 `match` 语句转译为 Go 的 `switch` 或 `if-else` 结构，具体取决于匹配类型：

1.  **普通值匹配**：转译为 Go 的 `switch expr { case val1, val2: ... }`。
2.  **类型匹配**：转译为 Go 的 `switch v := expr.(type) { case int: ... }`。
3.  **枚举解构**：
    *   首先通过 `switch v := expr.(type)` 进行类型断言。
    *   将 MyGo 的枚举变体（如 `Result.Ok`）映射到底层 Go 结构体类型（如 `Result_Ok`）。
    *   在 `case` 分支内部，生成临时变量绑定代码，例如 `val := v.Item1`，从而实现解构。

#### 限制
*   目前解构匹配仅支持单层解构，不支持嵌套解构（如 `Result.Ok(Option.Some(v))`）。
*   在同一个 `case` 中不能混合使用解构模式和普通值模式。

---

## 3. 循环 (Loops)

MyGo 提供了 `for`, `while`, `loop` 三种循环结构，满足不同场景的需求。

### 3.1 While 循环
最基础的条件循环，当条件为真时执行。

#### 语法定义
```antlr
whileStmt: 'while' expr block ;
```

#### 示例
```mygo
let i = 0;
while i < 5 {
    fmt.Println(i);
    i++;
}
```

### 3.2 Loop 循环
无限循环，必须配合 `break` 使用。这比 `while true` 更语义化且编译器优化更好。

#### 语法定义
```antlr
loopStmt: 'loop' block ;
```

#### 示例
```mygo
let count = 0;
loop {
    if count >= 10 {
        break; // 跳出循环
    }
    fmt.Println("Running...");
    count++;
}
```

### 3.3 For 循环
`for` 循环最为强大，支持三种形式。

#### 语法定义
```antlr
forStmt
    : 'for' '(' ID ':' expr '..' expr ')' block                      # RangeForStmt
    | 'for' '(' forInit? ';' cond=expr? ';' step=expr? ')' block     # TraditionalForStmt
    | 'for' '(' ID (',' ID)? ':' expr ')' block                      # IteratorForStmt
    ;
```

#### 形式 1: C 风格循环 (Traditional)
适用于需要精细控制索引的场景。
```mygo
for (let i = 0; i < 10; i++) {
    fmt.Println(i);
}
```

#### 形式 2: 范围循环 (Range)
适用于数值范围遍历。语法 `start..end` 表示左闭右开区间 `[start, end)`。
```mygo
// 遍历 0 到 9 (包含 0，不包含 10)
for (i : 0..10) {
    fmt.Printf("Number: %d\n", i);
}
```

#### 形式 3: 迭代器循环 (Iterator)
遍历数组、切片、Map 或任何实现了迭代器协议的集合。

##### 遍历数组/切片
```mygo
let arr = [10, 20, 30];

// 仅遍历值
for (val : arr) {
    fmt.Println(val);
}

// 遍历索引和值 (类似于 Go 的 range)
for (idx, val : arr) {
    fmt.Printf("Index: %d, Value: %d\n", idx, val);
}
```

##### 遍历 Map
Map 遍历时会同时返回键 (Key) 和值 (Value)。**注意：Map 的遍历顺序是不确定的。**

```mygo
let scores = Map<string, int>();
scores["Alice"] = 100;
scores["Bob"] = 90;

for (name, score : scores) {
    fmt.Printf("Name: %s, Score: %d\n", name, score);
}
```

#### 底层原理：循环优化 (Under the Hood: Loop Optimization)
*   **范围循环 (Range)**：类似 `0..10` 的表达式会在编译期被优化。编译器不会创建数组或迭代器对象，而是生成简单的整数计数循环。
*   **迭代器循环 (Iterator)**：对于数组和切片，编译器会生成基于索引的直接访问代码（`len` 检查 + 直接访问），避免接口调用的开销。
*   **无限循环 (`loop`)**：由于条件永远为真，编译器可以省略循环头部的条件跳转指令，比 `while true` 稍微高效一些。

---

## 4. 跳转语句 (Jump Statements)

MyGo 支持标准的跳转控制语句。

- **break**: 立即终止当前循环。
- **continue**: 跳过本次迭代，进入下一次循环。
- **return**: 从函数返回结果。

```mygo
fn findEven(numbers: int[]) {
    for (n : numbers) {
        if n % 2 != 0 {
            continue; // 跳过奇数
        }
        
        fmt.Println("Found even:", n);
        
        if n > 100 {
            fmt.Println("Found large even, stopping.");
            break; // 找到大偶数后停止
        }
    }
}
```
