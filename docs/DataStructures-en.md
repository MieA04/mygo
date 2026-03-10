# MyGo Data Structures

MyGo provides core data structures such as structs, enums, arrays, and slices, supporting generics and object-oriented programming paradigms.

## 1. Structs

Structs are the core of custom data types in MyGo, used to combine multiple fields of different types.

### Syntax Definition
```antlr
structDecl: whereClause? modifier? 'struct' ID typeParams? '{' (structField (',' structField)* ','?)? '}' ;
structField: ID ':' typeType ;
```

### 1.1 Defining Structs
Struct field names usually start with an uppercase letter to indicate public visibility (with `pub` keyword), or lowercase for private visibility.

```mygo
// Define a simple Point struct
struct Point {
    x: int,
    y: int
}

// Struct with modifiers (pub makes the struct itself visible outside the package)
pub struct User {
    pub ID: int,        // Field visible outside the package
    pub Name: string,
    pub Email: string,
    age: int,           // Field visible only within the package
    pri secret: string  // Field visible only within the current file
}

// Generic struct
struct Box<T> {
    Value: T
}
```

### 1.2 Instantiation and Usage
MyGo supports struct initialization syntax similar to JSON or Rust.

```mygo
fn main() {
    // Instantiate
    let p = Point{x: 10, y: 20};
    
    // Access fields
    fmt.Println(p.x);
    
    // Modify fields (if the variable is mutable)
    p.x = 100;
    
    // Generic instantiation
    let intBox = Box<int>{Value: 123};
    let strBox = Box<string>{Value: "MyGo"};
}
```

---

## 2. Enums

MyGo's Enums are powerful Algebraic Data Types (ADTs), which can define not only sets of constants but also carry associated data.

### Syntax Definition
```antlr
enumDecl: whereClause? modifier? 'enum' ID typeParams? '{' enumVariant (',' enumVariant)* ','? '}' ;
enumVariant: ID ('(' typeList ')')? ;
```

### 2.1 Simple Enums
Similar to C/Go enums, used to represent a set of named constants.

```mygo
enum Color {
    Red,
    Green,
    Blue
}
```

### 2.2 Enums with Data (Tagged Unions)
Enum members can carry data of different types, which is very useful for state machines or message passing.

```mygo
enum Message {
    Quit,                       // No data
    Move(int, int),             // Carries two ints (x, y)
    Write(string),              // Carries a string
    ChangeColor(int, int, int)  // Carries three ints (r, g, b)
}
```

### 2.3 Handling Enums with Match
The `match` statement is the best way to handle enums, ensuring all cases are covered (exhaustiveness check).

```mygo
fn process(msg: Message) {
    match msg {
        Message.Quit => fmt.Println("Quit"),
        Message.Move(x, y) => fmt.Printf("Move to %d, %d\n", x, y),
        Message.Write(s) => fmt.Println("Message:", s),
        // Use other to handle unlisted cases, or list all cases
        other => fmt.Println("Other message")
    }
}
```

---

## 3. Arrays & Slices

MyGo distinguishes between fixed-length arrays and dynamic-length slices.

### 3.1 Arrays
Arrays have a fixed length and are value types.

```mygo
// Type representation: [length]Type
let arr: [5]int = [1, 2, 3, 4, 5];

// Automatic inference
let names = ["Alice", "Bob"]; // Inferred as [2]string
```

### 3.2 Slices
Slices are references to underlying arrays and have variable length. This is the most commonly used sequence type in MyGo.

```mygo
// Type representation: []Type (no length)
let numbers: []int = [1, 2, 3]; 

// Create slice from array
let arr = [10, 20, 30, 40];
let slice = arr[1..3]; // Includes index 1, 2 (20, 30)
```

### 3.3 Common Operations
MyGo's slice operations are highly consistent with Go.

```mygo
let list = [1, 2, 3];

// Get length
let l = len(list);

// Access element
let first = list[0];

// Modify element
list[1] = 100;

// Iterate
for (v : list) {
    fmt.Println(v);
}
```

---

## 4. Maps

MyGo currently supports Maps via generic libraries or interoperability; the standard library typically provides a `Map<K, V>` type.

```mygo
// Example: Assuming the standard library provides a Map type
let scores = Map<string, int>{};
scores.Put("Alice", 100);
scores.Put("Bob", 95);

let aliceScore = scores.Get("Alice");
```

---

## 5. Tuples

Tuples can combine multiple values of different types into a single compound value.

```mygo
let pair: (string, int) = ("Alice", 30);
let (name, age) = pair;
```
