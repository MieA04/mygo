# MyGo 错误处理与并发 (Error Handling & Concurrency)

MyGo 继承了 Go 语言显式错误处理的哲学，但引入了现代化的语法糖来减少重复代码。同时，MyGo 原生支持并发模型。

## 1. 错误处理 (Error Handling)

MyGo 采用基于 `Result<T, E>` 的现代化错误处理模型，旨在平衡显式错误处理与语法的简洁性。

### 1.1 核心类型：`Result<T, E>`

MyGo 的错误处理围绕 `Result<T, E>` 枚举展开，它是所有可能失败操作的标准返回类型。

```mygo
enum Result<T, E> {
    Ok(T),
    Err(E)
}
```

在 RFC-015 阶段，`Option<T>` 也被纳入同一套解包操作符体系。`T?` 会在类型系统中归一化为 `Option<T>`，并可直接用于 `?!` 与 `?!!`。

### 1.2 错误传播操作符 (`?!`)

`?!` 操作符用于简化错误传播逻辑，它执行**早期返回 (Early Return)**。

- **语法**: `expr ?!`
- **语义**:
    - 计算 `expr` 的值。
    - 如果值为 `Result.Ok(v)`，则表达式求值为 `v`。
    - 如果值为 `Result.Err(e)`，则当前函数**立即返回**该错误（自动包装为 `Result.Err`）。
    - 如果值为 `Option.Some(v)`，则表达式求值为 `v`。
    - 如果值为 `Option.None`，则当前函数**立即返回** `Option.None`。

#### 示例
```mygo
fn readFileContent(path: string): Result<string, error> {
    // 如果 openFile 失败，readFileContent 直接返回错误
    let file = openFile(path) ?!;
    
    // 如果 readAll 失败，同样直接返回
    let content = readAll(file) ?!;
    
    return Result.Ok(content);
}
```

#### 底层原理
编译器将 `val = expr ?!` 解糖为：
```go
v, err := expr
if err != nil {
    return Result_Err(err)
}
val = v
```

### 1.3 Panic-Unwrap 操作符 (`?!!`)

`?!!` 用于那些你确信不会失败，或者一旦失败程序就应该崩溃的场景。

- **语法**: `expr ?!!`
- **语义**:
    - 计算 `expr` 的值。
    - 如果成功，解包并返回内部值。
    - 如果失败，触发运行时 **panic**。
    - 当 `expr` 为 `Option.None` 时，同样触发 panic。

#### 示例
```mygo
fn main() {
    // 如果出错，直接触发 panic
    let content = readFile("config.json") ?!!;
    fmt.Println("Config loaded:", content);
}
```

### 1.4 与 `match` 结合使用

当需要精细化处理错误（例如重试、本地修复或提供默认值）时，可以使用 `match` 语句。

#### 1.4.1 基础匹配与恢复
```mygo
let theme = match get_user_theme() {
    Result.Ok(t) => t,
    Result.Err(e) => "dark" // 消费掉错误，提供默认值
};
```

#### 1.4.2 多错误类型处理
```mygo
fn handleRequest(): Result<string, NetworkError> {
    let resp = sendRequest();
    
    match resp {
        Result.Ok(val) => Result.Ok(val),
        Result.Err(NetworkError.Timeout) => {
            retryRequest() // 尝试重试
        },
        other => resp // 其他错误原样透传
    }
}
```

#### 1.4.3 局部处理与传播混合
```mygo
fn complexTask(): Result<void, error> {
    match step1() {
        Result.Ok(v) => process(v) ?!,
        Result.Err(e) => {
            if is_recoverable(e) {
                recover_step1() ?!
            } else {
                return Result.Err(e);
            }
        }
    }
    return Result.Ok(nil);
}
```

### 1.5 与 Go 生态的互操作性

MyGo 能够自动处理 Go 函数返回 `(T, error)` 的情况，使其无缝适配 `?!` 操作符。

```mygo
// os.Open 是 Go 函数，返回 (*os.File, error)
let file = os.Open("test.txt") ?!;
```

编译器会自动生成 Go 风格的错误检查代码，并将 `error` 包装为 `Result.Err` 返回。

### 1.6 Defer (延迟执行)

`defer` 用于确保资源释放。MyGo 的 `defer` 支持代码块，比 Go 更灵活。

#### 底层原理：Defer 栈 (Under the Hood: Defer Stack)
*   **执行顺序**：`defer` 语句被压入一个栈中。当函数返回（或发生 panic）时，延迟块按照后进先出 (LIFO) 的顺序执行。
*   **实现**：在转译器中，这直接映射到 Go 的 `defer func() { ... }()`.
*   **性能**：Go 1.14+ 以后，`defer` 开销极小（通常会被内联），适用于性能敏感代码。

#### 语法定义
```antlr
deferStmt: 'defer' (block | exprStmt);
```

#### 示例
```mygo
fn processFile() {
    let f = openFile("data.txt") ?!!;
    
    defer {
        fmt.Println("Closing file...");
        f.Close();
    }
}
```

---

## 2. 并发 (Concurrency)

MyGo 兼容 Go 的并发模型（Goroutines 和 Channels），并提供原生语法支持。

### 2.1 Spawn (启动协程)

使用 `spawn` 关键字启动一个新的并发任务。

#### 底层原理：Goroutines (Under the Hood: Goroutines)
*   **M:N 调度器**：MyGo (通过 Go 运行时) 使用 M:N 调度器，将 M 个用户级线程 (Goroutine) 映射到 N 个内核线程上。
*   **栈管理**：Goroutine 初始栈很小（通常 2KB），并动态伸缩。这允许程序启动数百万个并发任务。
*   **抢占式调度**：调度器是协作式的，但包含了抢占点（如函数调用、循环头），防止单一任务长时间占用 CPU 导致饥饿。

#### 语法定义
```antlr
spawnStmt: 'spawn' (block | exprStmt) ;
```

#### 示例
```mygo
fn main() {
    spawn {
        fmt.Println("Running in background");
    }
    
    spawn processData();
}
```

### 2.2 Select (多路复用)

`select` 语句用于同时等待多个通信操作。

#### 语法定义
```antlr
selectStmt: 'select' '{' selectBranch (',' selectBranch)* ','? '}' ;
selectBranch
    : selectRead '=>' block    # SelectReadBranch
    | selectWrite '=>' block   # SelectWriteBranch
    | selectOther '=>' block   # SelectOtherBranch
    ;

selectRead: 'let' (ID | '(' ID (',' ID)? ')') '=' expr '.' method=ID '(' ')' ;
selectWrite: expr '.' method=ID '(' expr ')' ;
selectOther: 'other' ;
```

#### 示例
MyGo 的 `select` 语法使用**方法调用风格**来表示通道操作，这与 Go 的 `<-` 操作符有所不同。这种设计是为了保持语法的统一性。

- **接收**: `let val = ch.Recv()`
- **发送**: `ch.Send(val)`

> **注意**: 方法名（如 `Recv`, `Send`）在语法上是必需的，但目前编译器并不强制检查特定名称。建议使用 `Recv` 和 `Send` 以保持语义清晰。

```mygo
fn main() {
    let ch1 = make(chan int);
    let ch2 = make(chan string);
    
    select {
        // 从 ch1 接收数据
        let val = ch1.Recv() => {
            fmt.Println("Received from ch1:", val);
        },
        // 向 ch2 发送数据
        ch2.Send("hello") => {
            fmt.Println("Sent to ch2");
        },
        other => {
            fmt.Println("Default case");
        }
    }
}
```
