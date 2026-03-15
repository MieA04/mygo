# 注解与静态反射 (Annotations & Static Reflection)

MyGo 提供了强大的编译期元编程能力，主要通过**注解 (Annotations)** 和**静态反射 (Static Reflection)** 来实现。与 Go 语言运行时的 `reflect` 包不同，MyGo 的反射机制在编译期执行，零运行时开销，且类型安全。

## 1. 注解 (Annotations)

注解是一种将元数据附加到声明（如结构体、函数、字段等）上的机制。MyGo 的注解以 `@` 开头，支持传递参数。

### 1.1 内置注解

MyGo 目前支持以下内置注解：

#### `@Derive(MacroName)`
用于自动派生代码（宏展开）。当编译器遇到 `@Derive` 时，会查找对应的宏定义并执行它，将生成的代码注入到当前结构体或上下文中。

**语法：**
```mygo
@Derive(MacroName)
struct User {
    name: string
}
```

**工作原理：**
1. 编译器解析到 `@Derive(Json)`。
2. 查找名为 `DeriveJson` 的宏定义（通常是一个编译期执行的函数或脚本）。
3. 执行宏，传入当前结构体的元数据（MetaValue）。
4. 将宏返回的代码字符串插入到 AST 中。

### 1.2 自定义注解
虽然目前编译器主要处理内置注解，但在未来版本中，开发者可以通过定义宏来处理自定义注解。

### 1.3 字段标签 (Tags)
MyGo 支持类似 Go 的结构体字段标签，用于序列化库等场景。

**语法：**
```mygo
struct User {
    name: string "json:\"name\" xml:\"Name\""
    age: int     "json:\"age\""
}
```

---

## 2. 静态反射 (Static Reflection)

静态反射允许在编译期检查和操作类型信息。MyGo 提供了一组编译期内置函数和元数据对象来实现这一功能。

### 2.1 编译期内置函数

#### `find_all_annotated_with(annotationName: string) -> List<MetaValue>`
查找当前作用域内所有标记了指定注解的符号。

**示例：**
```mygo
// 查找所有标记了 @Entity 的结构体
let entities = find_all_annotated_with("Entity");
for entity in entities {
    println("Found entity: " + entity.name);
}
```

#### `get_tag(tagString: string, key: string) -> string`
解析字段标签字符串，获取指定键的值。

**示例：**
```mygo
let tag = "json:\"name\" xml:\"Name\"";
let jsonKey = get_tag(tag, "json"); // 返回 "name"
```

### 2.2 元数据对象 (MetaValue)

在编译期宏或反射上下文中，符号（Symbol）被表示为 `MetaValue` 对象。该对象包含以下属性：

| 属性名 | 类型 | 描述 |
| :--- | :--- | :--- |
| `name` | `string` | 符号在 MyGo 中的名称（如 `User`） |
| `go_name` | `string` | 符号在生成的 Go 代码中的名称（通常包含可见性处理） |
| `kind` | `string` | 符号类型（如 `struct`, `func`, `enum` 等） |
| `pkg` | `string` | 所属包名 |
| `annotations` | `List<MetaValue>` | 该符号上的注解列表 |
| `fields` | `List<MetaValue>` | （仅结构体）字段列表 |

**注解对象属性：**
- `name`: 注解名称（如 `Derive`）
- `args`: 参数列表

**字段对象属性：**
- `name`: 字段名
- `type`: 字段类型字符串
- `tag`: 字段标签字符串

### 2.3 宏与元编程 (Macros)

MyGo 的宏系统基于解释器在编译期执行。宏可以访问 `target` 变量，它代表了当前正在处理的符号的元数据。

**为什么生成 Go 代码？**
MyGo 编译器目前通过 Transpiler 模式将 MyGo 代码转换为 Go 代码。宏在编译期执行，其返回值会被注入到最终生成的 Go 源代码中。

**支持生成 MyGo 代码**
从新版本开始，宏也可以生成 **MyGo 代码**。编译器会自动尝试解析宏返回的字符串：
1. 如果是合法的 MyGo 代码块（如函数定义），编译器会将其转译为 Go 代码后注入。
2. 如果解析失败（包含 Go 特有语法），编译器会将其视为原始 Go 代码直接注入。

**宏定义示例 1：生成 Go 代码（底层控制）**
```mygo
// 定义一个宏，用于生成 String() 方法 (Go 语法)
@macro DeriveToString {
    let structName = target.go_name; // 获取结构体在 Go 中的名称（处理了可见性）
    let fields = target.fields;
    
    // 开始构建 Go 函数代码
    let code = "func (s *" + structName + ") String() string {\n";
    code += "    return \"" + structName + "{\" + \n";
    
    for (f : fields) {
        // 使用 fmt.Sprintf 格式化字段值，注意这里生成的是 Go 代码
        code += "        \"" + f.name + ":\" + fmt.Sprintf(\"%v\", s." + f.name + ") + \",\" + \n";
    }
    
    code += "    \"}\"\n";
    code += "}\n";
    
    return code;
}
```

**宏定义示例 2：生成 MyGo 代码（推荐）**
编写 MyGo 代码通常更简洁，且能利用 MyGo 的语法糖。

```mygo
// 定义一个宏，生成一个打印问候的函数 (MyGo 语法)
@macro DeriveHello {
    let name = target.name;
    // MyGo 函数语法: fn functionName() { ... }
    return "fn hello_" + name + "() { println(\"Hello from " + name + "!\"); }";
}
```

**使用宏：**
```mygo
@Derive(ToString)
@Derive(Hello)
struct Point {
    x: int
    y: int
}

fn main() {
    let p = Point{x: 1, y: 2};
    println(p.String()); // 调用生成的 Go 方法
    hello_Point();       // 调用生成的 MyGo 函数
}
```

---

## 3. 最佳实践

1. **零成本抽象**：尽可能使用静态反射代替运行时反射。例如，使用宏生成序列化代码，而不是在运行时使用 `reflect` 包。
2. **标签规范**：字段标签应遵循 Go 的 `key:"value"` 格式，以便兼容现有的 Go 生态库（如 `encoding/json`）。
3. **宏的调试**：由于宏在编译期执行，可以通过 `println` 在编译日志中打印调试信息。

## 4. 常见问题

**Q: MyGo 的反射能在运行时使用吗？**
A: 不可以。MyGo 的反射是静态的（编译期）。如果需要运行时反射，可以使用 MyGo 调用 Go 标准库的 `reflect` 包，但这会带来运行时开销。

**Q: 如何获取一个变量的类型名称？**
A: 目前可以通过宏在编译期获取，或者在运行时使用 Go 的 `%T` 格式化动词。

**Q: `@Derive` 支持多个宏吗？**
A: 支持。可以写成 `@Derive(MacroA, MacroB)`，编译器会依次执行它们。
