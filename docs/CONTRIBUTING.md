# Contributing to qrcode

Thank you for your interest in contributing to the `qrcode` library! This guide
covers everything you need to know to set up a development environment and submit
contributions.

## Table of Contents

- [Development Environment](#development-environment)
- [Project Structure](#project-structure)
- [Running Tests](#running-tests)
- [Linting and Formatting](#linting-and-formatting)
- [Code Style Guidelines](#code-style-guidelines)
- [Submitting Changes](#submitting-changes)
- [Pull Request Process](#pull-request-process)
- [Adding a New Payload Type](#adding-a-new-payload-type)
- [Adding a New Renderer](#adding-a-new-renderer)
- [Reporting Issues](#reporting-issues)

## Development Environment

### Prerequisites

- **Go 1.23.0** or later
- **golangci-lint** (latest) — for linting

### Setup

```bash
# Clone the repository
git clone https://github.com/os-gomod/qrcode.git
cd qrcode

# Install dependencies (no external dependencies — pure stdlib)
go mod download

# Verify the build compiles
go build ./...

# Run the example
cd example && go build -o qr-example && ./qr-example
```

## Project Structure

```
qrcode/
├── qrcode.go           # Client interface, constructors, Format/ECLevel types
├── generator.go        # Client implementation
├── builder.go          # Fluent Builder API
├── config.go           # Configuration struct and validation
├── options.go          # Functional options
├── helpers.go          # Quick* convenience functions
├── encoding/           # QR matrix encoding (Galois fields, masking, versioning)
├── payload/            # All 35 payload type implementations
├── renderer/           # Output renderers (PNG, SVG, Terminal, PDF, Base64)
├── batch/              # Batch processing with worker pool
├── logo/               # Logo loading, resizing, tinting, and overlay
├── errors/             # Structured error types and codes
├── internal/           # Non-public infrastructure (not importable externally)
│   ├── workerpool/     # Generic bounded goroutine pool
│   ├── storage/        # File I/O abstraction
│   ├── pool/           # Buffer pooling (sync.Pool wrapper)
│   ├── lifecycle/      # Close guard / lifecycle management
│   ├── hash/           # FNV-1a hashing
│   └── singleflight/   # In-flight request deduplication
├── testing/            # Test utilities and contract helpers
├── example/            # Comprehensive example program
├── docs/               # Documentation (ADR, migration guide)
└── benchmark_test.go   # Performance benchmarks
```

## Running Tests

```bash
# Run all tests with race detection
go test -race ./...

# Run all tests with coverage
go test -race -cover ./...

# Run a specific package
go test -race ./payload/...
go test -race ./renderer/...
go test -race ./encoding/...

# Run tests with verbose output
go test -race -v ./...

# Run a specific test
go test -race -run TestWiFiPayload ./payload/...
```

### Using Make

```bash
make test        # Run tests with race detection
make coverage    # Generate coverage report (coverage.html)
make bench       # Run benchmarks
make all         # fmt + vet + lint + test
make ci          # lint + vet + test + coverage
```

## Linting and Formatting

```bash
# Format code
go fmt ./...

# Run go vet
go vet ./...

# Run golangci-lint
golangci-lint run ./...
```

### Using Make

```bash
make fmt         # Format code
make vet         # Run go vet
make lint        # Run golangci-lint
```

## Code Style Guidelines

Follow the standard Go conventions as described in [Effective Go](https://go.dev/doc/effective_go).

### General Rules

1. **Run `go fmt`** before every commit — the project uses standard Go formatting.
2. **Keep exported names descriptive** — prefer `WithDefaultSize` over `WithSize`.
3. **Document all exported types and functions** — every `//` comment should start with the symbol name.
4. **Accept interfaces, return structs** — follow the Go convention.
5. **Use `context.Context` as the first parameter** on all public methods.
6. **Return errors, don't panic** — `Must*` functions are the only exception.

### Error Handling

```go
// DO — return errors
func New(opts ...Option) (Client, error) { ... }

// DON'T — panic in library code (except Must* variants)
func MustNew(opts ...Option) Client { ... }
```

### Adding Options

Use the functional options pattern:

```go
func WithMyNewOption(value int) Option {
    return func(c *Config) {
        c.MyNewField = value
    }
}
```

### Testing

- Write table-driven tests where possible.
- Use `t.Helper()` in test helpers.
- Use `t.Errorf` / `t.Fatalf` (not `log.Fatal`).
- Test error paths as well as happy paths.
- Use `testing.T` context for subtests: `t.Run("name", func(t *testing.T) { ... })`.

## Submitting Changes

### Commit Messages

Use [Conventional Commits](https://www.conventionalcommits.org/):

```
feat(payload): add BitcoinPayload for BTC addresses
fix(renderer): correct PNG color encoding for dark backgrounds
docs: update README with new payload types
refactor(internal): consolidate worker pool implementation
test(encoding): add edge case tests for version 40 QR codes
chore: update golangci-lint configuration
```

### Branch Naming

```
feat/add-bitcoin-payload
fix/png-color-encoding
docs/update-readme
```

## Pull Request Process

1. **Create a feature branch** from `main`.
2. **Make your changes** — ensure `go fmt`, `go vet`, and `golangci-lint run` all pass.
3. **Write tests** — all new code should have corresponding tests.
4. **Run the full suite**: `make all` or `go test -race -cover ./...`.
5. **Update documentation** — if you add a payload type or renderer, update the README.
6. **Open a PR** with a clear description of the change and motivation.
7. **Address review feedback** — be responsive to maintainer comments.

### PR Checklist

- [ ] `go build ./...` passes
- [ ] `go test -race -cover ./...` passes (all tests green)
- [ ] `golangci-lint run ./...` passes (no warnings)
- [ ] `go vet ./...` passes
- [ ] New code has test coverage
- [ ] Documentation updated (README, godoc comments)
- [ ] Commit messages follow Conventional Commits

## Adding a New Payload Type

1. Create a new file in `payload/` (e.g., `payload/mytype.go`).
2. Define a struct implementing the `Payload` interface:
   ```go
   type MyPayload struct {
       Field string
   }

   func (p *MyPayload) Encode() (string, error) { ... }
   func (p *MyPayload) Validate() error { ... }
   func (p *MyPayload) Type() string { return "mytype" }
   func (p *MyPayload) Size() int { return len(p.Field) }
   ```
3. Add tests in `payload/mytype_test.go`.
4. Update the README payload table.
5. Run `go test -race ./payload/...` to verify.

## Adding a New Renderer

1. Create a new file in `renderer/` (e.g., `renderer/myformat.go`).
2. Implement the `Renderer` interface:
   ```go
   type MyRenderer struct{}

   func (r *MyRenderer) Render(ctx context.Context, qr *encoding.QRCode, w io.Writer, opts ...RenderOption) error { ... }
   func (r *MyRenderer) ContentType() string { return "image/myformat" }
   ```
3. Register the renderer in the package `init()` function.
4. Add a new `Format` constant in `qrcode.go`.
5. Update the file extension mapping in `helpers.go`.
6. Add tests in `renderer/myformat_test.go`.
7. Update the README output formats section.

## Reporting Issues

When reporting bugs, please include:

1. **Go version**: `go version`
2. **Library version**: `git describe --tags` or commit hash
3. **Minimal reproduction code** — a short Go program that demonstrates the issue
4. **Expected vs. actual behavior**
5. **Error messages** (if any)

Feature requests are welcome! Please describe the use case and expected API.
