# goc — Mini Self-Hosted Go Compiler

A complete mini Go-to-C transpiling compiler written in Go.

```
Source .go  →  Lexer  →  Parser  →  AST  →  Codegen  →  C source  →  gcc  →  Binary
```

## Operating GoLet

> [!NOTE]
> All commands should be run from within the `GoLet` directory.

### 1. Initialize Go Environment (Optional)
The project comes with a `go.mod` file. If you need to re-initialize or change the module name:
```bash
go mod init GoLet
```

### 2. Build the Compiler
Requires Go 1.18+. This creates the `goc` executable in the current directory.
```bash
go build -o goc .
```

### 2. Compile a Go Program
Compile a single Go source file into a binary.
```bash
./goc -o hello tests/test1.go
```

### 3. Run the Resulting Binary
```bash
./hello
```

### Command Line Flags
| Flag | Description |
|------|-------------|
| `-o <file>` | Output binary name (default: `a.out`) |
| `-S` | Emit C source only (generated as `<output>.c`) |
| `-ast` | Print the Abstract Syntax Tree and exit |
| `-lex` | Print tokens and exit |
| `-v` | Verbose mode: print stage timings and dump C source on failure |

### Type System Notes
- **`int`**: Maps to C's `long long` (64-bit). When using `fmt.Printf`, use `%lld` for integer values.
- **`string`**: Maps to `const char*`.
- **`bool`**: Maps to `int` (0 for false, 1 for true).

## Testing
Run the comprehensive test suite to verify language features:
```bash
./goc -o test1 tests/test1.go && ./test1
```

## Architecture

| File | Role |
|------|------|
| `lexer.go` | Hand-written lexer with Go auto-semicolon insertion |
| `ast.go` | Full AST node hierarchy + type system |
| `parser.go` | Recursive-descent parser with Pratt precedence climbing |
| `codegen.go` | C code generator with scope/symbol table |
| `main.go` | CLI driver |

## Supported Language Features

### Types
- `int` (maps to `long long`)
- `string` (maps to `const char*`)
- `bool` (maps to `int`)

### Declarations
```go
var x int           // zero-initialized
var s string = "hi"
n := 42             // short variable declaration
a, b := 1, 2        // multi-assignment
```

### Control Flow
```go
if x > 0 { ... } else if x < 0 { ... } else { ... }
for { ... }                    // infinite loop
for i < 10 { ... }             // while-style
for i := 0; i < 10; i++ { ... } // C-style
break
continue
```

### Functions
```go
func add(a int, b int) int { return a + b }
func greet(name string) string { return "Hello, " + name }
func main() { ... }
```

### Operators
```
+  -  *  /  %
== != < > <= >=
&& || !
++ --
```

### Standard Library
```go
fmt.Println(...)     // prints with newline
fmt.Print(...)       // prints without newline
fmt.Printf(fmt, ...) // formatted print
fmt.Sprintf(fmt, ...)// returns formatted string
```

## Example

Input (`hello.go`):
```go
package main

import "fmt"

func factorial(n int) int {
    if n <= 1 {
        return 1
    }
    return n * factorial(n-1)
}

func main() {
    for i := 1; i <= 10; i++ {
        fmt.Printf("%d! = %d\n", i, factorial(i))
    }
}
```

Generated C (abbreviated):
```c
long long factorial(long long n) {
  if ((n <= 1LL)) {
    return 1LL;
  }
  return (n * factorial((n - 1LL)));
}

int main(void) {
  for (long long i = 1LL; (i <= 10LL); i++) {
    printf("%d! = %d\n", i, factorial(i));
  }
  return 0;
}
```

## Limitations

- Single file compilation only
- No arrays, slices, or maps
- No structs or methods
- No multiple return values
- No goroutines or channels
- No interfaces
- String operations limited to `+` concatenation
