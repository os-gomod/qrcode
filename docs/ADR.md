# Architecture Decision Records (ADR)

This document records key architectural decisions made during the qrcode v2 design and refactor. Each ADR follows the standard format: **Context → Decision → Consequences**.

---

## ADR-001: Client Interface as Public Contract

**Status**: Accepted

**Context**: The v1 codebase lacked a formal `Generator` interface. Consumers depended on the concrete `generator` struct directly, which made it difficult to test, mock, or swap implementations. Any internal refactoring risked breaking downstream code.

**Decision**: Define a stable `Client` interface as the primary public contract in `qrcode.go`. The interface exposes all public operations: `Generate`, `Render`, `Save`, `Batch`, `Close`, and more. The former `Generator` is preserved as a type alias (`= Client`) for backward compatibility.

**Consequences**:
- All implementations satisfy `Client` at compile time.
- Consumers can use interface-based dependency injection for testing.
- No breaking changes for code referencing `Generator` (it is a type alias).
- New code should prefer the canonical `Client` name.

---

## ADR-002: Payload Interface Design

**Status**: Accepted

**Context**: v1 used loose typing — payloads were strings passed to encoding functions with no validation. This led to runtime errors when invalid data was encoded, and no way to enumerate supported payload types.

**Decision**: Define a `Payload` interface with four methods:

```go
type Payload interface {
    Encode() (string, error)  // Returns the encoded string for QR encoding
    Validate() error          // Validates payload fields before encoding
    Type() string             // Returns the payload type name (e.g., "wifi", "url")
    Size() int                // Returns the encoded data size for version selection
}
```

Each payload type (TextPayload, WiFiPayload, VCardPayload, etc.) implements this interface. Validation runs before encoding, catching errors early.

**Consequences**:
- All 35 payload types share a uniform contract.
- Validation errors are caught before QR matrix encoding begins.
- The `Type()` method enables structured logging and batch result tracking.
- The `Size()` method enables accurate QR version auto-selection.

---

## ADR-003: Storage Abstraction

**Status**: Accepted

**Context**: v1 used `os.WriteFile` directly in multiple locations throughout the codebase. This made testing impossible without writing temporary files, and prevented supporting alternative storage backends (S3, GCS, in-memory, etc.).

**Decision**: Define a `Storage` interface in `internal/storage/` with `Save()` and `Read()` methods. The default implementation uses the local filesystem. The storage backend is injectable at construction time via options.

**Consequences**:
- File I/O is injectable — tests can use mock storage.
- Future cloud storage backends (S3, GCS, Azure Blob) require only implementing one interface.
- The `internal/storage` package is not importable by external consumers, preserving API stability.

---

## ADR-004: Worker Pool for Batch Processing

**Status**: Accepted

**Context**: v1's `Batch()` method launched unlimited goroutines (one per payload). At scale (thousands of QR codes), this caused goroutine leaks, memory exhaustion, and scheduler thrashing.

**Decision**: Implement a generic `workerpool.Pool[T, R]` in `internal/workerpool/` with:
- Configurable worker count (1–64, default 4)
- Bounded internal queue (default 1024)
- Context-based cancellation
- Optional per-task duration tracking

```go
pool := workerpool.New[int, *encoding.QRCode](workerCount)
results := pool.Run(ctx, tasks, encodeFunc)
```

**Consequences**:
- All concurrent work uses bounded goroutines — no more unbounded launches.
- The pool is generic and reusable across batch, generation, and future concurrent operations.
- Context cancellation propagates immediately to all workers.
- Memory usage is predictable regardless of batch size.

---

## ADR-005: Singleflight Deduplication for Generate

**Status**: Accepted

**Context**: Under concurrent load, multiple goroutines requesting QR codes for the same payload data would perform redundant encoding operations. For example, a web server handling 100 requests for the same URL QR code would encode the QR matrix 100 times.

**Decision**: Use the `internal/singleflight` package to deduplicate concurrent `Generate` calls for identical payloads. The cache key is derived from the FNV-1a hash of the encoded payload data.

**Consequences**:
- Encoding the same payload from 100 concurrent goroutines results in only 1 actual encode operation.
- Subsequent callers block until the first encode completes, then share the result.
- No persistent cache — deduplication is per in-flight request only (no memory leak risk).
- The `internal/singleflight` package follows the same pattern as `golang.org/x/sync/singleflight`.

---

## ADR-006: Functional Options Pattern for Configuration

**Status**: Accepted

**Context**: v1 used a `Config` struct with public fields, requiring callers to construct and mutate it directly. This led to invalid configurations propagating silently (e.g., `MinVersion > MaxVersion`), and made it impossible to add validation at construction time.

**Decision**: Adopt the functional options pattern (common in Go) where each configuration knob is an `Option` function:

```go
type Option func(*Config)

func WithDefaultSize(size int) Option {
    return func(c *Config) { c.DefaultSize = size }
}
```

Construction validates the config before returning:

```go
func New(opts ...Option) (Client, error) {
    cfg := defaultConfig()
    for _, opt := range opts { opt(cfg) }
    if err := cfg.Validate(); err != nil { return nil, err }
    return newGenerator(cfg)
}
```

**Consequences**:
- Configuration is self-documenting — each option has a clear name and purpose.
- Invalid configurations are caught at construction time, not at render time.
- Adding new options requires zero changes to existing call sites.
- The `Config` struct remains internal — consumers interact only with `Option` functions.
- Both `New()` and `SetOptions()` validate before accepting.

---

## ADR-007: Renderer Registry Pattern

**Status**: Accepted

**Context**: v1 dispatched renderers via a large `switch` statement in `generator.go`. Adding a new output format required modifying core generation code, violating the Open/Closed Principle.

**Decision**: Define a `Renderer` interface with `Render()` and `ContentType()` methods. A package-level registry maps `Format → Renderer`, enabling zero-modification extensibility:

```go
type Renderer interface {
    Render(ctx context.Context, qr *encoding.QRCode, w io.Writer, opts ...RenderOption) error
    ContentType() string
}
```

Registering a new format:

```go
func init() {
    RegisterRenderer(FormatPNG, &PNGRenderer{})
}
```

**Consequences**:
- Adding a new format requires only implementing the `Renderer` interface and calling `RegisterRenderer`.
- The type-switch dispatch is eliminated — replaced by a map lookup.
- Each renderer is independently testable.
- The registry is protected by `sync.Once` for thread safety.

---

## ADR-008: Buffer Pool for PNG Rendering

**Status**: Accepted

**Context**: PNG rendering allocates large byte buffers (potentially hundreds of KB per QR code). Under high throughput (e.g., batch processing thousands of QR codes), this creates significant GC pressure and allocation churn.

**Decision**: Use `internal/pool.BufferPool` (wrapping `sync.Pool`) for PNG and Base64 rendering paths. SVG, PDF, and Terminal renderers write directly to the provided `io.Writer` and do not need pooling.

**Consequences**:
- High-throughput PNG rendering reuses buffers, reducing allocations by ~80% per render call.
- `sync.Pool` handles buffer lifecycle automatically — buffers are returned to the pool when garbage collected.
- No pool overhead for SVG/PDF/Terminal — they stream directly to the writer.

---

## ADR-009: Context Propagation Everywhere

**Status**: Accepted

**Context**: v1 hardcoded `context.Background()` in 17 places throughout the codebase. This prevented callers from canceling long-running operations, setting timeouts, or propagating tracing/observability context.

**Decision**: All public methods accept `context.Context` as the first parameter. No `context.Background()` calls exist in library code — the caller always controls context.

**Consequences**:
- Users can cancel long-running batch operations: `ctx, cancel := context.WithTimeout(...)`
- Timeout support is free via `context.WithTimeout`.
- Observability context (traces, metrics) propagates through the entire call chain.
- The one exception is `Close()`, which does not need context (it is instantaneous).

---

## ADR-010: Backward-Compatible Type Aliases

**Status**: Accepted

**Context**: Renaming `Generator` to `Client` and `ErrorCorrectionLevel` to `ECLevel` would break all downstream code. The Go ecosystem convention (e.g., `io.Reader` renames) favors backward compatibility during transitions.

**Decision**: Preserve the old names as type aliases:

```go
// Generator is a type alias for Client, maintained for backward compatibility.
// Deprecated: Use Client instead.
type Generator = Client

// ErrorCorrectionLevel is a type alias for ECLevel.
// Deprecated: Use ECLevel instead.
type ErrorCorrectionLevel = ECLevel
```

**Consequences**:
- Existing v1 code compiles against v2 without changes.
- New code uses canonical names (`Client`, `ECLevel`, `NewClient`).
- Deprecation comments guide migration via `godoc`.
- Type aliases have zero runtime overhead — they are resolved at compile time.

---

## ADR-011: Pipeline Consolidation

**Status**: Accepted

**Context**: v1 had approximately 505 lines of code duplication across 8+ generation functions and 3 separate batch worker loops. Each path (Quick, Builder, Render, Save, Batch) had its own copy of the encode-and-render pipeline.

**Decision**: Consolidate all generation paths into a single `generate()` → `renderToWriter()` pipeline. Consolidate batch worker loops into a single `process()` method delegating to `workerpool.Pool`.

**Consequences**:
- All generation paths (Quick, Builder, Render, Save) funnel through a single code path.
- Bug fixes and performance improvements need to touch only one location.
- The total code reduction from deduplication was significant.
- The pipeline is easier to reason about and test.

---

## ADR-012: Logo Rendering as Post-Processing

**Status**: Accepted

**Context**: v1's logo overlay logic was tightly coupled with PNG rendering in the generator. This made it impossible to apply logos to SVG output, test logo processing independently, or reuse the logo logic for other image operations.

**Decision**: Logo processing lives in a separate `logo/` package with composable functions:

- `logo.New(path, ratio)` — Create a logo processor
- `logo.Load()` — Load and decode the logo image
- `logo.ResizeLogo(img, qrSize, ratio)` — Resize to fit the QR code
- `logo.TintLogo(img, color)` — Apply a tint color
- `logo.OverlayLogo(qrImg, logoImg, padding)` — Composite onto the QR image
- `logo.Validate(path)` — Verify the logo file exists and is a supported format

**Consequences**:
- Logo processing is format-agnostic — works with any `image.Image`.
- The `logo/` package has no dependency on the QR encoding engine.
- Logo processing is independently testable.
- Future formats (SVG logos, for example) can be supported by extending the package.

---

## ADR-013: Internal Package Isolation

**Status**: Accepted

**Context**: The v1 codebase had 12+ packages with circular dependency risks and no clear boundaries between public API and implementation details.

**Decision**: Introduce `internal/` packages for all non-public infrastructure:

| Package | Purpose |
|---------|---------|
| `internal/workerpool` | Generic bounded goroutine pool |
| `internal/storage` | File I/O abstraction |
| `internal/pool` | Buffer pooling (sync.Pool wrapper) |
| `internal/lifecycle` | Close guard / lifecycle management |
| `internal/hash` | FNV-1a hashing for deduplication |
| `internal/singleflight` | In-flight request deduplication |

**Consequences**:
- External consumers cannot import internal packages, enforcing API stability.
- Dependencies flow inward only — public packages never import internal.
- Implementation details can be refactored freely without breaking the public API.
- Go's `internal/` package mechanism provides compile-time enforcement.

---

## ADR-014: ConfigPatch with Pointer Fields for Zero-Value Safety

**Status**: Accepted

**Context**: The original `Config.Merge()` method used Go zero values to determine whether a field was "set" or "not set". This created ambiguity: a caller could not explicitly set `QuietZone` to `0` or `AutoSize` to `false` because those values were indistinguishable from "not provided". The merge logic also had subtle bugs around bool and float64 zero values.

**Decision**: Introduce `ConfigPatch` with pointer fields (`*int`, `*bool`, `*string`, `*float64`, `*time.Duration`). Only non-nil fields are applied via `ApplyPatch(base, patch)`, which returns a new `Config` without modifying the original. Pointer helper functions (`IntP`, `StringP`, `BoolP`, `Float64P`, `DurationP`) eliminate verbose `&value` expressions.

```go
patch := qrcode.ConfigPatch{
    WorkerCount: qrcode.IntP(8),
    QuietZone:   qrcode.IntP(0),  // explicitly set to 0 — no ambiguity
    AutoSize:    qrcode.BoolP(false), // explicitly disable
}
cfg := qrcode.ApplyPatch(defaultConfig(), patch)
```

`ValidatePatch()` checks only non-nil fields, preventing false validation errors for unset fields.

**Consequences**:
- Zero values can be explicitly applied without ambiguity.
- `Config.Merge()` is deprecated but preserved for backward compatibility.
- `ApplyPatch` returns a new `Config`, preventing accidental mutation.
- The `ConfigToPatch()` helper converts a full `Config` to a `ConfigPatch` for serialization.
- Pointer helpers (`IntP`, `BoolP`, etc.) reduce syntactic overhead.

---

## ADR-015: DomainError Interface for Structured Error Handling

**Status**: Accepted

**Context**: v1 used `fmt.Errorf` for all errors, providing no structured classification. Callers could not programmatically distinguish validation errors from encoding errors, determine retryability, or map errors to HTTP status codes. Error information was lost in string wrapping.

**Decision**: Introduce `DomainError` interface with three methods beyond `error`: `Code() ErrorCode`, `Retryable() bool`, and `Metadata() map[string]any`. The concrete `QRCodeError` type supports error chain wrapping via `Unwrap()`, metadata attachment via `WithMeta()`, and per-instance retryability overrides via `WithRetryable()`.

```go
type DomainError interface {
    error
    Code() ErrorCode
    Retryable() bool
    Metadata() map[string]any
}
```

Thirteen error codes cover all library error categories. Built-in HTTP status mapping (`VALIDATION → 400`, `TIMEOUT → 504`, etc.) enables seamless API server integration. `BatchError` aggregates per-item errors with index tracking.

**Consequences**:
- All library errors implement `DomainError`, enabling programmatic error handling.
- `errors.Is` and `errors.As` work correctly through the error chain.
- Sentinel errors (`ErrClosed`, `ErrDataTooLong`, `ErrInvalidConfig`, `ErrNilPayload`) support direct comparison.
- HTTP status mapping eliminates error-to-status boilerplate in API servers.
- The `errors.IsCode()` helper provides concise error classification.
- Retryability classification enables intelligent retry loops.

---

## ADR-016: Generic WorkerPool Interface for Testability

**Status**: Accepted

**Context**: The initial worker pool implementation was a concrete `*Pool[T, R]` struct with no interface. The `batch.Processor` depended directly on the concrete type, making it impossible to inject mock pools for testing without the `WithBatchPoolFactory` workaround.

**Decision**: Define a `WorkerPool[T, R]` interface with `Process()` and `Workers()` methods. The concrete `Pool[T, R]` implements this interface. The `batch.Processor` accepts a pool factory function, defaulting to `workerpool.New()` but allowing mock injection.

```go
type WorkerPool[T any, R any] interface {
    Process(ctx context.Context, jobs []T, fn JobFunc[T, R]) ([]Result[T, R], error)
    Workers() int
}
```

**Consequences**:
- The worker pool is fully mockable for testing.
- The `batch.Processor` can be tested without actual goroutine pools.
- New pool implementations (e.g., distributed, priority-based) can be swapped in.
- Compile-time interface satisfaction is enforced via `var _ WorkerPool[int, int] = (*Pool[int, int])(nil)`.

---

## ADR-017: Public API Surface Segregation (client.go)

**Status**: Accepted

**Context**: The root package grew organically, mixing public API definitions (interface, constructors, type aliases) with type definitions (ECLevel, Format), implementation details (generator struct), convenience functions (Quick helpers), and renderer re-exports. This made it difficult to identify the stable public contract.

**Decision**: Split the root package into clearly delineated files with documented responsibilities:

| File | Purpose |
|------|---------|
| `client.go` | Client interface, constructors, backward compat aliases, context helpers |
| `qrcode.go` | ECLevel, Format types and constants, renderer re-exports |
| `builder.go` | Fluent Builder API |
| `generator.go` | Client implementation (unexported `generator` struct) |
| `config.go` | Config, ConfigPatch, ApplyPatch, validation, pointer helpers |
| `options.go` | Functional option definitions |
| `helpers.go` | Quick* one-liner convenience functions |

**Consequences**:
- The stable public contract (Client interface) is clearly isolated in `client.go`.
- Internal implementation (`generator.go`) is never imported by external consumers.
- Each file has a single, well-defined responsibility.
- New contributors can quickly navigate to the relevant file.
- Backward compatibility shims are co-located with the API they support.

---

## ADR-018: Minimal Renderer Re-Export Layer

**Status**: Accepted

**Context**: Renderers live in `internal/renderer/`, which is inaccessible to external consumers. However, advanced users need direct renderer access for custom module styles, gradient fills, and manual rendering pipelines.

**Decision**: Provide a thin re-export layer in `qrcode.go` that exposes:
- `GetRenderer(format)`, `NewPNGRenderer()`, `NewSVGRenderer()`, etc.
- Type aliases: `ModuleStyle`, `RenderOption`
- Convenience re-exports: `WithModuleStyle()`, `WithRoundedModules()`, `WithCircleModules()`, `WithGradient()`, `WithTransparency()`, `DefaultModuleStyle()`

These are type aliases that point directly to the internal types — no wrapping or indirection.

**Consequences**:
- Advanced users can access full renderer capabilities without importing internal packages.
- Type aliases have zero runtime overhead.
- The re-export layer is stable even if internal implementation changes.
- Godoc links show the canonical location of the types.
