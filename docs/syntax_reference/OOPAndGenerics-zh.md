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

#### 底层原理：方法绑定 (Under the Hood: Method Binding)
当编译器遇到 `trait bind` 块时，**MethodCollector** 会执行以下步骤：
1.  **解析目标 (Resolve Targets)**：在符号表中解析目标类型（如 `Circle`）。
2.  **处理组合 (Process Combs)**：遍历 `combs(...)` 中列出的 Trait（如 `Shape`）。
    *   检查 Trait 间的方法冲突。
    *   应用 `ban` 和 `flip ban` 指令过滤方法。
3.  **挂载方法 (Attach Methods)**：将最终的方法集合（包括块内实现的和从 Trait 继承的）挂载到目标类型的符号上。
4.  **符号表更新**：目标类型的符号现在实际上"拥有"了这些方法，使其可用于方法调用和接口满足性检查。

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

## 2. Trait 组合指令 (Trait Composition Directives)

在传统的面向对象语言中，实现一个接口通常意味着必须实现其所有方法。这往往会导致"过度承诺"——为了满足类型约束,开发者被迫为不适用的子类编写空方法或在运行时抛出异常（例如，强制要求"鸵鸟"实现鸟类的 `fly()` 方法）。

为了提供更精确的类型描述并贯彻**接口隔离原则**，MyGo 引入了 Trait 组合指令。它允许开发者在绑定 Trait 时，在**编译期**对行为集合进行精确裁剪，彻底消灭运行时的无效调用错误。

---

### 2.1 Ban (同名方法全量裁剪)

`ban` 指令用于**一刀切地剔除所有被组合 Trait 中的同名方法**。当多个 Trait 包含相同名称的方法时，`ban` 会将它们全部从当前类型的行为签名中抹除。

**作用域**：仅针对**同名方法**。不同名的方法不受影响，依然会被正常继承。

**语义**：被 `ban` 裁剪的方法将彻底消失，任何试图调用该方法的代码都会在**编译期直接报错**。

```mygo
trait Logger {
    fn log(msg: string);
    fn flush();
}

trait Auditor {
    fn log(event: string); // 与 Logger.log 同名
    fn archive();
}

struct SecurityModule { ... }

trait bind (SecurityModule) combs(Logger, Auditor) {
    // ban [log] 会同时剔除 Logger.log 和 Auditor.log
    // flush() 和 archive() 不受影响，依然可用
    ban [log];

    fn flush() { ... }
    fn archive() { ... }
}
```

---

### 2.2 Flip Ban (精确白名单保留)

`flip ban` 是 `ban` 的反向操作，用于**精确指定要保留的方法及其来源 Trait**。它采用 `method: Trait` 语法，明确告诉编译器："在所有同名方法中，我只要这一个"。

**作用域**：仅针对**同名方法**。不同名的方法不受影响，依然会被正常继承。

**语义**：
1. 对于列表中指定的方法，保留其来源 Trait 的实现。
2. 对于同名但未被列出的方法，全部裁剪。
3. 对于不同名的方法，正常继承，无需在 `flip ban` 中声明。

```mygo
trait Logger {
    fn log(msg: string) { print("[LOG] " + msg); }
    fn flush();
}

trait Auditor {
    fn log(event: string) { print("[AUDIT] " + event); }
    fn archive();
}

struct SecurityModule { ... }

trait bind (SecurityModule) combs(Logger, Auditor) {
    // 精确保留 Logger 的 log，裁剪 Auditor 的 log
    // flush() 和 archive() 自动继承，无需声明
    flip ban [log: Logger];

    // 无需重写 log()，直接继承 Logger.log 的实现
    fn flush() { ... }
    fn archive() { ... }
}
```

**覆盖规则**：如果在 Trait 绑定体中显式重写了已通过 `flip ban` 引入的方法，则新实现将覆盖原有的 Trait 实现。

```mygo
trait bind (SecurityModule) combs(Logger) {
    flip ban [log: Logger];

    // 显式覆盖 Logger.log 的实现
    fn log(msg: string) {
        print("[SECURITY] " + msg);
    }
}
```

#### 底层原理：冲突解决 (Under the Hood: Conflict Resolution)
编译器在绑定过程中维护一个 `TraitCompositionContext`：
1.  **收集指令**：首先收集所有的 `ban` 和 `flip ban` 指令。
2.  **合并与过滤**：在合并组合 Trait 的方法时：
    *   如果方法在 `BannedMethods` 集合中，则跳过。
    *   如果方法在 `FlippedMethods` 映射中：
        *   如果来源 Trait 与 `flip ban` 指定的一致，则保留。
        *   否则，丢弃。
    *   如果发生冲突且没有指令解决，则报告语义错误。

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

#### 底层原理：类型擦除与单态化 (Under the Hood: Monomorphization)
*当前设计*：MyGo 主要采用**单态化 (Monomorphization)**（类似 Rust/C++ 模板）。
*   当泛型函数或类型被实例化时，编译器会为具体类型生成专门的代码版本。
*   这保证了零成本抽象，但可能增加二进制文件的大小。
*   泛型约束在编译期检查，确保具体类型满足 `where` 子句的要求。
