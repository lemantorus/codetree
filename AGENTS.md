# AGENTS.md

Coding agent guidelines for the codetree project.

## Project Overview

`codetree` is a Go CLI tool that builds repository tree visualizations with code signatures (functions, classes, methods, interfaces, etc.). It parses source files and outputs an ASCII tree with optional docstrings and full signatures.

Supported languages: Python, JavaScript, TypeScript (extensible via Parser interface).

## Build / Test / Lint Commands

```bash
# Build the binary
go build -o codetree ./cmd/codetree

# Run the tool
./codetree [directory] [flags]

# Run all tests
go test ./...

# Run tests for a specific package
go test ./internal/parser/...

# Run a single test by name
go test -run TestFunctionName ./internal/parser/...

# Run tests with verbose output
go test -v ./...

# Format code (always run before committing)
go fmt ./...

# Vet code for common errors
go vet ./...

# Update dependencies
go mod tidy

# Check for dependency updates
go list -u -m all
```

## Project Structure

```
codetree/
├── cmd/codetree/main.go    # CLI entry point, flag definitions
├── internal/
│   ├── model/
│   │   └── entity.go       # EntityType, CodeEntity, DirNode structs
│   ├── parser/
│   │   ├── parser.go       # Parser interface definition
│   │   ├── registry.go     # Parser registration and lookup
│   │   ├── python.go       # Python parser implementation
│   │   ├── javascript.go   # JavaScript parser implementation
│   │   └── typescript.go   # TypeScript/TSX parser implementation
│   ├── tree/
│   │   └── builder.go      # Directory tree builder, filtering logic
│   └── output/
│       └── formatter.go    # ASCII tree output formatter
├── go.mod
├── go.sum
└── .gitignore
```

## Code Style Guidelines

### Imports

Group imports in this order with blank lines between groups:
1. Standard library (fmt, os, strings, etc.)
2. Internal packages (codetree/internal/...)
3. External packages (github.com/...)

```go
import (
    "fmt"
    "os"
    "strings"

    "codetree/internal/model"
    "codetree/internal/parser"

    "github.com/spf13/cobra"
)
```

Use blank imports for parser registration side effects:
```go
import _ "codetree/internal/parser"
```

### Formatting

- Use tabs for indentation (Go standard)
- Run `go fmt ./...` before committing
- No trailing whitespace
- Keep lines under 100 characters when practical

### Naming Conventions

| Type | Convention | Example |
|------|------------|---------|
| Packages | lowercase, single word | `parser`, `model`, `tree` |
| Types | PascalCase | `CodeEntity`, `DirNode`, `PythonParser` |
| Interfaces | PascalCase + -er or noun | `Parser`, `BuilderOption` |
| Functions/Methods | PascalCase (exported), camelCase (unexported) | `Parse()`, `parseLines()` |
| Constants | PascalCase or camelCase | `EntityFunction`, `globalSkipDirs` |
| Private fields | camelCase | `showDocstrings`, `excludePattern` |
| CLI flags | kebab-case | `--no-docstrings`, `--include-libs` |

### Types

- Use `int` for line numbers, array indices
- Use `string` for paths, names, signatures
- Use `[]string` for lists (extensions, directories)
- Use `map[string]bool` for sets
- Prefer `*regexp.Regexp` for compiled patterns
- Define type aliases when appropriate:
  ```go
  type EntityType = model.EntityType
  type BuilderOption func(*Builder)
  ```

### Structs

Group related fields together. Order: exported fields first, then unexported.

```go
type Builder struct {
    // Configuration
    languages   map[string]bool
    entityTypes map[EntityType]bool
    maxDepth    int
    
    // Output options
    showDocstr  bool
    extensions  map[string]bool
    
    // Filtering
    includeLibs    bool
    excludePattern *regexp.Regexp
}
```

### Error Handling

- Return errors from functions, don't panic in library code
- Wrap errors with context using `fmt.Errorf`:
  ```go
  return fmt.Errorf("failed to build tree: %w", err)
  ```
- In CLI code, return errors from `RunE` functions
- Use `os.Exit(1)` only in `main()` or command execution failure
- Silently continue on non-critical errors (file read failures during tree walk)

### Function Patterns

Use functional options for complex constructors:
```go
type BuilderOption func(*Builder)

func WithMaxDepth(depth int) BuilderOption {
    return func(b *Builder) {
        b.maxDepth = depth
    }
}

func NewBuilder(opts ...BuilderOption) *Builder {
    b := &Builder{
        maxDepth: -1,  // defaults
    }
    for _, opt := range opts {
        opt(b)
    }
    return b
}
```

### Parser Interface

All language parsers must implement this interface:
```go
type Parser interface {
    Language() string      // e.g., "python", "javascript"
    Extensions() []string  // e.g., [".py", ".pyw"]
    LibDirs() []string     // dependency directories to skip
    Parse(content []byte) ([]model.CodeEntity, error)
}
```

Register parsers in `init()`:
```go
func init() {
    Register(NewPythonParser())
}
```

### Adding a New Language Parser

1. Create `internal/parser/<language>.go`
2. Implement the `Parser` interface
3. Register in `init()` function
4. Define `LibDirs()` with language-specific dependency directories
5. Use regex-based parsing for simplicity (no external AST dependencies)

Example skeleton:
```go
package parser

import (
    "regexp"
    "strings"
    "codetree/internal/model"
)

type GoParser struct {
    funcRegex  *regexp.Regexp
    structRegex *regexp.Regexp
}

func NewGoParser() *GoParser {
    return &GoParser{
        funcRegex: regexp.MustCompile(`(?m)^func\s+(\w+)\s*\(([^)]*)\)`),
    }
}

func (p *GoParser) Language() string { return "go" }
func (p *GoParser) Extensions() []string { return []string{".go"} }
func (p *GoParser) LibDirs() []string { return []string{"vendor"} }

func (p *GoParser) Parse(content []byte) ([]model.CodeEntity, error) {
    // Implementation
}

func init() { Register(NewGoParser()) }
```

## CLI Flags Reference

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-d, --depth` | int | -1 | Max search depth (-1 = unlimited) |
| `-l, --lang` | []string | all | Languages to parse |
| `-t, --type` | []string | all | Entity types (func,class,method,etc.) |
| `--no-docstrings` | bool | false | Hide docstrings |
| `--no-signatures` | bool | false | Hide signatures (show type + name only) |
| `-o, --output` | string | stdout | Output file path |
| `--ext` | []string | auto | File extensions |
| `--include-libs` | bool | false | Include dependency directories |
| `--exclude-dirs` | string | "" | Regex pattern for dirs to skip |
| `--exclude-dirs-ignore-case` | bool | false | Case-insensitive exclude matching |

## Testing Guidelines

- Place tests in the same package with `_test.go` suffix
- Use table-driven tests for multiple cases:
  ```go
  func TestParse(t *testing.T) {
      tests := []struct {
          name     string
          input    string
          expected []model.CodeEntity
      }{
          {"simple function", "def foo(): pass", ...},
      }
      for _, tt := range tests {
          t.Run(tt.name, func(t *testing.T) {
              // test logic
          })
      }
  }
  ```

## Commit Guidelines

- Use conventional commit format: `type: description`
- Types: `feat`, `fix`, `docs`, `refactor`, `test`, `chore`
- Keep commits focused and atomic
- Run `go fmt ./...` and `go vet ./...` before committing
