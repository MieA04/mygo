# MyGo Functions & Closures

Functions are the fundamental units of execution in MyGo. MyGo supports named functions, anonymous functions (closures), and higher-order functions, with deep integration into the generics system.

## 1. Function Declaration

Use the `fn` keyword to define functions. MyGo's function definitions are highly flexible, supporting multiple return values and generic constraints.

### Syntax Definition
```antlr
fnDecl: whereClause? modifier? 'fn' ID typeParams? '(' paramList? ')' (':' typeType)? block ;
```

### 1.1 Basic Example
```mygo
// No parameters and no return value
fn sayHello() {
    fmt.Println("Hello!");
}

// With parameters and a single return value
fn add(a: int, b: int): int {
    return a + b;
}
```

### 1.2 Multiple Return Values
MyGo supports multiple return values, implemented via tuples under the hood.

```mygo
fn divMod(a: int, b: int): (int, int) {
    return (a / b, a % b);
}

fn main() {
    let (div, mod) = divMod(10, 3);
    fmt.Println(div, mod);
}
```

---

## 2. Anonymous Functions & Closures (Lambdas)

MyGo supports a lightweight Lambda expression syntax, which can capture variables from the surrounding scope.

### Syntax Definition
```antlr
lambdaExpr: '(' paramList? ')' (':' typeType)? '=>' block ;
```

### 2.1 Basic Usage
```mygo
fn main() {
    // Define an anonymous function and assign it to a variable
    let multiply = (a: int, b: int): int => {
        return a * b;
    };
    
    fmt.Println(multiply(3, 4)); // Output 12
}
```

### 2.2 Closures
Closures can reference variables defined in their external scope.

```mygo
fn main() {
    let factor = 10;
    
    // Capture external variable factor
    let scaler = (x: int): int => {
        return x * factor;
    };
    
    fmt.Println(scaler(5)); // Output 50
}
```

### 2.3 Passing as Arguments (Higher-Order Functions)
Functions are first-class citizens in MyGo and can be passed as arguments to other functions.

```mygo
// Define a function that accepts a function as an argument
// The type of parameter op is fn(int): int
fn apply(val: int, op: fn(int): int): int {
    return op(val);
}

fn main() {
    let result = apply(10, (x: int): int => {
        return x * x;
    });
    fmt.Println(result); // 100
}
```

---

## 3. Generic Functions

Functions can define type parameters for generic programming. MyGo's generics support powerful type constraints.

### 3.1 Basic Generics
```mygo
fn printIfEqual<T>(a: T, b: T) {
    if a == b {
        fmt.Println("Equal");
    }
}
```

### 3.2 Generic Constraints (Where Clauses)
You can use `where` clauses to impose constraints on generic parameters, requiring them to implement specific traits.
**Note: In MyGo, the `where` clause must be written at the very beginning of the function declaration.**

```mygo
// Requires T to implement the Addable Trait
where T: Addable
fn addGeneric<T>(a: T, b: T): T {
    return a + b;
}

// Multiple constraints
where T: Show + Debug
fn printDetails<T>(item: T) {
    fmt.Println(item.String());
}
```

---

## 4. Function Types

Function types describe the signature of a function, typically used for variable declarations or parameter type annotations.

### Syntax Structure
```antlr
typeType: 'fn' '(' typeList? ')' (':' typeType)? ;
```

### Example
```mygo
// Declare a variable of type "function accepting two ints and returning an int"
let operation: fn(int, int): int;

operation = add;
```
