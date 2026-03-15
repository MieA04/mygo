# MyGo Package Management & Visibility

MyGo adopts a modular design, using Packages to organize code. It is highly compatible with Go's package system but provides finer-grained, explicit visibility control.

##### 1. Package Declaration

Every source file must belong to a package. The package declaration is located on the first line of the file (excluding comments).

#### Under the Hood: Module System
*   **Go Modules Integration**: MyGo fully integrates with Go Modules. A `go.mod` file at the project root defines the module path.
*   **Directory-Based**: Like Go, all files in a directory must belong to the same package.
*   **Compilation**: The compiler resolves imports relative to `GOPATH` or `go.mod`, ensuring compatibility with existing Go libraries.

### Syntax Definition
```antlr
packageDecl: 'package' ID ';'? ;
```

### Example
```mygo
package main

// Or define a library package
package math_utils
```

- `package main`: Defines the entry package for an executable program, which must contain a `main` function.
- Other package names: Usually consistent with the directory name, used for library code.

---

## 2. Import

Use the `import` keyword to import dependency packages. MyGo supports single-line and block imports.

### Syntax Definition
```antlr
importStmt
    : 'import' '{' importSpec (',' importSpec)* ','? '}' ';'? # BlockImport
    | 'import' importSpec ';'?                                # SingleImport
    ;
importSpec: STRING ('as' ID)? ;
```

### Example
```mygo
// Single line import
import "fmt";

// Block import
import {
    "net/http",
    "os" as std_os // Aliased import
}
```

---

## 3. Visibility Modifiers

MyGo discards Go's implicit rule of "capitalization determines visibility" in favor of explicit keyword modifiers. This makes code intent clearer and unrestricted by naming styles.

### Syntax Definition
```antlr
modifier: 'pub' | 'pkg' | 'pri' ;
```

### 3.1 Level Description

| Keyword | Level | Description | Scope |
| :--- | :--- | :--- | :--- |
| **pub** | Public | Publicly visible | Accessible by any code importing this package |
| **pkg** | Package | Package visible | Accessible by all files within the same package (default level) |
| **pri** | Private | File private | Visible only within the current source file |

#### Under the Hood: Symbol Mangling
To implement these visibility levels on top of Go's simpler model (Capitalized=Public, Lowercase=Private), MyGo uses name mangling during transpilation:
*   `pub fn add` -> `func Add` (Exported)
*   `pkg fn helper` -> `func helper` (Package-private)
*   `pri fn internal` -> `func internal_Hash123` (File-private, suffixed with file hash to prevent collision)

### 3.2 Detailed Example

Suppose we have a math library package `math_lib`.

```mygo
package math_lib

// pub: Struct exposed to the public
pub struct Vector {
    // pub: Field exposed to the public
    pub X: float,
    pub Y: float,
    
    // pkg: Field visible only within the package (default, pkg if no modifier)
    cachedLength: float,
    
    // pri: Private field visible only in the current file
    pri secretKey: string
}

// pub: Function exposed to the public
pub fn NewVector(x: float, y: float): Vector {
    return Vector{X: x, Y: y, cachedLength: 0.0, secretKey: "hidden"};
}
```

### 3.3 Transpilation Mapping

When MyGo code is transpiled to Go code, the compiler automatically handles naming conventions to comply with Go's visibility rules:

- `pub` symbols -> Transpiled to **Capitalized** initial (e.g., `NewVector`).
- `pkg` / `pri` symbols -> Transpiled to **Lowercase** initial (e.g., `newVector` or `newVector_suffix`).

This design allows you to freely use lowercase public functions in MyGo (like `pub fn add()`), and the compiler is responsible for converting it to `Add()` which Go can export.

---

## 4. Mixed Compilation

MyGo aims to integrate seamlessly with the existing Go ecosystem.

1.  **Direct Import of Go Packages**: You can directly `import "encoding/json"` in MyGo, and the compiler will automatically recognize the standard library.
2.  **Project Coexistence**: MyGo projects can contain `.go` files. The MyGo compiler handles these files during the build process, allowing MyGo code to call Go code and vice versa (subject to visibility rules).
3.  **Go Modules**: MyGo reuses `go.mod` and `go.sum` for dependency management without needing extra package management tools.

```bash
# Initialize project
go mod init myproject

# Build MyGo project
mygo build
```
