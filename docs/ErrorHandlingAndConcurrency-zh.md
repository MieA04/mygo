# MyGo 错误处理与并发 (Error Handling & Concurrency)

MyGo 继承了 Go 语言显式错误处理的哲学，但引入了现代化的语法糖来减少重复代码。同时，MyGo 原生支持并发模型。

## 1. 错误处理 (Error Handling)

MyGo 使用 `error` 接口作为标准错误类型。任何实现了 `Error() string` 方法的类型都可以作为错误。

### 1.1 Try-Unwrap 操作符 (`?!`)

MyGo 引入了 `?!` 操作符来简化 `if err != nil` 的模式。
语法：`expr ?! block`。
- 如果 `expr` 返回的最后一个值为 `nil` (无错误)，则表达式求值为除错误外的返回值。
- 如果有错误，则执行 `block`。`block` 必须中断当前控制流（如 `return`, `break`, `panic`）。

#### 语法定义
```antlr
tryUnwrapExpr: expr '?!' (block | statement)? ;
```

#### 示例
```mygo
fn main() {
    // 如果 readFile 成功，content 获取内容
    // 如果失败，执行 {} 内的代码
    let content = readFile("test.txt") ?! {
        fmt.Println("Read failed, using default");
        return; // 必须中断
    };
    
    fmt.Println(content);
}
```

### 1.2 Panic-Unwrap 操作符 (`?!!`)

`?!!` 用于那些你确信不会失败，或者一旦失败程序就应该崩溃的场景。

#### 语法定义
```antlr
panicUnwrapExpr: expr '?!!' ;
```

#### 示例
```mygo
fn main() {
    // 如果出错，直接触发 panic
    let content = readFile("config.json") ?!!;
    
    fmt.Println("Config loaded:", content);
}
```

### 1.3 Defer (延迟执行)

`defer` 用于确保资源释放。MyGo 的 `defer` 支持代码块，比 Go 更灵活。

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
```

#### 示例
```mygo
fn main() {
    let ch1 = make(chan int);
    let ch2 = make(chan string);
    
    select {
        val = <-ch1 => {
            fmt.Println("Received from ch1:", val);
        },
        ch2 <- "hello" => {
            fmt.Println("Sent to ch2");
        },
        other => {
            fmt.Println("Default case");
        }
    }
}
```
