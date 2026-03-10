# MyGo 数据结构 (Data Structures)

MyGo 提供了结构体 (`struct`)、枚举 (`enum`) 和数组/切片等核心数据结构，支持泛型和面向对象编程范式。

## 1. 结构体 (Struct)

结构体是 MyGo 中自定义数据类型的核心，用于组合多个不同类型的字段。

### 语法定义
```antlr
structDecl: whereClause? modifier? 'struct' ID typeParams? '{' (structField (',' structField)* ','?)? '}' ;
structField: ID ':' typeType ;
```

### 1.1 定义结构体
结构体字段名通常首字母大写以表示公开（配合 `pub` 关键字），或小写表示私有。

```mygo
// 定义一个简单的 Point 结构体
struct Point {
    x: int,
    y: int
}

// 带修饰符的结构体 (pub 表示结构体本身对包外可见)
pub struct User {
    pub ID: int,        // 字段对包外可见
    pub Name: string,
    pub Email: string,
    age: int,           // 字段仅包内可见
    pri secret: string  // 字段仅当前文件可见
}

// 泛型结构体
struct Box<T> {
    Value: T
}
```

### 1.2 实例化与使用
MyGo 支持类似于 JSON 或 Rust 的结构体初始化语法。

```mygo
fn main() {
    // 实例化
    let p = Point{x: 10, y: 20};
    
    // 访问字段
    fmt.Println(p.x);
    
    // 修改字段 (如果变量是可变的)
    p.x = 100;
    
    // 泛型实例化
    let intBox = Box<int>{Value: 123};
    let strBox = Box<string>{Value: "MyGo"};
}
```

---

## 2. 枚举 (Enum)

MyGo 的枚举是强大的代数数据类型 (Algebraic Data Types, ADT)，不仅可以定义常量集合，还可以携带关联数据。

### 语法定义
```antlr
enumDecl: whereClause? modifier? 'enum' ID typeParams? '{' enumVariant (',' enumVariant)* ','? '}' ;
enumVariant: ID ('(' typeList ')')? ;
```

### 2.1 简单枚举
类似于 C/Go 的枚举，用于表示一组命名的常量。

```mygo
enum Color {
    Red,
    Green,
    Blue
}
```

### 2.2 带数据的枚举 (Tagged Union)
枚举成员可以携带不同类型的数据，这在处理状态机或消息传递时非常有用。

```mygo
enum Message {
    Quit,                       // 无数据
    Move(int, int),             // 携带两个 int (x, y)
    Write(string),              // 携带一个 string
    ChangeColor(int, int, int)  // 携带三个 int (r, g, b)
}
```

### 2.3 使用 Match 处理枚举
`match` 语句是处理枚举的最佳方式，它能确保所有情况都被覆盖（穷尽性检查）。

```mygo
fn process(msg: Message) {
    match msg {
        Message.Quit => fmt.Println("Quit"),
        Message.Move(x, y) => fmt.Printf("Move to %d, %d\n", x, y),
        Message.Write(s) => fmt.Println("Message:", s),
        // 使用 other 处理未列出的情况，或者列出所有情况
        other => fmt.Println("Other message")
    }
}
```

---

## 3. 数组与切片 (Arrays & Slices)

MyGo 区分固定长度的数组和动态长度的切片。

### 3.1 数组 (Array)
数组长度固定，是值类型。

```mygo
// 类型表示: [长度]类型
let arr: [5]int = [1, 2, 3, 4, 5];

// 自动推导
let names = ["Alice", "Bob"]; // 推导为 [2]string
```

### 3.2 切片 (Slice)
切片是对底层数组的引用，长度可变。这是 MyGo 中最常用的序列类型。

```mygo
// 类型表示: []类型 (无长度)
let numbers: []int = [1, 2, 3]; 

// 通过数组创建切片
let arr = [10, 20, 30, 40];
let slice = arr[1..3]; // 包含索引 1, 2 (20, 30)
```

### 3.3 常用操作
MyGo 的切片操作与 Go 语言高度一致。

```mygo
let list = [1, 2, 3];

// 获取长度
let l = len(list);

// 访问元素
let first = list[0];

// 修改元素
list[1] = 100;

// 遍历
for (v : list) {
    fmt.Println(v);
}
```

---

## 4. 映射 (Map)

MyGo 目前通过泛型库或互操作支持 Map，标准库中通常提供 `Map<K, V>` 类型。

```mygo
// 示例：假设标准库提供了 Map 类型
let scores = Map<string, int>{};
scores.Put("Alice", 100);
scores.Put("Bob", 95);

let aliceScore = scores.Get("Alice");
```

---

## 5. 元组 (Tuples)

元组可以将多个不同类型的值组合成一个复合值。

```mygo
let pair: (string, int) = ("Alice", 30);
let (name, age) = pair;
```
