# README.md - Comprehensive Documentation

## QRcode - Feature-rich QR code generation library for Go

[![Go Reference](https://pkg.go.dev/badge/github.com/os-gomod/qrcode.svg)](https://pkg.go.dev/github.com/os-gomod/qrcode)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.22%2B-00ADD8.svg)](go.mod)
[![Zero Dependencies](https://img.shields.io/badge/deps-zero-green.svg)]()

A production-ready, feature-rich QR code generation library for Go with **zero external dependencies**. Generate QR codes from **30+ payload types**, render them in **5 output formats**, and customize appearance with advanced module shapes, gradient fills, transparency, and logo overlays.

## Overview

**qrcode** is a comprehensive Go library for generating QR codes in any application. It provides three progressively powerful API levels — quick one-liner functions, a reusable `Generator` interface, and a fluent `Builder` — so you can choose the right abstraction for your use case.

The library encodes data through strongly-typed payload structs that validate their own fields before encoding, covering everything from plain text and URLs to WiFi credentials, vCard contacts, calendar events, cryptocurrency payment URIs, social media profiles, Zoom meeting links, iBeacon advertisements, and more. Rendering supports PNG, SVG, PDF, terminal Unicode art, and base64 data URIs. Module appearance is fully customizable: choose square, rounded, circle, or diamond shapes; apply linear gradient fills at any angle; and control transparency for overlay-friendly QR codes. A logo overlay feature lets you brand your codes with a centered image.

The library is built with zero external dependencies — everything is implemented using the Go standard library. Internal utilities include buffer pooling for reduced allocations, singleflight deduplication of in-flight requests, lifecycle management for graceful shutdown, and structured logging.

## Installation

```bash
go get github.com/os-gomod/qrcode
```

## Quick Start

Generate a PNG QR code from a URL in just three lines:

```go
package main

import (
    "os"

    "github.com/os-gomod/qrcode"
)

func main() {
    data, err := qrcode.Quick("https://github.com/os-gomod/qrcode", 300)
    if err != nil {
        panic(err)
    }
    os.WriteFile("qr.png", data, 0644)
}
```

The `Quick` function creates a generator internally, renders the QR code at 300x300 px, and returns the raw PNG bytes. The optional `size` parameter defaults to 256 px when omitted.

## Usage

### Basic Generation

The library offers three API levels depending on how much control you need.

#### Quick Functions

One-shot helpers for the most common payloads. Each creates a short-lived generator, renders to PNG, and returns bytes.

```go
// Plain text
data, err := qrcode.Quick("Hello, World!", 300)

// URL
data, err := qrcode.QuickURL("https://example.com", 300)

// Write directly to a file (format inferred from extension)
err := qrcode.QuickFile("Hello, World!", "output.svg", 300)
err := qrcode.QuickFile("Hello, World!", "output.pdf", 300)
```

#### Generator Interface

For repeated use, create a `Generator` once and reuse it across calls. The generator manages its own resources — call `Close` when done.

```go
ctx := context.Background()

gen, err := qrcode.New(
    qrcode.WithDefaultSize(300),
    qrcode.WithErrorCorrection(qrcode.LevelH),
    qrcode.WithQuietZone(4),
)
if err != nil {
    panic(err)
}
defer gen.Close(ctx)

p := &payload.TextPayload{Text: "https://example.com"}
qr, err := gen.Generate(ctx, p)
if err != nil {
    panic(err)
}
// qr is an *encoding.QRCode — encode/decode as needed
```

The `Generator` interface exposes the following methods:

```go
type Generator interface {
    Generate(ctx context.Context, p payload.Payload) (*encoding.QRCode, error)
    GenerateWithOptions(ctx context.Context, p payload.Payload, opts ...Option) (*encoding.QRCode, error)
    GenerateToWriter(ctx context.Context, p payload.Payload, w io.Writer, format Format) error
    Batch(ctx context.Context, payloads []payload.Payload, opts ...Option) ([]*encoding.QRCode, error)
    Close(ctx context.Context) error
    Closed() bool
    SetOptions(opts ...Option) error
}
```

#### Builder Pattern

The `Builder` provides a fluent, chainable API for constructing a generator. It supports all the same options plus convenience methods for one-off generation.

```go
gen, err := qrcode.NewBuilder().
    Size(400).
    ErrorCorrection(qrcode.LevelH).
    ForegroundColor("#1A56DB").
    BackgroundColor("#F0F9FF").
    QuietZone(4).
    Build()
if err != nil {
    panic(err)
}
defer gen.Close(context.Background())
```

The builder also has shortcut methods that build a generator, render, and return the result in a single call:

```go
png, err := qrcode.NewBuilder().
    Size(300).
    ForegroundColor("#000000").
    QuickURL("https://example.com")

svg, err := qrcode.NewBuilder().
    Size(300).
    QuickSVG("Hello, World!")

err := qrcode.NewBuilder().
    Size(300).
    QuickFile("https://example.com", "output.svg")
```

### Payload Types

Every payload implements the `payload.Payload` interface:

```go
type Payload interface {
    Encode() (string, error)
    Type() string
    Validate() error
    Size() int
}
```

Payloads can be created directly as struct literals or through the validated builder functions in the `payload` package. The builder functions call `Validate()` before returning, so you get immediate feedback on malformed input.

#### Text and URL

```go
p, err := payload.Text("Hello, World!")
p, err := payload.URL("https://example.com")
```

#### WiFi

```go
// Standard WiFi network
p, err := payload.WiFi("MyNetwork", "secret123", "WPA2")

// Hidden SSID
p, err := payload.WiFiWithHidden("MyNetwork", "secret123", "WPA2")
```

#### vCard Contact

```go
p, err := payload.Contact("John", "Doe",
    payload.WithPhone("+1-555-0123"),
    payload.WithEmail("john@example.com"),
    payload.WithOrganization("Acme Corp"),
    payload.WithTitle("Engineer"),
    payload.WithAddress("123 Main St, Springfield"),
    payload.WithURL("https://example.com"),
    payload.WithNote("QR contact"),
)
```

#### MeCard

A lightweight alternative to vCard, commonly used on feature phones:

```go
p := &payload.MeCardPayload{
    Name:    "John Doe",
    Phone:   "+1-555-0123",
    Email:   "john@example.com",
    URL:     "https://example.com",
    Note:    "MeCard contact",
}
```

#### Email

```go
p, err := payload.Email("hello@example.com", "Subject Line", "Body text", "cc1@example.com", "cc2@example.com")
```

#### SMS

```go
p, err := payload.SMS("+15550123", "Hello from QR code!")
```

#### Phone

```go
p, err := payload.Phone("+15550123")
```

#### MMS

```go
p, err := payload.MMS("+15550123", "Check this out!")
```

#### Geo Location

```go
p, err := payload.Geo(37.421999, -122.084015)
```

#### Maps

```go
// Google Maps by coordinates
p, err := payload.GoogleMaps(37.421999, -122.084015)

// Google Maps by search query
p, err := payload.GoogleMapsQuery("Coffee shop near Times Square")

// Apple Maps
p, err := payload.AppleMaps(37.421999, -122.084015)
```

#### Calendar Event

```go
start := time.Date(2025, 9, 15, 9, 0, 0, 0, time.UTC)
end := time.Date(2025, 9, 15, 10, 0, 0, 0, time.UTC)

p, err := payload.Event("Team Standup", "Room 101", start, end,
    payload.WithAllDay(),
    payload.WithDescription("Weekly sync meeting"),
)
```

#### Event Ticket

```go
p := &payload.EventPayload{
    EventID:    "evt-42",
    EventName:  "Go Conference 2025",
    Venue:      "Convention Center",
    StartTime:  time.Date(2025, 9, 20, 9, 0, 0, 0, time.UTC),
    Category:   "Tech",
    Seat:       "A-12",
    Organizer:  "Go Foundation",
}
```

#### Social Media

```go
p, err := payload.Twitter("golang")
p, err := payload.LinkedIn("https://www.linkedin.com/in/johndoe")
p, err := payload.Instagram("golang")
p, err := payload.Facebook("https://www.facebook.com/golang")
p, err := payload.Telegram("golang")
p, err := payload.SpotifyTrack("3n3Ppam7vgaVa1iaRUc9Lp")
p, err := payload.YouTubeVideo("dQw4w9WgXcQ")
```

#### WhatsApp

```go
p, err := payload.WhatsApp("15550123", "Hi there!")
```

#### Zoom

```go
p, err := payload.Zoom("123-456-7890", "passcode")
```

#### Payment

```go
// PayPal
p, err := payload.Payment("johndoe", "25.00", "USD")

// Crypto (BTC, ETH, LTC)
p := &payload.CryptoPayload{
    Address:    "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa",
    Amount:     "0.001",
    Label:      "Donation",
    Message:    "Thanks!",
    CryptoType: payload.CryptoBTC,
}
```

#### iBeacon

```go
p, err := payload.IBeacon("A1B2C3D4-E5F6-7890-ABCD-EF1234567890", 1, 42)
```

#### NTP

```go
p := &payload.NTPLocalePayload{
    Host:        "time.google.com",
    Port:        "123",
    Version:     4,
    Description: "Google Public NTP",
}
```

#### App Store / Google Play

```go
p, err := payload.AppStore("id1234567890")
p, err := payload.PlayStore("com.example.app")
```

#### Swiss QR-Bill (PID)

```go
p := &payload.PIDPayload{
    PIDType:       "QRR",
    CreditorName:  "Acme Corp",
    IBAN:          "CH44 3199 9123 0008 8901 2",
    Reference:     "21 00000 00003 13947 14300 09017",
    Amount:        "100.00",
    Currency:      "CHF",
    DebtorName:    "John Doe",
    RemittanceInfo: "Invoice 2025-001",
}
```

### Rendering Options

The library renders QR codes through the `renderer` package, which supports multiple output formats and visual styles.

#### Output Formats

| Format | Constant | Description |
|--------|----------|-------------|
| PNG | `qrcode.FormatPNG` | Raster image with full styling support |
| SVG | `qrcode.FormatSVG` | Scalable vector graphic |
| PDF | `qrcode.FormatPDF` | PDF document |
| Terminal | `qrcode.FormatTerminal` | Unicode block characters for CLI output |
| Base64 | `qrcode.FormatBase64` | Base64-encoded PNG data URI (for HTML `<img>` tags) |

```go
// PNG bytes
pngData, err := qrcode.GeneratePNG(ctx, gen, p)

// SVG string
svgStr, err := qrcode.GenerateSVG(ctx, gen, p)

// Terminal (Unicode blocks)
asciiArt, err := qrcode.GenerateASCII(ctx, gen, p)
fmt.Print(asciiArt)

// PDF
var pdfBuf bytes.Buffer
err := gen.GenerateToWriter(ctx, p, &pdfBuf, qrcode.FormatPDF)

// Base64 (embed in HTML)
b64, err := qrcode.GenerateBase64(ctx, gen, p)
fmt.Printf(`<img src="%s" alt="QR Code">`, b64)

// Auto-detect format from file extension
err := qrcode.Save(ctx, gen, p, "output.svg")
err := qrcode.Save(ctx, gen, p, "output.pdf")
err := qrcode.Save(ctx, gen, p, "output.png")
```

#### Module Shapes

Customize the visual appearance of individual QR code modules using `renderer.ModuleStyle`:

| Shape | Value | Description |
|-------|-------|-------------|
| Square | `"square"` | Classic QR code modules |
| Rounded | `"rounded"` | Rounded rectangles with configurable corner radius |
| Circle | `"circle"` | Circular dots |
| Diamond | `"diamond"` | Diamond/rotated square modules |

```go
// Circle modules
style := &renderer.ModuleStyle{
    Shape: "circle",
}
```

#### Gradient Fill

Apply a linear gradient across the QR code modules:

```go
style := &renderer.ModuleStyle{
    Shape:            "rounded",
    Roundness:        0.5,
    GradientEnabled:  true,
    GradientStart:    "#FF6B35",
    GradientEnd:      "#004E89",
    GradientAngle:    135,
}
```

#### Transparency

Control module opacity for overlay-friendly QR codes:

```go
style := &renderer.ModuleStyle{
    Shape:        "circle",
    Transparency: 0.7, // 70% opacity
}
```

#### Complete Styled Rendering Example

```go
style := &renderer.ModuleStyle{
    Shape:            "rounded",
    Roundness:        0.4,
    GradientEnabled:  true,
    GradientStart:    "#6366F1",
    GradientEnd:      "#EC4899",
    GradientAngle:    45,
    Transparency:     0.9,
}
```

### Logo Overlay

Embed a logo in the center of the QR code. Higher error correction (LevelH) is recommended to ensure scannability.

```go
gen, err := qrcode.New(
    qrcode.WithDefaultSize(400),
    qrcode.WithErrorCorrection(qrcode.LevelH),
    qrcode.WithLogo("logo.png", 0.25),
    qrcode.WithLogoTint("#1A56DB"),
)
if err != nil {
    panic(err)
}
defer gen.Close(context.Background())

qr, err := gen.Generate(context.Background(), &payload.URLPayload{
    URL: "https://example.com",
})
```

Builder equivalent:

```go
gen, err := qrcode.NewBuilder().
    Size(400).
    ErrorCorrection(qrcode.LevelH).
    Logo("logo.png", 0.25).
    LogoTint("#1A56DB").
    Build()
```

Logo options:

| Option | Description |
|--------|-------------|
| `WithLogo(source, sizeRatio)` | Set logo image path and fractional size (0.05-0.40) |
| `WithLogoOverlay(enabled)` | Enable or disable logo overlay |
| `WithLogoTint(color)` | Apply a tint color to the logo |

### Error Correction Levels

Error correction determines how much of the QR code can be damaged while remaining scannable. Higher levels allow more damage but increase the QR code size for the same data.

| Level | Recovery | Use Case |
|-------|----------|----------|
| `LevelL` (~7%) | Minimal damage | Clean environments, maximum data capacity |
| `LevelM` (~15%) | Moderate damage | **Default** — good balance of size and durability |
| `LevelQ` (~25%) | Significant damage | Decorative QR codes, moderate logos |
| `LevelH` (~30%) | Extensive damage | Logo overlays, printed materials, harsh environments |

```go
// Set at construction time
gen, _ := qrcode.New(qrcode.WithErrorCorrection(qrcode.LevelH))

// Override per call
qr, _ := gen.GenerateWithOptions(ctx, p, qrcode.WithErrorCorrection(qrcode.LevelQ))
```

### Options Reference

#### Core Options

| Option | Type | Description |
|--------|------|-------------|
| `WithVersion(v int)` | Config | Set QR code version (1-40) |
| `WithMinVersion(v int)` | Config | Minimum version for auto-sizing |
| `WithMaxVersion(v int)` | Config | Maximum version for auto-sizing |
| `WithErrorCorrection(level)` | Config | Default error correction level |
| `WithAutoSize(bool)` | Config | Enable automatic version selection |
| `WithMaskPattern(int)` | Config | Mask pattern (0-7), -1 for auto |

#### Rendering Options

| Option | Type | Description |
|--------|------|-------------|
| `WithDefaultFormat(f Format)` | Config | Default output format |
| `WithDefaultSize(int)` | Config | Image size in pixels |
| `WithQuietZone(int)` | Config | Quiet zone (margin) module count |
| `WithForegroundColor(string)` | Config | Module color (`"#RRGGBB"`) |
| `WithBackgroundColor(string)` | Config | Background color (`"#RRGGBB"`) |

#### Logo Options

| Option | Type | Description |
|--------|------|-------------|
| `WithLogo(source, ratio)` | Config | Logo image path and size ratio |
| `WithLogoOverlay(bool)` | Config | Enable or disable logo overlay |
| `WithLogoTint(color)` | Config | Logo tint color |

#### Concurrency Options

| Option | Type | Description |
|--------|------|-------------|
| `WithWorkerCount(int)` | Config | Batch worker goroutine count |
| `WithQueueSize(int)` | Config | Internal work queue capacity |
| `WithConcurrency(int)` | Config | Alias for `WithWorkerCount` |

#### Miscellaneous

| Option | Type | Description |
|--------|------|-------------|
| `WithPrefix(string)` | Config | URI prefix for encoded data |

### Batch Generation

Generate multiple QR codes concurrently using the `Batch` method:

```go
payloads := []payload.Payload{
    &payload.URLPayload{URL: "https://example.com/1"},
    &payload.URLPayload{URL: "https://example.com/2"},
    &payload.URLPayload{URL: "https://example.com/3"},
}

gen, _ := qrcode.New(
    qrcode.WithWorkerCount(4),
    qrcode.WithQueueSize(16),
)
defer gen.Close(context.Background())

results, err := gen.Batch(context.Background(), payloads)
```

## Go Doc Reference

### Types

#### type Config

```go
type Config struct {
    DefaultVersion  int
    DefaultECLevel  string
    MinVersion      int
    MaxVersion      int
    AutoSize        bool
    WorkerCount     int
    QueueSize       int
    DefaultFormat   string
    DefaultSize     int
    QuietZone       int
    ForegroundColor string
    BackgroundColor string
    MaskPattern     int
    LogoSource      string
    LogoSizeRatio   float64
    LogoOverlay     bool
    LogoTint        string
    Prefix          string
    SlowOperation   time.Duration
}
```

Config holds all configuration parameters for the QR code generator. It is created internally via `defaultConfig()` and modified through `Option` functions. Use `New()` or `NewBuilder()` to construct a generator — do not create Config directly.

Methods:

- `func (c *Config) Clone() *Config` — Returns a shallow copy of the configuration.
- `func (c *Config) Merge(other *Config)` — Merges non-zero fields from other into c.
- `func (c *Config) Validate() error` — Checks that all configuration fields are within valid ranges.

#### type ErrorCorrectionLevel

```go
type ErrorCorrectionLevel int
```

Constants: `LevelL`, `LevelM`, `LevelQ`, `LevelH`.

Method: `func (l ErrorCorrectionLevel) String() string` — Returns "L", "M", "Q", or "H".

#### type Format

```go
type Format int
```

Constants: `FormatPNG`, `FormatSVG`, `FormatTerminal`, `FormatPDF`, `FormatBase64`.

Method: `func (f Format) String() string` — Returns "png", "svg", "terminal", "pdf", or "base64".

#### type Generator

```go
type Generator interface {
    Generate(ctx context.Context, p payload.Payload) (*encoding.QRCode, error)
    GenerateWithOptions(ctx context.Context, p payload.Payload, opts ...Option) (*encoding.QRCode, error)
    GenerateToWriter(ctx context.Context, p payload.Payload, w io.Writer, format Format) error
    Batch(ctx context.Context, payloads []payload.Payload, opts ...Option) ([]*encoding.QRCode, error)
    Close(ctx context.Context) error
    Closed() bool
    SetOptions(opts ...Option) error
}
```

The primary interface for creating QR codes. Use `New()` or `NewBuilder().Build()` to obtain an implementation. Call `Close()` when the generator is no longer needed to release resources.

#### type Option

```go
type Option func(*Config)
```

A functional option that modifies a Config at construction time. Options are passed to `New()`, `MustNew()`, or `NewBuilder().Options()`.

#### type Builder

```go
type Builder struct { /* unexported */ }
```

A fluent API for constructing a Generator with chained configuration.

Key methods:

- `func NewBuilder() *Builder`
- `func (b *Builder) Size(int) *Builder`
- `func (b *Builder) Margin(int) *Builder`
- `func (b *Builder) ErrorCorrection(ErrorCorrectionLevel) *Builder`
- `func (b *Builder) Version(int) *Builder`
- `func (b *Builder) MinVersion(int) *Builder`
- `func (b *Builder) MaxVersion(int) *Builder`
- `func (b *Builder) MaskPattern(int) *Builder`
- `func (b *Builder) Format(Format) *Builder`
- `func (b *Builder) ForegroundColor(string) *Builder`
- `func (b *Builder) BackgroundColor(string) *Builder`
- `func (b *Builder) Logo(string, float64) *Builder`
- `func (b *Builder) LogoOverlay(bool) *Builder`
- `func (b *Builder) LogoTint(string) *Builder`
- `func (b *Builder) WorkerCount(int) *Builder`
- `func (b *Builder) QueueSize(int) *Builder`
- `func (b *Builder) Prefix(string) *Builder`
- `func (b *Builder) AutoSize(bool) *Builder`
- `func (b *Builder) Options(...Option) *Builder`
- `func (b *Builder) Build() (Generator, error)`
- `func (b *Builder) MustBuild() Generator`
- `func (b *Builder) Clone() *Builder`
- `func (b *Builder) Quick(string, ...int) ([]byte, error)`
- `func (b *Builder) QuickSVG(string, ...int) (string, error)`
- `func (b *Builder) QuickFile(string, string, ...int) error`
- `func (b *Builder) QuickURL(string, ...int) ([]byte, error)`
- `func (b *Builder) QuickWiFi(string, string, string, ...int) ([]byte, error)`
- `func (b *Builder) QuickContact(string, string, string, string, ...int) ([]byte, error)`
- `func (b *Builder) QuickSMS(string, string, ...int) ([]byte, error)`
- `func (b *Builder) QuickEmail(string, string, string, ...int) ([]byte, error)`
- `func (b *Builder) QuickGeo(float64, float64, ...int) ([]byte, error)`
- `func (b *Builder) QuickEvent(string, string, time.Time, time.Time, ...int) ([]byte, error)`
- `func (b *Builder) BuildAndGeneratePNG(context.Context, payload.Payload) ([]byte, error)`
- `func (b *Builder) BuildAndGenerateSVG(context.Context, payload.Payload) (string, error)`
- `func (b *Builder) BuildAndSave(context.Context, payload.Payload, string) error`

### Package-level Functions

- `func New(opts ...Option) (Generator, error)` — Creates a new Generator with the given options.
- `func MustNew(opts ...Option) Generator` — Like New but panics on error.
- `func Quick(data string, size ...int) ([]byte, error)` — Generates a PNG QR code from text.
- `func QuickSVG(data string, size ...int) (string, error)` — Generates an SVG QR code from text.
- `func QuickFile(data, path string, size ...int) error` — Generates a QR code and writes to file.
- `func QuickURL(url string, size ...int) ([]byte, error)` — Generates a PNG QR code from a URL.
- `func QuickWiFi(ssid, password, encryption string, size ...int) ([]byte, error)` — Generates a WiFi QR code.
- `func QuickContact(firstName, lastName, phone, email string, size ...int) ([]byte, error)` — Generates a vCard QR code.
- `func QuickSMS(phone, message string, size ...int) ([]byte, error)` — Generates an SMS QR code.
- `func QuickEmail(to, subject, body string, size ...int) ([]byte, error)` — Generates an email QR code.
- `func QuickGeo(lat, lng float64, size ...int) ([]byte, error)` — Generates a geo location QR code.
- `func QuickEvent(title, location string, start, end time.Time, size ...int) ([]byte, error)` — Generates a calendar event QR code.
- `func QuickPayment(username, amount string, size ...int) ([]byte, error)` — Generates a PayPal payment QR code.
- `func GeneratePNG(ctx context.Context, gen Generator, p payload.Payload) ([]byte, error)` — Renders to PNG bytes.
- `func GenerateSVG(ctx context.Context, gen Generator, p payload.Payload) (string, error)` — Renders to SVG string.
- `func GenerateASCII(ctx context.Context, gen Generator, p payload.Payload) (string, error)` — Renders to terminal Unicode blocks.
- `func GenerateBase64(ctx context.Context, gen Generator, p payload.Payload) (string, error)` — Renders to base64 data URI.
- `func SavePNG(ctx context.Context, gen Generator, p payload.Payload, path string) error` — Saves as PNG file.
- `func SaveSVG(ctx context.Context, gen Generator, p payload.Payload, path string) error` — Saves as SVG file.
- `func Save(ctx context.Context, gen Generator, p payload.Payload, path string) error` — Saves with format inferred from extension.
- `func ContextWithQR(ctx context.Context, gen Generator) context.Context` — Stores a Generator in context.
- `func QRFromContext(ctx context.Context) (Generator, bool)` — Retrieves a Generator from context.

### Sub-packages

| Package | Description |
|---------|-------------|
| `encoding` | QR encoding engine — Galois fields, Reed-Solomon error correction, masking, matrix construction, version info tables |
| `payload` | 30+ typed payload structs with `Encode()`, `Validate()`, `Type()`, and `Size()` methods. Each payload validates its own fields. Builder functions (e.g., `payload.Text()`, `payload.URL()`) return validated payloads. |
| `renderer` | Output renderers for PNG, SVG, PDF, terminal, and base64. Supports module styling (shape, gradient, transparency) via `ModuleStyle`. |
| `logo` | Logo overlay processing — load images, resize, tint, and composite onto QR codes. |
| `batch` | Concurrent batch processing with configurable worker pool and work queue. |
| `errors` | Structured error types with error codes for programmatic handling. |
| `testing` | Contract tests (`GeneratorContractTest`) and assertion utilities for testing Generator implementations. |
| `internal/hash` | Fast non-cryptographic payload hashing for singleflight deduplication. |
| `internal/lifecycle` | Component lifecycle management (open/close state tracking). |
| `internal/pool` | Sync.Pool-based buffer pool for reducing allocations during encoding. |
| `internal/singleflight` | Deduplication of in-flight identical requests. |

---

## License

This project is licensed under the MIT License.

---

**Maintained by [os-gomod](https://github.com/os-gomod)** | [Report Bug](https://github.com/os-gomod/config/issues) | [Request Feature](https://github.com/os-gomod/config/issues)
