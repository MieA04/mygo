# MyGo Data Structures

MyGo provides core data structures such as structs, enums, arrays, and slices, supporting generics and object-oriented programming paradigms.

## 1. Structs

Structs are the core of custom data types in MyGo, used to combine multiple fields of different types.

### Syntax Definition
```antlr
structDecl: annotationUsage* whereClause? modifier? 'struct' ID typeParams? '{' (structField (',' structField)* ','?)? '}' ;
structField: ID ':' typeType (STRING)? ;
```

### 1.1 Defining Structs
Struct field visibility is determined solely by **capitalization**: Uppercase starts are Public, lowercase starts are Private. The struct itself can use modifiers (like `pub`) to control its visibility.

```mygo
// Define a simple Point struct
struct Point {
    x: int,
    y: int
}

// Struct with modifiers (pub makes the struct itself visible outside the package)
pub struct User {
    ID: int,            // Uppercase: Field visible outside the package
    Name: string,
    Email: string,
    age: int,           // Lowercase: Field visible only within the package
    secret: string      // Lowercase: Field visible only within the package
}

// Struct with Tags (Commonly used for JSON serialization)
struct Config {
    Host: string "json:\"host\"",
    Port: int    "json:\"port\""
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

#### Under the Hood: Memory Layout
*   **Contiguous Memory**: Structs in MyGo are laid out contiguously in memory. Fields are stored in the order they are defined.
*   **Padding**: The compiler may insert padding bytes between fields to satisfy alignment requirements of the target architecture (e.g., aligning `int64` to 8-byte boundaries).
*   **Value Semantics**: Structs are value types. Assigning a struct to another variable copies the entire data structure.

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

#### Under the Hood: Tagged Unions
*   **Representation**: Enums with data are compiled as a struct with a hidden "Tag" field (discriminator) and a data field.
*   **Union Storage**: In memory, the data fields of different variants share the same memory space (like a C union), sized to the largest variant.
*   **Type Safety**: The compiler ensures you only access the data valid for the current tag, typically via `match`.

---

## 3. Arrays & Slices

MyGo distinguishes between fixed-length arrays and dynamic-length slices.

### 3.1 Arrays
Arrays have a fixed length and are value types.

```mygo
// Type representation: Type[length]
let arr: int[5] = [1, 2, 3, 4, 5];

// Automatic inference
let names = ["Alice", "Bob"]; // Inferred as string[2]
```

### 3.2 Slices
Slices are references to underlying arrays and have variable length. This is the most commonly used sequence type in MyGo.

```mygo
// Type representation: Type[] (no length)
let numbers: int[] = [1, 2, 3]; 

// Create slice from array
let arr = [10, 20, 30, 40];
let slice = arr[1..3]; // Includes index 1, 2 (20, 30)
```

#### Under the Hood: Slice Header
*   **Three Fields**: A slice consists of a pointer to the underlying array, a length, and a capacity.
*   **Reference Semantics**: Passing a slice passes these three fields by value, but the underlying data is shared.
*   **Safety**: Bounds checking is performed at runtime. Accessing `slice[len]` causes a panic.

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
