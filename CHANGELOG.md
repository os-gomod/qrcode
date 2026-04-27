# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [2.0.0] - 2026-04-27

### Added

- New internal package architecture: `encoding/`, `renderer/`, `payload/`, `batch/`, `errors/`, `logo/`, `internal/workerpool`, `internal/storage`, `internal/pool`, `internal/singleflight`, `internal/lifecycle`, `internal/hash`
- **Builder API** — fluent method chaining for constructing `Client` instances (`NewBuilder().Size(512).ErrorCorrection(LevelH).Build()`)
- **35 payload types** with full validation: Text, URL, WiFi, VCard, MeCard, SMS, MMS, Phone, Email, Geo, Google Maps, Google Maps Directions, Google Maps Place, Apple Maps, Calendar, Event, Twitter, Instagram, Facebook, LinkedIn, Telegram, YouTube Channel, YouTube Video, Spotify Track, Spotify Playlist, WhatsApp, Zoom, Market (Google Play/Apple App), PayPal, Crypto (BTC/ETH), iBeacon, and NTP Locale
- **5 renderers**: PNG, SVG, Terminal, PDF, Base64 — dispatched via `Renderer` interface registry
- **Batch processing** with `batch.Processor` supporting concurrent generation, JSON/CSV input, and directory output
- **Storage abstraction** (`storage.Storage` interface) for file I/O with `FileSystem` default implementation
- **Worker pool** (`internal/workerpool`) — generic bounded-concurrency pool with context cancellation
- **Singleflight deduplication** (`internal/singleflight`) — prevents concurrent duplicate generation calls
- **Lifecycle management** (`internal/lifecycle`) — tracks Client open/closed state
- **Buffer pool** (`internal/pool`) — `sync.Pool` wrapper for encoding buffers to reduce GC pressure
- **Context support** — all public methods accept `context.Context` as first parameter with cancellation propagation
- `Client` interface — canonical public contract
- `NewClient()` / `MustNewClient()` — canonical constructor aliases
- `WithECLevel()` — canonical option alias for `WithErrorCorrection()`
- `ECLevel` type — canonical alias for `ErrorCorrectionLevel`
- `FormatPNG`, `FormatSVG`, `FormatTerminal`, `FormatPDF`, `FormatBase64` — format type constants
- `Format.Extension()` — file extension from format
- `ModuleStyle` with support for rounded, circle, diamond, gradient, and transparent modules
- `WithRoundedModules()`, `WithCircleModules()`, `WithGradient()`, `WithTransparency()` render options
- `Payload` interface — formal contract for all payload types
- `errors.DomainError` interface with error codes, retryability, and HTTP status mapping
- `errors.BatchError` with proper error counting and per-item failure tracking
- `QRCodeError.WithMeta()` for structured error metadata
- Context helpers: `ContextWithQR()`, `QRFromContext()`

### Changed

- **Generator → Client rename**: `Generator` is now a type alias for `Client`; new code should use `Client`
- **Functional options pattern**: all configuration via `Option` functions (`WithDefaultSize()`, `WithWorkerCount()`, etc.)
- **Module-style rendering** with support for rounded, circle, diamond, gradient, and transparency effects
- **Gradient support**: linear gradient fill for QR modules via `renderer.WithGradient()`
- **Transparency support**: configurable per-module transparency via `renderer.WithTransparency()`
- File I/O abstracted through `storage.Storage` interface (no direct `os.WriteFile` in library code)
- Batch processing uses bounded `workerpool.Pool` instead of unlimited goroutines
- Renderer dispatch uses interface registry instead of type-switch
- Generation pipeline consolidated: all paths funnel through single `generate()` → `renderToWriter()` flow
- `qrcode_quick.go` absorbed into `helpers.go` (169 lines deleted)
- Batch worker loops consolidated into single `process()` method (~150 lines of duplication eliminated)

### Fixed

- **PNG renderer**: validation errors and encode errors are now properly propagated instead of being silently swallowed
- **WiFi payload**: escape handling for SSID and password fields with special characters (backslash, semicolon, colon)
- 3 compilation blockers (missing `Generator` interface, `New()`, `MustNew()`)
- `BatchError.Error()` printing wrong count for failed items
- Dead code: `if version == 0 && cfg.AutoSize { version = 0 }`
- `generate()` ignoring context parameter (no cancellation propagation)
- `BufferPool` allocated but never used — now properly integrated into encoding pipeline
- Config `Merge()` zero-value ambiguity (fields couldn't be explicitly set to zero)
- 17 hardcoded `context.Background()` calls eliminated

### Removed

Nothing. v1 compatibility is fully preserved through type aliases.

### Security

N/A — no security vulnerabilities identified or fixed in this release.

### Performance

- **Concurrent batch processing**: bounded worker pool enforces `WorkerCount` limit, preventing unbounded goroutine spawning
- **Singleflight deduplication**: identical concurrent generation requests are coalesced into a single call
- **Buffer pool for encoding**: `sync.Pool`-backed buffer reuse reduces GC pressure during batch operations

### Testing

- **87.7% overall coverage** across all packages (qrcode 94.0%, encoding 92.0%, errors 100.0%, renderer 92.5%)
- **24 benchmarks** covering Quick, Generate, Render, Batch, Builder, Encode, and parallel operations
- **Race condition tests**: zero data races with `go test -race`
- **Contract tests**: interface compliance verified at compile time for `Client`, `Renderer`, `Storage`, `Payload`
- 14 test files with ~2200+ test cases across all packages
