# RFC-008: 静态反射与增强型元编程

- **状态**: 草案 (Draft)
- **作者**: Trae AI
- **创建日期**: 2026-03-11
- **目标版本**: MyGo v0.3.0

## 1. 摘要 (Abstract)
本 RFC 提议为 MyGo 引入一套全面的**静态反射（编译期反射）**机制。其目标是在编译过程中将详细的类型信息（如结构体字段、Tag 和方法签名）暴露给宏系统。这将使强大的代码生成功能（如序列化、ORM 映射、依赖注入）能够在没有运行时开销的情况下实现，使 MyGo 与 Rust（过程宏）和 Zig (comptime) 等现代系统编程语言保持一致。

## 2. 动机 (Motivation)
在目前的实现中（RFC-007），MyGo 的元编程能力仅限于：
1.  **符号发现**: 查找带有特定注解的符号 (`find_all_annotated_with`)。
2.  **基础元数据**: 访问符号名称、包名和类型种类。
3.  **原始主体访问**: 以字符串形式获取函数体的源代码。

**局限性:**
-   **缺乏字段内省**: 宏无法遍历结构体的字段。无法生成依赖于类型结构的代码（例如 `toString()`, `toJson()`）。
    -   **缺乏元数据 Tag**: 结构体字段无法携带元数据（类似于 Go 的结构体 Tag `` `json:"id"` `` 或 Rust 的属性 `#[serde(rename = "id")]`）。
-   **字段无序**: 当前编译器的符号表将字段存储在 `map[string]string` 中，这丢失了原始声明顺序——这对于二进制序列化或 C 语言互操作至关重要。

为了支持 `@Derive(Json)` 或 ORM 框架等特性，我们需要一种在编译期检查类型的方法。

## Design Proposal (设计提案)

### 3.1 语法扩展 (结构体 Tag)
我们提议扩展语法，允许在结构体的字段声明后添加可选的字符串字面量作为 Tag。

**提议的 EBNF 变更:**
```ebnf
// 当前
structField: ID ':' typeType ;

// 提议
structField: ID ':' typeType (STRING)? ;
```

**用法示例:**
```mygo
struct User {
    id: int "json:\"id\" db:\"primary_key\"",
    name: string "json:\"name\"",
    age: int // 可选 Tag
}
```

### 3.2 符号表重构
必须更新 `Struct` 符号的内部表示，以保留字段顺序并存储 Tag。

**当前 (`pkg/compiler/symbols/symbol.go`):**
```go
type Symbol struct {
    // ...
    Fields map[string]string // 名称 -> 类型
}
```

**提议:**
```go
type FieldSymbol struct {
    Name string
    Type string
    Tag  string // 原始 Tag 字符串，例如 "json:\"id\""
}

type Symbol struct {
    // ...
    Fields []FieldSymbol // 有序字段列表
    // 如果需要，可以使用辅助 map 进行快速查找
    FieldMap map[string]*FieldSymbol 
}
```

### 3.3 Meta API 增强
运行宏的解释器 (`Interpreter`) 需要通过 `target` 对象暴露这些新信息。

**`target` 对象的新属性 (针对结构体):**
-   `target.fields`: 字段对象列表。
    -   `field.name` (string): 字段名称。
    -   `field.type` (string): 字段类型名称。
    -   `field.tag` (string): 原始 Tag 字符串（或为空）。
-   `target.methods`: 方法对象列表。

**宏用法示例:**
```javascript
@macro DeriveJson {
    let struct_name = target.name;
    let fields = target.fields;
    
    let json_obj = "";
    for (f : fields) {
        let key = f.name;
        // 简单的 Tag 解析（模拟）
        if (f.tag != "") {
            key = parse_tag(f.tag, "json"); 
        }
        json_obj += "\"" + key + "\": " + f.name + ",";
    }
    
    return #quote {
        fn to_json(self: *${struct_name}) string {
            return "{" + ${json_obj} + "}";
        }
    };
}
```

## 4. 实施计划 (Implementation Plan)

### 第一阶段：语法与 AST
1.  更新 `MyGo.g4` 以支持结构体 Tag。
2.  重新生成 ANTLR 词法分析器/解析器代码。
3.  通过基础解析测试验证 AST 结构。

### 第二阶段：核心符号表
1.  定义 `FieldSymbol` 结构。
2.  重构 `Symbol` 结构，使用切片 `[]FieldSymbol` 代替 map。
3.  更新 `DeclarationCollector`（语义分析），解析 Tag 并填充有序字段列表。

### 第三阶段：编译器适配
1.  重构 `transpiler.go` 以使用新的字段结构。
2.  确保生成的 Go 代码保留结构体 Tag（如果兼容，将 MyGo Tag 映射到 Go Tag）。
3.  修复 `semantic` 包中的类型检查逻辑以处理有序字段。

### 第四阶段：Meta API 与宏环境
1.  更新 `interpreter.go` 的 `createSymbolMeta` 函数。
2.  将 `[]FieldSymbol` 转换为 `MetaValue` 的 `ListValue`。
3.  实现 Tag 解析的辅助函数（可选，或留给用户层宏）。

## 5. 开发检查表 (打卡表)

此检查表用于跟踪 RFC-008 的实施进度。

### 🛠️ 第一阶段：语法与 AST
- [ ] 修改 `MyGo.g4`，在 `structField` 中添加 `(STRING)?`。
- [ ] 运行 `antlr4` 重新生成 Go 解析器文件。
- [ ] 创建测试文件 `tests/rfc008_syntax.mygo` 以验证 Tag 解析。

### 📦 第二阶段：符号表与语义分析
- [ ] 在 `pkg/compiler/symbols` 中定义 `FieldSymbol` 结构。
- [ ] 将 `Symbol.Fields` 从 map 重构为切片。
- [ ] 更新 `DeclarationCollector.VisitStructDecl` 以执行以下操作：
    - [ ] 按顺序收集字段。
    - [ ] 从 AST 中提取 Tag 字符串。
    - [ ] 存储在 `FieldSymbol` 中。

### ⚙️ 第三阶段：转译器与类型检查器
- [ ] 修复 `transpiler.go`: `VisitStructDecl` 遍历有序字段。
- [ ] 修复 `transpiler.go`: 确保生成的 Go 结构体字段包含 Tag。
- [ ] 修复 `semantic/checker.go`: 更新字段查找逻辑（map 访问 -> 切片搜索或辅助 map）。

### 🔮 第四阶段：Meta API 与反射
- [ ] 更新 `interpreter.go`: 在 `createSymbolMeta` 中添加 `fields` 属性。
- [ ] 实现 `FieldSymbol` 到 `MetaValue` 的转换。
- [ ] 在 `tests/rfc008_macro.mygo` 中验证宏对字段的访问。

### 📚 第五阶段：文档与验证
- [ ] 更新 `MyGo_Syntax_Reference.md`，添加结构体 Tag 语法。
- [ ] 编写 `RFC-008-Test-Report.md`。
- [ ] 使用新的反射 API 实现一个可运行的 `@Derive(Json)` 示例。
