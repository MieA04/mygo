# MyGo Basic Syntax

MyGo is a modern statically typed programming language deeply compatible with the Go ecosystem while introducing numerous features from modern languages like Rust and Swift. This chapter details MyGo's basic syntax, type system, variable declarations, and operators.

## 1. Core Design Philosophy

- **Explicit over Implicit**: Whether it's error handling or visibility control, MyGo emphasizes explicit declaration.
- **Safety**: Reducing runtime errors through null-safe operators and stricter type checking.
- **Composition over Inheritance**: Discarding class inheritance in favor of struct and trait composition for polymorphism.

---

## 2. Program Structure

Every MyGo source file consists of the following sections in order:
1. **Package Declaration (`package`)**: Defines the logical package the file belongs to.
2. **Import Statements (`import`)**: Imports external dependencies.
3. **Top-level Declarations**: Includes functions, structs, enums, traits, etc.

#### Under the Hood: Packages & Imports
*   **Logical Mapping**: MyGo maps packages directly to directories. All files in the same directory must declare the same `package` name.
*   **Compilation Units**: The compiler treats each package as a compilation unit.
*   **Visibility**: MyGo supports explicit keywords for visibility control:
    *   `pub`: Visible to all packages (Public).
    *   `pkg`: Visible only within the current package (Package Private).
    *   `pri`: Visible only within the current file (File Private).
    *   **Default Rule**: If no keyword is specified, MyGo follows Go's convention: Uppercase starts are `pub` by default, lowercase starts are `pkg` by default.

### Syntax Definition
```antlr
program: packageDecl? importStmt* (statement | annotationDecl)+ EOF;
```

---

## 3. Variables & Constants

MyGo uses `let` for variable declarations and `const` for constants.

### 3.1 Variable Declaration (`let`)
Variable declarations support explicit type annotations and automatic type inference.

#### Syntax Structure
```antlr
modifier? 'let' ID (':' typeType)? ('=' expr)? ';'
```

#### Complete Examples
- **Basic Declaration**: `let x: int = 10;`
- **Type Inference**: `let y = 20.5;` (inferred as `float`)
- **Delayed Assignment**: `let z: string; z = "hello";`

### 3.2 Tuple Destructuring
MyGo supports declaring and assigning multiple variables at once.

#### Syntax Structure
```antlr
modifier? 'let' '(' ID (',' ID)* ')' '=' expr ';'
```

#### Examples
```mygo
let (a, b) = (1, 2);
let (status, message) = getResult();
```

### 3.3 Constant Declaration (`const`)
Constants must be assigned upon declaration, and their values cannot be modified during runtime.

#### Examples
```mygo
const PI = 3.1415926;
const APP_NAME: string = "MyGoApp";
```

---

## 4. Basic Types

MyGo's type system is designed to balance low-level performance and high-level abstraction.

| Type Name | Description | Example |
| :--- | :--- | :--- |
| **int** | Platform-dependent signed integer | `let a: int = 100;` |
| **float** | 64-bit floating-point number (equivalent to Go's float64) | `let b: float = 3.14;` |
| **bool** | Boolean value | `let c: bool = true;` |
| **string** | UTF-8 encoded string | `let d: string = "Hello";` |
| **nil** | Null literal | `let p: *User = nil;` |

### Pointers
MyGo supports pointers but not pointer arithmetic (for safety considerations).
- **Definition**: Use `*` prefix, e.g., `*int`.
- **Address-of**: Use `&` operator.
- **Dereference**: Use `*` operator.

---

## 5. Operators

MyGo provides a rich set of operators covering arithmetic, logic, bitwise operations, etc.

### 5.1 Arithmetic Operators
- `+`, `-`, `*`, `/`, `%`
- **Increment/Decrement**: `++`, `--`
    - **Prefix (`++i`)**: Increments then returns the new value.
    - **Suffix (`i++`)**: Returns the old value then increments.
    - Can be used as expressions.

### 5.2 Comparison Operators
- `==`, `!=`, `>`, `<`, `>=`, `<=`
- **Type Check**: `is`, `!is` (e.g., `x is string`)

### 5.3 Logical Operators
- `&&` (AND), `||` (OR), `!` (NOT)

### 5.4 Special Operators
- **Ternary Operator**: `condition ? expr1 : expr2`
- **Error Handling Operators**:
  - `?!`: Error propagation (Early Return, like Rust's `?`).
  - `?!!`: Force unwrap, panic on failure.

```mygo
// If os.Open fails, it returns the error immediately
let f = os.Open("config.json") ?!;

// If it fails, trigger panic
let re = regexp.Compile(`\d+`) ?!!;
```
- **Type Casting**: `to` (e.g., `10 to float`)

---

## 6. Compiler Directives

MyGo introduces special directives within the Trait system to control behavior:
- **ban**: Explicitly forbids the implementation or invocation of certain methods.
- **flip**: Reverses the `ban` logic, typically used for "whitelist" mode.

These directives are detailed in [OOP & Generics](./OOPAndGenerics-en.md).
