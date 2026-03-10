# MyGo 面向对象与泛型 (OOP & Generics)

MyGo 摒弃了传统的类继承机制，采用 **Struct (数据)** + **Trait (行为)** 的组合模式。这种设计深受 Rust 和 Swift 的影响，旨在提供更灵活、更安全的抽象能力。

## 1. Trait 系统 (Trait System)

Trait 定义了一组行为契约（接口）。任何实现了这些行为的类型都满足该 Trait。

### 语法定义
```antlr
traitDecl
    : 'trait' ID typeParams? '{' traitFnDecl* '}'                                # PureTraitDecl
    | 'trait' 'bind' typeParams? '(' bindTarget ('|' bindTarget)* ')' 
      ('combs' '(' ID (',' ID)* ')')? '{' traitBodyItem* '}'                     # BindTraitDecl
    ;
```

### 1.1 定义 Trait
```mygo
trait Shape {
    fn Area(): float;
    fn Perimeter(): float;
}

trait Display {
    fn ToString(): string;
}
```

### 1.2 实现 Trait (Bind Trait)
MyGo 使用 `trait bind` 语法将行为绑定到数据上。这是 MyGo 最具特色的语法之一，它允许你为已有类型（包括外部类型）添加方法或实现 Trait。

```mygo
struct Circle {
    Radius: float
}

// 为 Circle 实现 Shape Trait
trait bind (Circle) combs(Shape) {
    fn Area(): float {
        return 3.14 * this.Radius * this.Radius;
    }
    
    fn Perimeter(): float {
        return 2.0 * 3.14 * this.Radius;
    }
}
```

### 1.3 扩展方法 (Extension Methods)
即使不实现特定的 Trait，也可以使用 `trait bind` 为类型直接添加方法。

```mygo
// 为 Circle 添加一个自定义方法
trait bind (Circle) {
    fn Scale(factor: float) {
        this.Radius = this.Radius * factor;
    }
}
```

### 1.4 多类型绑定 (Multi-Type Binding)
可以将同一组逻辑绑定到多个类型上，结合 `match this` 实现共享逻辑。

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

## 2. 编译器指令 (Compiler Directives)

在 Trait 组合与复用过程中，MyGo 提供了强大的指令来精细控制方法的可见性与冲突解决。

### 2.1 Ban (禁止)
`ban` 指令用于显式禁止某个方法的实现或导出。

```mygo
trait ReadWrite {
    fn Read();
    fn Write();
}

struct ReadOnlyFile { ... }

trait bind (ReadOnlyFile) combs(ReadWrite) {
    // 禁止 Write 方法，调用时会产生编译错误
    ban [Write]; 
    
    fn Read() { ... }
}
```

### 2.2 Flip Ban (反转禁止/白名单)
`flip ban` 表示"除了列表中的方法外，禁止其他所有方法"。

```mygo
trait bind (ReadOnlyFile) combs(ReadWrite) {
    // 只允许 Read，隐含 ban [Write]
    flip ban [Read];
    
    fn Read() { ... }
}
```

### 2.3 Repeat (重复策略)
当组合多个 Trait 且存在同名方法时，可以使用 `ban repeat` 策略来处理冲突（具体行为取决于编译器实现，通常用于解决菱形继承问题）。

---

## 3. 泛型 (Generics)

MyGo 的泛型系统贯穿了结构体、函数和 Trait。

### 3.1 泛型约束 (Where Clauses)
MyGo 采用前置 `where` 子句来声明泛型约束，这使得函数签名更加整洁。

#### 语法定义
```antlr
whereClause: 'where' genericConstraint (',' genericConstraint)* ;
genericConstraint: ID ':' typeType ('+' typeType)* ;
```

#### 示例
```mygo
// 定义泛型函数，要求 T 实现 Shape 和 Display
where T: Shape + Display
fn printInfo<T>(item: T) {
    fmt.Println(item.ToString());
    fmt.Println("Area:", item.Area());
}
```

### 3.2 泛型 Trait
Trait 本身也可以带类型参数。

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
