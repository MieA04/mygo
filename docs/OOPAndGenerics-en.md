# MyGo OOP & Generics

MyGo discards traditional class inheritance, adopting a **Struct (Data)** + **Trait (Behavior)** composition model. This design is heavily influenced by Rust and Swift, aiming to provide more flexible and safer abstractions.

## 1. Trait System

A Trait defines a set of behavioral contracts (interfaces). Any type that implements these behaviors satisfies the Trait.

### Syntax Definition
```antlr
traitDecl
    : 'trait' ID typeParams? '{' traitFnDecl* '}'                                # PureTraitDecl
    | 'trait' 'bind' typeParams? '(' bindTarget ('|' bindTarget)* ')' 
      ('combs' '(' ID (',' ID)* ')')? '{' traitBodyItem* '}'                     # BindTraitDecl
    ;
```

### 1.1 Defining Traits
```mygo
trait Shape {
    fn Area(): float;
    fn Perimeter(): float;
}

trait Display {
    fn ToString(): string;
}
```

### 1.2 Implementing Traits (Bind Trait)
MyGo uses `trait bind` syntax to bind behaviors to data. This is one of MyGo's most distinctive features, allowing you to add methods or implement traits for existing types (including external types).

```mygo
struct Circle {
    Radius: float
}

// Implement Shape Trait for Circle
trait bind (Circle) combs(Shape) {
    fn Area(): float {
        return 3.14 * this.Radius * this.Radius;
    }
    
    fn Perimeter(): float {
        return 2.0 * 3.14 * this.Radius;
    }
}
```

### 1.3 Extension Methods
Even without implementing a specific Trait, you can use `trait bind` to add methods directly to a type.

```mygo
// Add a custom method to Circle
trait bind (Circle) {
    fn Scale(factor: float) {
        this.Radius = this.Radius * factor;
    }
}
```

### 1.4 Multi-Type Binding
You can bind the same logic to multiple types, using `match this` to implement shared logic.

```mygo
struct Rect { W: float, H: float }

trait bind (Circle | Rect) combs(Shape) {
    fn Area(): float {
        match this {
            is Circle => return 3.14 * this.Radius * this.Radius,
            is Rect => return this.W * this.H,
        }
    }
    // ...
}
```

---

## 2. Compiler Directives

During Trait composition and reuse, MyGo provides powerful directives to finely control method visibility and conflict resolution.

### 2.1 Ban
The `ban` directive is used to explicitly forbid the implementation or export of a certain method.

```mygo
trait ReadWrite {
    fn Read();
    fn Write();
}

struct ReadOnlyFile { ... }

trait bind (ReadOnlyFile) combs(ReadWrite) {
    // Forbid Write method, calling it will cause a compilation error
    ban [Write]; 
    
    fn Read() { ... }
}
```

### 2.2 Flip Ban (Whitelist)
`flip ban` means "forbid all other methods except those listed".

```mygo
trait bind (ReadOnlyFile) combs(ReadWrite) {
    // Only allow Read, implying ban [Write]
    flip ban [Read];
    
    fn Read() { ... }
}
```

### 2.3 Repeat (Repeat Strategy)
When combining multiple Traits with methods of the same name, `ban repeat` strategy can be used to handle conflicts (specific behavior depends on compiler implementation, usually used to solve diamond inheritance problems).

---

## 3. Generics

MyGo's generics system permeates structs, functions, and Traits.

### 3.1 Generic Constraints (Where Clauses)
MyGo uses preceding `where` clauses to declare generic constraints, making function signatures cleaner.

#### Syntax Definition
```antlr
whereClause: 'where' genericConstraint (',' genericConstraint)* ;
genericConstraint: ID ':' typeType ('+' typeType)* ;
```

#### Example
```mygo
// Define generic function, requiring T to implement Shape and Display
where T: Shape + Display
fn printInfo<T>(item: T) {
    fmt.Println(item.ToString());
    fmt.Println("Area:", item.Area());
}
```

### 3.2 Generic Traits
Traits themselves can also carry type parameters.

```mygo
trait Converter<From, To> {
    fn Convert(input: From): To;
}

struct StringToInt {}

trait bind (StringToInt) combs(Converter<string, int>) {
    fn Convert(input: string): int {
        // ... implementation
        return 0;
    }
}
```
