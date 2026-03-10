# MyGo 函数与闭包 (Functions & Closures)

函数是 MyGo 程序执行的基本单元。MyGo 支持具名函数、匿名函数（闭包）以及高阶函数，并且深度集成了泛型系统。

## 1. 函数声明 (Function Declaration)

使用 `fn` 关键字定义函数。MyGo 的函数定义非常灵活，支持多返回值和泛型约束。

### 语法定义
```antlr
fnDecl: whereClause? modifier? 'fn' ID typeParams? '(' paramList? ')' (':' typeType)? block ;
```

### 1.1 基础示例
```mygo
// 无参数无返回值
fn sayHello() {
    fmt.Println("Hello!");
}

// 带参数和单返回值
fn add(a: int, b: int): int {
    return a + b;
}
```

### 1.2 多返回值 (Multiple Return Values)
MyGo 支持多返回值，底层通过元组实现。

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

## 2. 匿名函数与闭包 (Lambdas / Closures)

MyGo 支持轻量级的 Lambda 表达式语法，可以捕获外部作用域的变量。

### 语法定义
```antlr
lambdaExpr: '(' paramList? ')' (':' typeType)? '=>' block ;
```

### 2.1 基础用法
```mygo
fn main() {
    // 定义一个匿名函数并赋值给变量
    let multiply = (a: int, b: int): int => {
        return a * b;
    };
    
    fmt.Println(multiply(3, 4)); // 输出 12
}
```

### 2.2 闭包 (Closures)
闭包可以引用其外部作用域中定义的变量。

```mygo
fn main() {
    let factor = 10;
    
    // 捕获外部变量 factor
    let scaler = (x: int): int => {
        return x * factor;
    };
    
    fmt.Println(scaler(5)); // 输出 50
}
```

### 2.3 作为参数传递 (High-Order Functions)
函数在 MyGo 中是一等公民，可以作为参数传递给其他函数。

```mygo
// 定义一个接受函数作为参数的函数
// 参数 op 的类型是 fn(int): int
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

## 3. 泛型函数 (Generic Functions)

函数可以定义类型参数，实现泛型编程。MyGo 的泛型支持强大的类型约束。

### 3.1 基础泛型
```mygo
fn printIfEqual<T>(a: T, b: T) {
    if a == b {
        fmt.Println("Equal");
    }
}
```

### 3.2 泛型约束 (Where Clauses)
可以使用 `where` 子句对泛型参数施加约束，要求其实现特定的 Trait。
**注意：MyGo 中 `where` 子句必须写在函数声明的最前面。**

```mygo
// 要求 T 必须实现 Addable Trait
where T: Addable
fn addGeneric<T>(a: T, b: T): T {
    return a + b;
}

// 多个约束
where T: Show + Debug
fn printDetails<T>(item: T) {
    fmt.Println(item.String());
}
```

---

## 4. 函数类型 (Function Types)

函数类型用于描述函数的签名，通常用于变量声明或参数类型标注。

### 语法结构
```antlr
typeType: 'fn' '(' typeList? ')' (':' typeType)? ;
```

### 示例
```mygo
// 声明一个变量，类型为"接受两个int返回一个int的函数"
let operation: fn(int, int): int;

operation = add;
```
