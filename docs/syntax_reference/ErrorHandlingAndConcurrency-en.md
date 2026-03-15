# MyGo Error Handling & Concurrency

## 1. Error Handling

MyGo adopts a modern error handling model based on `Result<T, E>`, designed to balance explicit error handling with syntactic conciseness.

### 1.1 Core Type: `Result<T, E>`

Error handling in MyGo revolves around the `Result<T, E>` enum, which is the standard return type for all operations that can fail.

```mygo
enum Result<T, E> {
    Ok(T),
    Err(E)
}
```

In RFC-015 phase scope, `Option<T>` is also handled by the same unwrap operators. `T?` is normalized to `Option<T>` in the type system, and can be used directly with `?!` and `?!!`.

### 1.2 Error Propagation Operator (`?!`)

The `?!` operator is used to simplify error propagation logic by performing an **Early Return**.

- **Syntax**: `expr ?!`
- **Semantics**:
    - Evaluates `expr`.
    - If the value is `Result.Ok(v)`, the expression evaluates to `v`.
    - If the value is `Result.Err(e)`, the current function **returns immediately** with that error (automatically wrapped in `Result.Err`).
    - If the value is `Option.Some(v)`, the expression evaluates to `v`.
    - If the value is `Option.None`, the current function **returns immediately** with `Option.None`.

#### Example
```mygo
fn readFileContent(path: string): Result<string, error> {
    // If openFile fails, readFileContent returns the error immediately
    let file = openFile(path) ?!;
    
    // If readAll fails, it also returns immediately
    let content = readAll(file) ?!;
    
    return Result.Ok(content);
}
```

#### Under the Hood
The compiler desugars `val = expr ?!` into:
```go
v, err := expr
if err != nil {
    return Result_Err(err)
}
val = v
```

### 1.3 Panic-Unwrap Operator (`?!!`)

`?!!` is used for scenarios where you are certain it won't fail, or if it fails, the program should crash.

- **Syntax**: `expr ?!!`
- **Semantics**:
    - Evaluates `expr`.
    - If successful, unwraps and returns the inner value.
    - If it fails, triggers a runtime **panic**.
    - For `Option.None`, it also triggers a panic.

#### Example
```mygo
fn main() {
    // If an error occurs, panic immediately
    let content = readFile("config.json") ?!!;
    fmt.Println("Config loaded:", content);
}
```

### 1.4 Integration with `match`

When fine-grained error handling is needed (e.g., retrying, local fixing, or providing default values), the `match` statement can be used.

#### 1.4.1 Basic Matching and Recovery
```mygo
let theme = match get_user_theme() {
    Result.Ok(t) => t,
    Result.Err(e) => "dark" // Consumes the error, provides a default value
};
```

#### 1.4.2 Multiple Error Types
```mygo
fn handleRequest(): Result<string, NetworkError> {
    let resp = sendRequest();
    
    match resp {
        Result.Ok(val) => Result.Ok(val),
        Result.Err(NetworkError.Timeout) => {
            retryRequest() // Attempt retry
        },
        other => resp // Pass through other errors as-is
    }
}
```

#### 1.4.3 Mixing Local Handling and Propagation
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

### 1.5 Go Ecosystem Interoperability

MyGo automatically handles Go functions that return `(T, error)`, making them seamlessly compatible with the `?!` operator.

```mygo
// os.Open is a Go function returning (*os.File, error)
let file = os.Open("test.txt") ?!;
```

The compiler automatically generates Go-style error checking code and wraps the `error` into `Result.Err` for return.

### 1.6 Defer

`defer` is used to ensure resource release. MyGo's `defer` supports code blocks, making it more flexible than Go.

#### Under the Hood: Defer Stack
*   **Execution Order**: `defer` statements are pushed onto a stack. When a function returns (or panics), deferred blocks are executed in LIFO (Last-In-First-Out) order.
*   **Implementation**: In the transpiler, this maps directly to Go's `defer func() { ... }()`.
*   **Performance**: Since Go 1.14+, `defer` overhead is minimal (often inlined), making it suitable for performance-critical code.

#### Syntax Definition
```antlr
deferStmt: 'defer' (block | exprStmt);
```

#### Example
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

## 2. Concurrency

MyGo is compatible with Go's concurrency model (Goroutines and Channels) and provides native syntax support.

### 2.1 Spawn (Start Coroutine)

Use the `spawn` keyword to start a new concurrent task.

#### Under the Hood: Goroutines
*   **M:N Scheduler**: MyGo (via Go runtime) uses an M:N scheduler, mapping M user-level threads (Goroutines) onto N kernel threads.
*   **Stack**: Goroutines start with a small stack (typically 2KB) which grows and shrinks dynamically. This allows spawning millions of concurrent tasks.
*   **Preemption**: The scheduler is cooperative but includes preemption points (e.g., function calls, loop headers) to prevent starvation.

#### Syntax Definition
```antlr
spawnStmt: 'spawn' (block | exprStmt) ;
```

#### Example
```mygo
fn main() {
    spawn {
        fmt.Println("Running in background");
    }
    
    spawn processData();
}
```

### 2.2 Select (Multiplexing)

The `select` statement is used to wait for multiple communication operations simultaneously.

#### Syntax Definition
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

#### Example
MyGo's `select` syntax uses a **method call style** for channel operations, differing from Go's `<-` operator. This design maintains syntactic uniformity.

- **Receive**: `let val = ch.Recv()`
- **Send**: `ch.Send(val)`

> **Note**: While method names (like `Recv`, `Send`) are syntactically required, the compiler currently doesn't enforce specific names. It is recommended to use `Recv` and `Send` for clarity.

```mygo
fn main() {
    let ch1 = make(chan int);
    let ch2 = make(chan string);
    
    select {
        // Receive from ch1
        let val = ch1.Recv() => {
            fmt.Println("Received from ch1:", val);
        },
        // Send to ch2
        ch2.Send("hello") => {
            fmt.Println("Sent to ch2");
        },
        other => {
            fmt.Println("Default case");
        }
    }
}
```
