# codetree

A CLI tool that builds repository tree visualizations with code signatures (functions, classes, methods, interfaces, structs, etc.).

![Go Version](https://img.shields.io/badge/Go-1.22%2B-00ADD8?style=flat&logo=go)
![License](https://img.shields.io/badge/License-MIT-green.svg)

## Features

- **Multi-language support**: Go, Python, JavaScript, TypeScript
- **Code signatures**: Extracts function/method signatures, struct fields, interface methods
- **Docstrings**: Captures documentation comments (godoc, docstrings, JSDoc)
- **Flexible filtering**: By language, entity type, depth, file extension
- **ASCII tree output**: Clean, readable directory structure with embedded code info

## Installation

### From source

```bash
go install github.com/lemantorus/codetree@latest
```

### Build locally

```bash
git clone https://github.com/lemantorus/codetree.git
cd codetree
go build -o codetree .
```

## Usage

```bash
codetree [directory] [flags]
```

### Basic Examples

```bash
# Show tree for current directory
codetree

# Show tree for specific directory
codetree ./myproject

# Only Go files
codetree . -l go

# Only Python files, hide docstrings
codetree . -l python --no-docstrings

# Only functions and methods
codetree . -t func,method

# Limit depth
codetree . -d 2

# Save to file
codetree . -o tree.txt
```

### Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--depth` | `-d` | -1 | Max search depth (-1 = unlimited) |
| `--lang` | `-l` | all | Languages to parse (comma-separated) |
| `--type` | `-t` | all | Entity types to show (comma-separated) |
| `--no-docstrings` | | false | Hide docstrings |
| `--no-signatures` | | false | Hide signatures (show type + name only) |
| `--output` | `-o` | stdout | Output file path |
| `--ext` | | auto | File extensions to include |
| `--include-libs` | | false | Include dependency directories |
| `--exclude-dirs` | | | Regex pattern for directories to skip |
| `--exclude-dirs-ignore-case` | | false | Case-insensitive exclude matching |
| `--max-signature` | | 0 | Max signature length (0 = unlimited) |
| `--max-docstring` | | 0 | Max docstring length (0 = unlimited) |

### Supported Languages

| Language | Extensions | Entities |
|----------|------------|----------|
| **Go** | `.go` | struct, interface, func, method, type, const |
| **Python** | `.py`, `.pyw`, `.pyi` | class, def (function/method) |
| **JavaScript** | `.js`, `.mjs`, `.cjs` | class, function, method |
| **TypeScript** | `.ts`, `.tsx`, `.mts`, `.cts` | class, interface, type, enum, function, method |

### Entity Types

Use `--type` / `-t` to filter:

- `func`, `function` - Functions
- `class` - Classes
- `method` - Methods
- `interface` - Interfaces
- `struct`, `type` - Structs and type aliases
- `enum` - Enums (TypeScript)
- `const`, `constant` - Constants

## Example Output

```
codetree/
├── cmd/
│   └── codetree/
│       └── main.go
│           ├── func main()
│           ├── func run(cmd *cobra.Command, args []string) error
│           ├── func parseEntityTypes(types []string) map[model.EntityType]bool
│           └── func validateLanguages(langs []string)
├── internal/
│   ├── model/
│   │   └── entity.go
│   │       ├── type EntityType int
│   │       ├── type CodeEntity struct { Name, Type, Signature, Docstring, LineStart, LineEnd, Children }
│   │       ├── type FileResult struct
│   │       └── type DirNode struct
│   ├── parser/
│   │   ├── go.go
│   │   │   ├── type GoParser struct
│   │   │   ├── func NewGoParser() *GoParser
│   │   │   ├── func (p *GoParser) Parse(content []byte) ([]model.CodeEntity, error)
│   │   │   └── ...
│   │   └── ...
│   └── ...
└── go.mod
```

## Project Structure

```
codetree/
├── main.go                    # CLI entry point
├── internal/
│   ├── model/entity.go        # Data models
│   ├── parser/
│   │   ├── parser.go          # Parser interface
│   │   ├── registry.go        # Parser registration
│   │   ├── go.go              # Go parser
│   │   ├── python.go          # Python parser
│   │   ├── javascript.go      # JavaScript parser
│   │   └── typescript.go      # TypeScript parser
│   ├── tree/builder.go        # Directory tree builder
│   └── output/formatter.go    # ASCII output formatter
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

## Adding a New Language Parser

1. Create `internal/parser/<language>.go`
2. Implement the `Parser` interface:

```go
type Parser interface {
    Language() string                    // e.g., "go", "python"
    Extensions() []string                // e.g., [".go"]
    LibDirs() []string                   // dependency dirs to skip
    Parse(content []byte) ([]model.CodeEntity, error)
}
```

3. Register in `init()`:

```go
func init() {
    Register(NewMyParser())
}
```

## Development

```bash
# Build
go build -o codetree ./cmd/codetree

# Run tests
go test ./...

# Format code
go fmt ./...

# Vet
go vet ./...
```

## License

[MIT](LICENSE)
