# RFC-008: Static Reflection & Enhanced Metaprogramming

- **Status**: Draft
- **Author**: Trae AI
- **Created**: 2026-03-11
- **Target Version**: MyGo v0.3.0

## 1. Abstract
This RFC proposes a comprehensive **Static Reflection (Compile-Time Reflection)** mechanism for MyGo. It aims to expose detailed type information—such as struct fields, tags, and method signatures—to the macro system during compilation. This will enable powerful code generation capabilities (e.g., serialization, ORM mapping, dependency injection) without runtime overhead, aligning MyGo with modern systems programming languages like Rust (procedural macros) and Zig (comptime).

## 2. Motivation
Currently (RFC-007 implementation), MyGo's metaprogramming capabilities are limited to:
1.  **Symbol Discovery**: Finding symbols annotated with specific attributes (`find_all_annotated_with`).
2.  **Basic Metadata**: Accessing symbol name, package, and kind.
3.  **Raw Body Access**: Retrieving the source code of a function body as a string.

**Limitations:**
-   **No Field Introspection**: Macros cannot iterate over the fields of a struct. It is impossible to generate code that depends on the structure of a type (e.g., `toString()`, `toJson()`).
-   **No Metadata Tags**: Struct fields cannot carry metadata (like Go's struct tags `json:"id"` or Rust's attributes `#[serde(rename = "id")]`).
-   **Unordered Fields**: The current compiler symbol table stores fields in a `map[string]string`, which loses the original declaration order—critical for binary serialization or C-interop.

To support features like `@Derive(Json)` or an ORM framework, we need a way to inspect types at compile-time.

## 3. Design Proposal

### 3.1 Grammar Extensions (Struct Tags)
We propose extending the grammar to allow an optional string literal after a field declaration in a struct.

**Proposed EBNF Change:**
```ebnf
// Current
structField: ID ':' typeType ;

// Proposed
structField: ID ':' typeType (STRING)? ;
```

**Example Usage:**
```mygo
struct User {
    id: int "json:\"id\" db:\"primary_key\"",
    name: string "json:\"name\"",
    age: int // Optional tag
}
```

### 3.2 Symbol Table Refactoring
The internal representation of `Struct` symbols must be updated to preserve field order and store tags.

**Current (`pkg/compiler/symbols/symbol.go`):**
```go
type Symbol struct {
    // ...
    Fields map[string]string // Name -> Type
}
```

**Proposed:**
```go
type FieldSymbol struct {
    Name string
    Type string
    Tag  string // Raw tag string, e.g., "json:\"id\""
}

type Symbol struct {
    // ...
    Fields []FieldSymbol // Ordered list of fields
    // Helper map for fast lookups if needed
    FieldMap map[string]*FieldSymbol 
}
```

### 3.3 Meta API Enhancements
The `Interpreter` (which runs macros) needs to expose this new information via the `target` object.

**New Properties on `target` (for Structs):**
-   `target.fields`: A list of field objects.
    -   `field.name` (string): The name of the field.
    -   `field.type` (string): The type name of the field.
    -   `field.tag` (string): The raw tag string (or empty).
-   `target.methods`: A list of method objects.

**Example Macro Usage:**
```javascript
@macro DeriveJson {
    let struct_name = target.name;
    let fields = target.fields;
    
    let json_obj = "";
    for (f : fields) {
        let key = f.name;
        // Simple tag parsing (mockup)
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

## 4. Implementation Plan

### Phase 1: Grammar & AST
1.  Update `MyGo.g4` to support struct tags.
2.  Regenerate ANTLR lexer/parser code.
3.  Verify AST structure with basic parsing tests.

### Phase 2: Core Symbol Table
1.  Define `FieldSymbol` structure.
2.  Refactor `Symbol` struct to use slice `[]FieldSymbol` instead of map.
3.  Update `DeclarationCollector` (semantic analysis) to parse tags and populate the ordered field list.

### Phase 3: Compiler Adaptation
1.  Refactor `transpiler.go` to use the new field structure.
2.  Ensure generated Go code preserves struct tags (mapping MyGo tags to Go tags if compatible).
3.  Fix type checking logic in `semantic` package to handle ordered fields.

### Phase 4: Meta API & Macro Environment
1.  Update `interpreter.go`'s `createSymbolMeta` function.
2.  Convert `[]FieldSymbol` to `ListValue` of `MetaValue`s.
3.  Implement helper functions for tag parsing (optional, or leave to userland macros).

## 5. Development Checklist (Punch Card)

This checklist tracks the progress of RFC-008 implementation.

### 🛠️ Phase 1: Grammar & AST
- [ ] Modify `MyGo.g4` to add `(STRING)?` to `structField`.
- [ ] Run `antlr4` to regenerate Go parser files.
- [ ] Create test file `tests/rfc008_syntax.mygo` to verify tag parsing.

### 📦 Phase 2: Symbol Table & Semantic Analysis
- [ ] Define `FieldSymbol` struct in `pkg/compiler/symbols`.
- [ ] Refactor `Symbol.Fields` from map to slice.
- [ ] Update `DeclarationCollector.VisitStructDecl` to:
    - [ ] Collect fields in order.
    - [ ] Extract tag string from AST.
    - [ ] Store in `FieldSymbol`.

### ⚙️ Phase 3: Transpiler & Type Checker
- [ ] Fix `transpiler.go`: `VisitStructDecl` to iterate over ordered fields.
- [ ] Fix `transpiler.go`: Ensure tags are included in generated Go struct fields.
- [ ] Fix `semantic/checker.go`: Update field lookup logic (map access -> slice search or auxiliary map).

### 🔮 Phase 4: Meta API & Reflection
- [ ] Update `interpreter.go`: Add `fields` property to `createSymbolMeta`.
- [ ] Implement `MetaValue` conversion for `FieldSymbol`.
- [ ] Verify macro access to fields in `tests/rfc008_macro.mygo`.

### 📚 Phase 5: Documentation & Validation
- [ ] Update `MyGo_Syntax_Reference.md` with Struct Tag syntax.
- [ ] Write `RFC-008-Test-Report.md`.
- [ ] Implement a working `@Derive(Json)` example using the new reflection API.
