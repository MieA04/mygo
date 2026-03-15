# MyGo Control Flow

MyGo provides rich and modern control flow statements, including conditional branching, powerful `match` pattern matching, and various loop structures.

## 1. Conditional Statements (If-Else)

The `if` statement executes a code block based on a condition. MyGo's `if` syntax is concise; the condition expression does not need parentheses, but the execution body must use curly braces.

### Syntax Definition
```antlr
ifStmt: 'if' expr block ('else' 'if' expr block)* ('else' block)? ;
```

### Example
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

### Nesting
`if` statements can be nested arbitrarily, but for readability, it is recommended to use `match` for complex branching logic.

---

## 2. Pattern Matching (Match)

`match` is the most powerful control flow structure in MyGo, replacing the traditional `switch` statement and providing greater expressiveness. `match` supports value matching, type matching, and destructuring matching.

### Syntax Definition
```antlr
matchStmt: 'match' expr '{' matchCase+ '}' ;
matchCase
    : expr (',' expr)* '=>' (block | statement)    // Value Match / Destructuring
    | 'is' typeType '=>' (block | statement)       // Type Match
    | 'other' '=>' (block | statement)             // Default Case
    ;
```

> **Note**: If the right-hand side is a single-line statement, it must end with a semicolon `;`. If it is a block `{ ... }`, no semicolon is needed.

### 2.1 Value Matching
Matches literals, variables, or constants. Supports multiple items separated by commas.

```mygo
let status = 200;

match status {
    200, 201 => {
        fmt.Println("Success");
    }
    400 => fmt.Println("Bad Request"); // Single statement requires semicolon
    404 => fmt.Println("Not Found");
    500 => fmt.Println("Server Error");
    other => fmt.Println("Unknown status: ", status);
}
```

### 2.2 Variable Binding
If the match pattern is an unbound identifier (and not `true`/`false`), it binds the matched value to that variable, making it available within the branch. This is often used instead of `other` to capture the value.

```mygo
let x = 42;

match x {
    0 => fmt.Println("Zero");
    val => fmt.Println("Value is:", val); // val is bound to 42
}
```

### 2.3 Type Matching
`match` can branch based on runtime type (similar to Go's Type Switch).

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

### 2.4 Enum Destructuring
When matching Enums (ADTs), internal associated values can be destructured and extracted.

```mygo
enum Result {
    Ok(int),
    Err(string)
}

fn handleResult(res: Result) {
    match res {
        // Match Result.Ok and bind inner value to val
        Result.Ok(val) => fmt.Printf("Got value: %d\n", val);
        
        // Match Result.Err and bind inner value to msg
        Result.Err(msg) => fmt.Printf("Error occurred: %s\n", msg);
    }
}
```

#### Nested Generic Enum Matching
MyGo also supports matching on generic enums.

```mygo
// Assuming Option<T> is defined as enum Option<T> { Some(T), None }
fn checkOption(opt: Option<int>) {
    match opt {
        Option.Some(v) => fmt.Println("Got integer:", v);
        Option.None => fmt.Println("Got nothing");
    }
}
```

### 2.5 Under the Hood

#### Compilation Strategy
The MyGo compiler transpiles `match` statements into Go's `switch` or `if-else` structures, depending on the match type:

1.  **Value Matching**: Transpiled to Go's `switch expr { case val1, val2: ... }`.
2.  **Type Matching**: Transpiled to Go's `switch v := expr.(type) { case int: ... }`.
3.  **Enum Destructuring**:
    *   First, performs a type assertion via `switch v := expr.(type)`.
    *   Maps MyGo enum variants (e.g., `Result.Ok`) to underlying Go struct types (e.g., `Result_Ok`).
    *   Inside the `case` branch, generates temporary variable binding code, e.g., `val := v.Item1`, to achieve destructuring.

#### Limitations
*   Currently, destructuring is only supported for a single level; nested destructuring (e.g., `Result.Ok(Option.Some(v))`) is not supported.
*   You cannot mix destructuring patterns with regular value patterns in the same `case`.

---

## 3. Loops

MyGo provides three loop structures: `for`, `while`, and `loop`, to meet different needs.

### 3.1 While Loop
The most basic conditional loop, executes while the condition is true.

#### Syntax Definition
```antlr
whileStmt: 'while' expr block ;
```

#### Example
```mygo
let i = 0;
while i < 5 {
    fmt.Println(i);
    i++;
}
```

### 3.2 Loop Loop
An infinite loop, must be used with `break`. This is more semantic and better optimized by the compiler than `while true`.

#### Syntax Definition
```antlr
loopStmt: 'loop' block ;
```

#### Example
```mygo
let count = 0;
loop {
    if count >= 10 {
        break; // Break loop
    }
    fmt.Println("Running...");
    count++;
}
```

### 3.3 For Loop
The `for` loop is the most powerful, supporting three forms.

#### Syntax Definition
```antlr
forStmt
    : 'for' '(' ID ':' expr '..' expr ')' block                      # RangeForStmt
    | 'for' '(' forInit? ';' cond=expr? ';' step=expr? ')' block     # TraditionalForStmt
    | 'for' '(' ID (',' ID)? ':' expr ')' block                      # IteratorForStmt
    ;
```

#### Form 1: Traditional Loop (C-style)
Suitable for scenarios requiring fine-grained index control.
```mygo
for (let i = 0; i < 10; i++) {
    fmt.Println(i);
}
```

#### Form 2: Range Loop
Suitable for iterating over numeric ranges. Syntax `start..end` denotes a left-closed, right-open interval `[start, end)`.
```mygo
// Iterate from 0 to 9 (inclusive 0, exclusive 10)
for (i : 0..10) {
    fmt.Printf("Number: %d\n", i);
}
```

#### Form 3: Iterator Loop
Iterates over arrays, slices, or any collection implementing the iterator protocol.
```mygo
let arr = [10, 20, 30];

// Iterate over values only
for (val : arr) {
    fmt.Println(val);
}

// Iterate over index and value (similar to Go's range)
for (idx, val : arr) {
    fmt.Printf("Index: %d, Value: %d\n", idx, val);
}
```

#### Under the Hood: Loop Optimization
*   **Range Loops**: Expressions like `0..10` are optimized at compile-time. No array or iterator object is created; it compiles to a simple integer counter loop.
*   **Iterator Loops**: For arrays and slices, the compiler generates standard index-based access (`len` check + direct access), avoiding the overhead of interface-based iterators.
*   **Infinite Loops (`loop`)**: Since the condition is always true, the compiler can omit the conditional jump instruction at the loop header, slightly improving performance over `while true`.

---

## 4. Jump Statements

MyGo supports standard jump control statements.

- **break**: Immediately terminates the current loop.
- **continue**: Skips the current iteration and proceeds to the next loop cycle.
- **return**: Returns a result from a function.

```mygo
fn findEven(numbers: int[]) {
    for (n : numbers) {
        if n % 2 != 0 {
            continue; // Skip odd numbers
        }
        
        fmt.Println("Found even:", n);
        
        if n > 100 {
            fmt.Println("Found large even, stopping.");
            break; // Stop after finding a large even number
        }
    }
}
```
