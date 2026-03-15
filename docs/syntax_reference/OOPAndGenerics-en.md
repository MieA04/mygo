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

#### Under the Hood: Method Binding
When the compiler encounters a `trait bind` block, the **MethodCollector** performs the following steps:
1.  **Resolve Targets**: It resolves the target type (e.g., `Circle`) in the symbol table.
2.  **Process Combs**: It iterates through traits listed in `combs(...)` (e.g., `Shape`).
    *   It checks for method conflicts between traits.
    *   It applies `ban` and `flip ban` directives to filter methods.
3.  **Attach Methods**: It attaches the final set of methods (including those implemented in the block and those inherited from traits) to the target type's symbol.
4.  **Symbol Table**: The target type's symbol now effectively "owns" these methods, making them available for method calls and interface satisfaction checks.

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

During Trait composition and reuse, MyGo provides powerful directives to finely control method visibility and conflict resolution. This allows for precise "Interface Segregation" at compile time.

### 2.1 Ban (Method Pruning)

The `ban` directive is used to **explicitly forbid** the implementation or export of certain methods from composed traits. This is useful when a type inherits most behaviors from a trait but explicitly rejects some.

**Semantics**:
- Methods listed in `ban` are removed from the target type's method set.
- Calling a banned method results in a **compile-time error**.
- Only affects methods with the specified names; others are inherited normally.

```mygo
trait Logger {
    fn log(msg: string);
    fn flush();
}

struct ReadOnlyFile { ... }

trait bind (ReadOnlyFile) combs(Logger) {
    // Forbid log method, calling it will cause a compilation error
    ban [log]; 
    
    // flush() is inherited normally
    fn flush() { ... }
}
```

### 2.2 Flip Ban (Whitelist)

`flip ban` is the inverse of `ban`. It allows you to **explicitly specify which methods to keep** and from which Trait, pruning everything else with the same name.

**Syntax**: `flip ban [MethodName : TraitName];`

**Semantics**:
1.  For methods listed, keep the implementation from the specified Trait.
2.  For methods with the same name but NOT listed, remove them (prune).
3.  Methods with unique names (no conflict) are inherited normally.

```mygo
trait Logger {
    fn log(msg: string);
    fn flush();
}

trait Auditor {
    fn log(event: string);
    fn archive();
}

struct SecurityModule { ... }

trait bind (SecurityModule) combs(Logger, Auditor) {
    // Conflict resolution:
    // Keep Logger's log method, discard Auditor's log method.
    flip ban [log: Logger];
    
    // flush() (from Logger) and archive() (from Auditor) are inherited automatically
}
```

#### Under the Hood: Conflict Resolution
The compiler maintains a `TraitCompositionContext` during the binding process:
1.  **Directives Collection**: It first collects all `ban` and `flip ban` directives.
2.  **Merge & Filter**: When merging methods from composed traits:
    *   If a method is in the `BannedMethods` set, it is skipped.
    *   If a method exists in the `FlippedMethods` map:
        *   If the source trait matches the one specified in `flip ban`, it is kept.
        *   Otherwise, it is discarded.
    *   If a method conflict occurs and no directive resolves it, a semantic error is reported.

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

#### Under the Hood: Type Erasure vs Monomorphization
*Current Design*: MyGo primarily uses **monomorphization** (similar to Rust/C++ templates).
*   When a generic function or type is instantiated with concrete types, the compiler generates a specialized version of the code for those types.
*   This ensures zero-cost abstractions but may increase binary size.
*   Constraints are checked at compile-time to ensure the concrete types satisfy the `where` clauses.
