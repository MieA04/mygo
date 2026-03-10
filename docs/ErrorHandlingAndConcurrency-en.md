# MyGo Error Handling & Concurrency

MyGo inherits Go's explicit error handling philosophy but introduces modern syntax sugar to reduce boilerplate code. At the same time, MyGo natively supports the concurrency model.

## 1. Error Handling

MyGo uses the `error` interface as the standard error type. Any type implementing the `Error() string` method can serve as an error.

### 1.1 Try-Unwrap Operator (`?!`)

MyGo introduces the `?!` operator to simplify the `if err != nil` pattern.
Syntax: `expr ?! block`.
- If the last return value of `expr` is `nil` (no error), the expression evaluates to the return values excluding the error.
- If there is an error, `block` is executed. The `block` must interrupt the current control flow (e.g., `return`, `break`, `panic`).

#### Syntax Definition
```antlr
tryUnwrapExpr: expr '?!' (block | statement)? ;
```

#### Example
```mygo
fn main() {
    // If readFile succeeds, content gets the value
    // If it fails, the code inside {} is executed
    let content = readFile("test.txt") ?! {
        fmt.Println("Read failed, using default");
        return; // Must interrupt
    };
    
    fmt.Println(content);
}
```

### 1.2 Panic-Unwrap Operator (`?!!`)

`?!!` is used for scenarios where you are certain it won't fail, or if it fails, the program should crash.

#### Syntax Definition
```antlr
panicUnwrapExpr: expr '?!!' ;
```

#### Example
```mygo
fn main() {
    // If an error occurs, panic immediately
    let content = readFile("config.json") ?!!;
    
    fmt.Println("Config loaded:", content);
}
```

### 1.3 Defer

`defer` is used to ensure resource release. MyGo's `defer` supports code blocks, making it more flexible than Go.

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
```

#### Example
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
