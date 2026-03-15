# MyGo 数据结构 (Data Structures)

MyGo 提供了结构体 (`struct`)、枚举 (`enum`) 和数组/切片等核心数据结构，支持泛型和面向对象编程范式。

## 1. 结构体 (Struct)

结构体是 MyGo 中自定义数据类型的核心，用于组合多个不同类型的字段。

### 语法定义
```antlr
structDecl: annotationUsage* whereClause? modifier? 'struct' ID typeParams? '{' (structField (',' structField)* ','?)? '}' ;
structField: ID ':' typeType (STRING)? ;
```

### 1.1 定义结构体
结构体字段的可见性完全由**首字母大小写**决定：首字母大写为公开 (Public)，小写为私有 (Private)。结构体本身可以使用修饰符（如 `pub`）控制可见性。

```mygo
// 定义一个简单的 Point 结构体
struct Point {
    x: int,
    y: int
}

// 带修饰符的结构体 (pub 表示结构体本身对包外可见)
pub struct User {
    ID: int,            // 首字母大写：字段对包外可见
    Name: string,
    Email: string,
    age: int,           // 首字母小写：字段仅包内可见
    secret: string      // 首字母小写：字段仅包内可见
}

// 带标签 (Tag) 的结构体 (常用于 JSON 序列化)
struct Config {
    Host: string "json:\"host\"",
    Port: int    "json:\"port\""
}

// 泛型结构体
struct Box<T> {
    Value: T
}
```

### 1.2 实例化与使用
MyGo 支持类似于 JSON 或 Rust 的结构体初始化语法。

#### 显式初始化
```mygo
fn main() {
    // 实例化
    let p = Point{x: 10, y: 20};
    
    // 访问字段
    fmt.Println(p.x);
    
    // 修改字段 (如果变量是可变的)
    p.x = 100;
}
```

#### 底层原理：内存布局 (Under the Hood: Memory Layout)
*   **连续内存**：MyGo 的结构体在内存中是连续布局的。字段按照定义的顺序存储。
*   **内存对齐**：编译器可能会在字段之间插入填充字节 (Padding)，以满足目标架构的内存对齐要求（例如，将 `int64` 对齐到 8 字节边界）。
*   **值语义**：结构体是值类型。将一个结构体赋值给另一个变量会拷贝整个数据结构。

#### 零值初始化 (未赋值声明)
当你声明一个结构体变量而不赋初值时，MyGo 会将其编译为 Go 的变量声明，从而触发**零值初始化**。所有字段都会被赋予其类型的默认值（如 `int` 为 `0`, `string` 为 `""`）。

**注意：未初始化的结构体不是 `nil`。**

```mygo
fn main() {
    let p: Point; // 此时 p.x = 0, p.y = 0
    fmt.Println(p.x); // 输出 0
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

#### 底层原理：带标签联合体 (Under the Hood: Tagged Unions)
*   **表示**：带数据的枚举被编译为一个包含隐藏的 "Tag" 字段（判别符）和数据字段的结构体。
*   **联合存储**：在内存中，不同变体的数据字段共享同一块内存空间（类似 C 语言的 union），其大小取决于最大的变体。
*   **类型安全**：编译器确保你只能访问与当前 Tag 匹配的数据，通常通过 `match` 语句实现。

---

## 3. 数组与切片 (Arrays & Slices)

MyGo 区分固定长度的数组和动态长度的切片。

### 3.1 数组 (Array)
数组长度固定，是值类型。

```mygo
// 类型表示: 类型[长度]
let arr: int[5] = [1, 2, 3, 4, 5];

// 自动推导
let names = ["Alice", "Bob"]; // 推导为 string[2]
```

### 3.2 切片 (Slice)
切片是对底层数组的引用，长度可变。这是 MyGo 中最常用的序列类型。

```mygo
// 类型表示: 类型[] (无长度)
let numbers: int[] = [1, 2, 3]; 

// 通过数组创建切片
let arr = [10, 20, 30, 40];
let slice = arr[1..3]; // 包含索引 1, 2 (20, 30)
```

#### 底层原理：切片头 (Under the Hood: Slice Header)
*   **三个字段**：切片在运行时由三部分组成：指向底层数组的指针、长度 (len) 和容量 (cap)。
*   **引用语义**：传递切片时，拷贝的是这三个字段（浅拷贝），但底层的数组数据是共享的。
*   **安全性**：运行时会进行边界检查。访问 `slice[len]` 会触发 Panic。

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

MyGo 中的 `Map` 是一个强大的内置引用类型，对应 Go 的 `map[K]V`。

### 4.1 声明与初始化
Map 使用 `Map<KeyType, ValueType>()` 构造函数进行初始化（底层对应 `make(map[K]V)`）。目前**不支持**大括号字面量初始化。

```mygo
// 显式声明并初始化
let scores: Map<string, int> = Map<string, int>();

// 类型推导初始化
let counts = Map<int, string>();
```

### 4.2 操作与注意事项
- **键的唯一性**：Map 是哈希表，键必须是唯一的。
- **引用类型**：Map 是引用类型。将其赋值给另一个变量后，两者指向同一个底层哈希表。
- **空值检查**：访问不存在的键会返回该值类型的**零值**。建议配合 `contains` 方法（如可用）或通过 Go 的 `ok` 习语进行检查。
- **并发安全**：Map 本身不是并发安全的，在并发场景下建议加锁或使用专用并发库。

```mygo
let scores = Map<string, int>();
scores["Alice"] = 100;

let s = scores["Alice"]; // s = 100
let nonExistent = scores["Unknown"]; // nonExistent = 0 (int 的零值)
```


---

## 5. 元组 (Tuples)

元组可以将多个不同类型的值组合成一个复合值。

```mygo
let pair: (string, int) = ("Alice", 30);
let (name, age) = pair;
```
