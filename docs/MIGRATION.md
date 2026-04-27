# v1 → v2 Migration Guide

This guide helps you migrate from `os-gomod/qrcode` (v1) to `os-gomod/qrcode` (v2).

## Table of Contents

- [Module Path](#module-path-change)
- [Type Renames](#type-renames)
- [Signature Changes](#signature-changes)
- [Removed Functions](#removed-package-level-functions)
- [Context Parameter](#new-context-parameter)
- [Builder Pattern](#builder-pattern)
- [Batch Processing](#batch-processing)
- [Option Changes](#option-changes)
- [Error Handling](#error-handling)
- [New Payload Types](#new-payload-types)
- [Deprecations](#deprecations)
- [Quick Migration Example](#quick-migration-example)
- [Compatibility Notes](#compatibility-notes)

---

## Module Path Change

```go
// v1
import "github.com/os-gomod/qrcode"

// v2
import "github.com/os-gomod/qrcode"
```

The module path has **not changed**. Go's major version handling is managed via `go.mod`. Both versions can coexist if needed by using a `replace` directive during migration.

## Type Renames

| v1 Type | v2 Type | Notes |
|---------|---------|-------|
| `Generator` | `Client` | Type alias — compiles without changes |
| `ErrorCorrectionLevel` | `ECLevel` | Type alias — compiles without changes |
| — | `NewClient()` | Alias for `New()` |
| — | `MustNewClient()` | Alias for `MustNew()` |
| — | `WithECLevel()` | Alias for `WithErrorCorrection()` |

**No code changes required** — `Generator` is a type alias for `Client`, and `ErrorCorrectionLevel` is a type alias for `ECLevel`. Your existing v1 code will compile against v2 without modification.

Recommended for new code:

```go
// v1 style (still works, but deprecated)
gen, _ := qrcode.New()
gen, _ := qrcode.New(qrcode.WithErrorCorrection(qrcode.LevelH))

// v2 canonical style (recommended)
client, _ := qrcode.New()
client, _ := qrcode.New(qrcode.WithECLevel(qrcode.LevelH))
```

## Signature Changes

### Close()

```go
// v1
err := client.Close(ctx)

// v2 — context parameter removed
err := client.Close()
```

The `context.Context` parameter was removed because `Close()` is instantaneous — it only marks a guard as closed and releases no resources that need cancellation.

**Action required**: Remove the `ctx` argument from all `Close()` calls. This is the only signature change that will break compilation.

## Removed Package-Level Functions

These standalone functions were removed from the public API in v2:

| v1 Function | v2 Replacement |
|-------------|---------------|
| `GeneratePNG(p)` | `client.Render(ctx, p, FormatPNG)` |
| `GenerateSVG(p)` | `client.Render(ctx, p, FormatSVG)` |
| `GenerateASCII(p)` | `client.Render(ctx, p, FormatTerminal)` |
| `GenerateBase64(p)` | `client.Render(ctx, p, FormatBase64)` |
| `SavePNG(p, path)` | `client.Save(ctx, p, path)` |
| `SaveSVG(p, path)` | `client.Save(ctx, p, path)` |
| `Save(p, path)` | `client.Save(ctx, p, path)` |
| `generateBytes(p)` | `client.Render(ctx, p, FormatPNG)` |
| `generateString(p)` | `client.Render(ctx, p, FormatSVG)` |

**Note**: The `Quick*` helper functions (`Quick`, `QuickSVG`, `QuickFile`, `QuickURL`, `QuickWiFi`, `QuickContact`, `QuickSMS`, `QuickEmail`, `QuickGeo`, `QuickEvent`, `QuickPayment`) are **still available** in v2. They internally create a temporary client, render, and close it.

## New Context Parameter

All client methods now require `context.Context` as the first parameter:

```go
// v1
qr, err := client.Generate(p)
data, err := client.Render(p, FormatPNG)
_ = client.Save(p, "out.png")
results, err := client.Batch(ctx, payloads)

// v2 — context required on every method
qr, err := client.Generate(ctx, p)
data, err := client.Render(ctx, p, FormatPNG)
_ = client.Save(ctx, p, "out.png")
results, err := client.Batch(ctx, payloads)
```

This enables cancellation and timeout support for all operations:

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
data, err := client.Render(ctx, p, qrcode.FormatPNG)
```

## Builder Pattern

The Builder API is unchanged between v1 and v2:

```go
// v1 and v2 — identical API
b := qrcode.NewBuilder()
b.Size(512)
b.ErrorCorrection(qrcode.LevelH)
client, _ := b.Build()

// v2 also supports canonical names:
b2 := qrcode.NewBuilder()
// b2.ECLevel(qrcode.LevelH) — use ErrorCorrection for now
client2, _ := b2.Build()
```

New Builder features in v2:

- `b.Clone()` — Create a copy of the builder with the same options
- `b.MustBuild()` — Panics on invalid configuration (like `MustNew`)
- Builder `Quick*` helpers now inherit the builder's accumulated options

## Batch Processing

The batch API is unchanged:

```go
// v1 and v2 — identical API
client, _ := qrcode.New(qrcode.WithWorkerCount(4))
defer client.Close()

payloads := []payload.Payload{
    &payload.TextPayload{Text: "item-1"},
    &payload.TextPayload{Text: "item-2"},
}
results, _ := client.Batch(ctx, payloads)
```

The advanced `batch.Processor` API is also available:

```go
import "github.com/os-gomod/qrcode/batch"

proc := batch.NewProcessor(client,
    batch.WithBatchFormat(qrcode.FormatPNG),
    batch.WithBatchOutputDir("./output"),
    batch.WithBatchConcurrency(8),
)

items := []batch.Item{
    {ID: "qr1", Data: "Hello"},
    {ID: "qr2", Data: "World"},
}

results, stats, _ := proc.ProcessWithStats(ctx, items)
```

## Option Changes

| v1 Option | v2 Status | Notes |
|-----------|-----------|-------|
| `WithConcurrency(n)` | **Removed** | Duplicate of `WithWorkerCount` |
| All other options | Unchanged | Same function signatures |
| `WithQueueSize(n)` | New in v2 | Configure internal queue buffer |
| `WithLogoTint(color)` | New in v2 | Tint color for logo overlay |
| `WithSlowOperation(d)` | New in v2 | Duration threshold for slow operation logging |

## Error Handling

v2 introduces structured error types via the `errors` sub-package:

```go
import qrerrors "github.com/os-gomod/qrcode/errors"

if err != nil {
    switch {
    case qrerrors.IsCode(err, qrerrors.ErrCodeValidation):
        // Handle payload or config validation errors
    case qrerrors.IsCode(err, qrerrors.ErrCodeEncoding):
        // Handle QR encoding errors
    case qrerrors.IsCode(err, qrerrors.ErrCodeRendering):
        // Handle rendering errors
    case qrerrors.IsCode(err, qrerrors.ErrCodeStorage):
        // Handle file I/O errors
    case qrerrors.IsCode(err, qrerrors.ErrCodeClosed):
        // Handle operations on a closed client
    }
}
```

## New Payload Types

v2 adds several new payload types not available in v1:

| Payload | Type |
|---------|------|
| Event Ticket | `EventPayload` |
| Twitter Follow | `TwitterFollowPayload` |
| YouTube Video | `YouTubeVideoPayload` |
| Spotify Track | `SpotifyTrackPayload` |
| Spotify Playlist | `SpotifyPlaylistPayload` |
| Apple Music | `AppleMusicTrackPayload` |
| WhatsApp | `WhatsAppPayload` |
| Zoom | `ZoomPayload` |
| Google Maps | `GoogleMapsPayload` |
| Google Maps Place | `GoogleMapsPlacePayload` |
| Google Maps Directions | `GoogleMapsDirectionsPayload` |
| Apple Maps | `AppleMapsPayload` |
| App Store/Google Play | `MarketPayload` |
| Crypto | `CryptoPayload` |
| iBeacon | `IBeaconPayload` |
| NTP Locale | `NTPLocalePayload` |
| PID | `PIDPayload` |

## Deprecations

The following items are deprecated in v2 but still compile correctly:

| Deprecated | Replacement | Status |
|------------|-------------|--------|
| `Generator` type | `Client` | Type alias — compiles |
| `ErrorCorrectionLevel` type | `ECLevel` | Type alias — compiles |
| `Close(context.Context)` | `Close() error` | **Signature changed** — will not compile |
| `renderer.Format` | `qrcode.Format` | Use root package format constants |

## Quick Migration Example

```go
// ===== v1 code =====
package main

import (
    "context"
    "github.com/os-gomod/qrcode"
    "github.com/os-gomod/qrcode/payload"
)

func main() {
    ctx := context.Background()
    gen, _ := qrcode.New(qrcode.WithDefaultSize(256))
    defer gen.Close(ctx)                              // v1: Close takes context

    qr, _ := gen.Generate(ctx, &payload.TextPayload{Text: "hello"})
    png, _ := gen.GeneratePNG(ctx, &payload.TextPayload{Text: "hello"})  // removed
    _ = gen.SavePNG(ctx, &payload.TextPayload{Text: "hello"}, "out.png") // removed
}

// ===== v2 code (minimal changes) =====
package main

import (
    "context"
    "github.com/os-gomod/qrcode"
    "github.com/os-gomod/qrcode/payload"
)

func main() {
    ctx := context.Background()
    client, _ := qrcode.New(qrcode.WithDefaultSize(256))
    defer client.Close()                              // v2: no context argument

    qr, _ := client.Generate(ctx, &payload.TextPayload{Text: "hello"})
    png, _ := client.Render(ctx, &payload.TextPayload{Text: "hello"}, qrcode.FormatPNG)
    _ = client.Save(ctx, &payload.TextPayload{Text: "hello"}, "out.png")
}
```

## Compatibility Notes

1. **Type aliases guarantee source compatibility** — `Generator` and `ErrorCorrectionLevel` compile as-is.
2. **Only one breaking change** — `Close()` lost its `context.Context` parameter. A simple search-and-replace of `.Close(ctx)` with `.Close()` handles this.
3. **Removed functions have drop-in replacements** — `GeneratePNG` → `Render(ctx, p, FormatPNG)`, etc.
4. **Quick helpers are preserved** — All `Quick*` functions continue to work unchanged.
5. **Internal packages are inaccessible** — The `internal/` directory structure prevents external imports of implementation details, which improves API stability going forward.
6. **Configuration validation is stricter** — v2 validates config at construction time. Invalid values (e.g., size outside 100–4000, worker count outside 1–64) that were silently accepted in v1 now return errors.
