## QRcode - Feature-rich QR code generation library for Go

[![Go Reference](https://pkg.go.dev/badge/github.com/os-gomod/qrcode.svg)](https://pkg.go.dev/github.com/os-gomod/qrcode)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.22%2B-00ADD8.svg)](go.mod)
[![Zero Dependencies](https://img.shields.io/badge/deps-zero-green.svg)]()
A high-performance, production-ready QR code generation library for Go.
Zero external dependencies — pure stdlib only.

## Features

- **5 output formats** — PNG, SVG, Terminal (ANSI), PDF, Base64 data URI
- **35 payload types** — Text, URL, WiFi, vCard, MeCard, SMS, Email, Geo, Calendar, PayPal, Crypto, Social media, and more
- **Fluent Builder API** — Chain configuration for clean, readable code
- **Bounded batch processing** — Built-in worker pool with configurable concurrency
- **Logo overlay** — Embed logos with automatic resizing and tinting
- **Custom module styles** — Rounded, circle, diamond, gradient, and transparent modules
- **Context support** — Full `context.Context` propagation for cancellation and timeouts
- **Thread-safe** — Mutex-protected configuration with lifecycle management
- **Zero dependencies** — Pure Go standard library — no CGo, no external packages

## Installation

```bash
go get github.com/os-gomod/qrcode
```

## Quick Start

### Text QR Code

```go
package main

import (
    "context"
    "fmt"
    "os"

    "github.com/os-gomod/qrcode"
    "github.com/os-gomod/qrcode/payload"
)

func main() {
    client := qrcode.MustNew(qrcode.WithDefaultSize(512))
    defer client.Close()

    data, err := client.Render(context.Background(),
        &payload.TextPayload{Text: "Hello, World!"},
        qrcode.FormatPNG,
    )
    if err != nil {
        panic(err)
    }
    os.WriteFile("qrcode.png", data, 0644)
    fmt.Println("QR code saved to qrcode.png")
}
```

### URL QR Code

```go
data, _ := client.Render(ctx,
    &payload.URLPayload{URL: "https://example.com"},
    qrcode.FormatPNG,
)
```

### WiFi QR Code

```go
data, _ := client.Render(ctx,
    &payload.WiFiPayload{
        SSID:       "MyNetwork",
        Password:   "password123",
        Encryption: "WPA2",
    },
    qrcode.FormatPNG,
)
```

### SVG Output

```go
svg, _ := client.Render(ctx,
    &payload.TextPayload{Text: "https://example.com"},
    qrcode.FormatSVG,
)
os.WriteFile("qrcode.svg", svg, 0644)
```

### Save to File

The `Save` method infers the format from the file extension:

```go
_ = client.Save(ctx, &payload.TextPayload{Text: "Hello"}, "output.png")  // PNG
_ = client.Save(ctx, &payload.TextPayload{Text: "Hello"}, "output.svg")  // SVG
_ = client.Save(ctx, &payload.TextPayload{Text: "Hello"}, "output.pdf")  // PDF
```

## One-Liner Helpers

For simple use cases, the package-level `Quick*` functions create a temporary client, render, and close automatically:

```go
// PNG bytes
png, _ := qrcode.Quick("Hello, World!", 256)

// SVG string
svg, _ := qrcode.QuickSVG("Hello, World!", 256)

// Save directly to file
_ = qrcode.QuickFile("Hello, World!", "output.png", 512)

// URL QR code
urlQR, _ := qrcode.QuickURL("https://example.com")

// WiFi QR code
wifiQR, _ := qrcode.QuickWiFi("MyNetwork", "password123", "WPA2")

// vCard QR code
vcardQR, _ := qrcode.QuickContact("John", "Doe", "+1234567890", "john@example.com")

// SMS QR code
smsQR, _ := qrcode.QuickSMS("+1234567890", "Hi there!")

// Email QR code
emailQR, _ := qrcode.QuickEmail("hello@example.com", "Subject", "Body")

// Geo-location QR code
geoQR, _ := qrcode.QuickGeo(37.7749, -122.4194)

// Calendar event QR code
eventQR, _ := qrcode.QuickEvent("Meeting", "Room 1", start, end)

// PayPal payment QR code
payQR, _ := qrcode.QuickPayment("pay@example.com", "25.00")
```

## Builder API

The fluent Builder API provides a clean way to configure complex QR codes:

```go
client, err := qrcode.NewBuilder().
    Size(512).
    ErrorCorrection(qrcode.LevelH).
    Margin(8).
    ForegroundColor("#FF0000").
    BackgroundColor("#FFFFFF").
    Logo("logo.png", 0.25).
    WorkerCount(8).
    Build()
if err != nil {
    panic(err)
}
defer client.Close()
```

The Builder also supports `Quick*` helpers that inherit the builder's configuration:

```go
b := qrcode.NewBuilder().
    Size(256).
    ForegroundColor("#1A56DB").
    BackgroundColor("#F0F9FF")

png, _ := b.Quick("Custom branded QR code")
svg, _ := b.QuickSVG("Custom branded QR code")
_ = b.QuickFile("Custom branded QR code", "branded.png")
urlQR, _ := b.QuickURL("https://example.com")
```

## All Output Formats

```go
client := qrcode.MustNew()
ctx := context.Background()
p := &payload.TextPayload{Text: "Hello, World!"}

// PNG (raster image)
png, _ := client.Render(ctx, p, qrcode.FormatPNG)

// SVG (vector — scales to any size)
svg, _ := client.Render(ctx, p, qrcode.FormatSVG)

// Terminal (ANSI block characters for CLI output)
terminal, _ := client.Render(ctx, p, qrcode.FormatTerminal)

// PDF (embeddable in documents)
pdf, _ := client.Render(ctx, p, qrcode.FormatPDF)

// Base64 data URI (for embedding in HTML <img> tags)
b64, _ := client.Render(ctx, p, qrcode.FormatBase64)
```

## Payload Types (35 total)

| # | Payload | Type | Key Fields |
|---|---------|------|------------|
| 1 | **Text** | `TextPayload` | `Text string` |
| 2 | **URL** | `URLPayload` | `URL string, Title string` |
| 3 | **WiFi** | `WiFiPayload` | `SSID, Password, Encryption, Hidden bool` |
| 4 | **vCard** | `VCardPayload` | `FirstName, LastName, Phone, Email, Organization, Title, URL, Address, Note` |
| 5 | **MeCard** | `MeCardPayload` | `Name, Phone, Email, URL, Birthday, Note, Address, Nickname` |
| 6 | **SMS** | `SMSPayload` | `Phone string, Message string` |
| 7 | **MMS** | `MMSPayload` | `Phone, Subject, Message` |
| 8 | **Phone** | `PhonePayload` | `Number string` |
| 9 | **Email** | `EmailPayload` | `To, Subject, Body, CC []string` |
| 10 | **Geo** | `GeoPayload` | `Latitude, Longitude float64` |
| 11 | **Calendar** | `CalendarPayload` | `Title, Description, Location, Start, End time.Time, AllDay bool` |
| 12 | **Event Ticket** | `EventPayload` | `EventID, EventName, Venue, StartTime, Category, Seat, Organizer, Description, URL` |
| 13 | **PayPal** | `PayPalPayload` | `Username, Amount, Currency, Reference` |
| 14 | **Crypto** | `CryptoPayload` | `Address, Amount, CryptoType` |
| 15 | **Twitter** | `TwitterPayload` | `Username string` |
| 16 | **Twitter Follow** | `TwitterFollowPayload` | `Username string` |
| 17 | **Instagram** | `InstagramPayload` | `Username string` |
| 18 | **Facebook** | `FacebookPayload` | `PageURL string` |
| 19 | **LinkedIn** | `LinkedInPayload` | `ProfileURL string` |
| 20 | **Telegram** | `TelegramPayload` | `Username string` |
| 21 | **YouTube Video** | `YouTubeVideoPayload` | `VideoID string` |
| 22 | **YouTube Channel** | `YouTubeChannelPayload` | `ChannelID string` |
| 23 | **Spotify Track** | `SpotifyTrackPayload` | `TrackID string` |
| 24 | **Spotify Playlist** | `SpotifyPlaylistPayload` | `PlaylistID string` |
| 25 | **Apple Music** | `AppleMusicTrackPayload` | `TrackID, StoreFront string` |
| 26 | **WhatsApp** | `WhatsAppPayload` | `Phone, Message` |
| 27 | **Zoom** | `ZoomPayload` | `MeetingID, Password, DisplayName` |
| 28 | **Google Maps** | `GoogleMapsPayload` | `Latitude, Longitude, Query, Zoom int` |
| 29 | **Google Maps Place** | `GoogleMapsPlacePayload` | `PlaceName string` |
| 30 | **Google Maps Directions** | `GoogleMapsDirectionsPayload` | `Origin, Destination, TravelMode` |
| 31 | **Apple Maps** | `AppleMapsPayload` | `Latitude, Longitude, Query` |
| 32 | **App Store/Google Play** | `MarketPayload` | `Platform, PackageID, AppName, Campaign` |
| 33 | **iBeacon** | `IBeaconPayload` | `UUID, Major, Minor int, Manufacturer` |
| 34 | **NTP Locale** | `NTPLocalePayload` | `Host, Port, Version int, Description` |
| 35 | **PID** | `PIDPayload` | See `payload/pid.go` for fields |

## Advanced Features

### Batch Processing

Generate multiple QR codes concurrently with a bounded worker pool:

```go
client := qrcode.MustNew(qrcode.WithWorkerCount(8))
defer client.Close()

payloads := []payload.Payload{
    &payload.TextPayload{Text: "item-1"},
    &payload.TextPayload{Text: "item-2"},
    &payload.TextPayload{Text: "item-3"},
}
results, _ := client.Batch(ctx, payloads)
```

Advanced batch with the `batch` package:

```go
import "github.com/os-gomod/qrcode/batch"

proc := batch.NewProcessor(client,
    batch.WithBatchFormat(qrcode.FormatPNG),
    batch.WithBatchConcurrency(8),
)

items := []batch.Item{
    {ID: "qr1", Data: "Hello"},
    {ID: "qr2", Data: "World"},
}

results, stats, _ := proc.ProcessWithStats(ctx, items)
fmt.Printf("Generated %d QR codes in %v (avg: %v)\n", stats.Total, stats.TotalTime, stats.AvgTime)
```

### Custom Renderer Options

Use the public renderer re-exports for fine-grained rendering control:

```go
import (
    "bytes"
    qrcode "github.com/os-gomod/qrcode"
    "github.com/os-gomod/qrcode/payload"
)

qr, _ := client.Generate(ctx, &payload.TextPayload{Text: "Styled"})

var buf bytes.Buffer
pngR := qrcode.NewPNGRenderer()
data, _ := pngR.Render(ctx, qr,
    qrcode.WithModuleStyle(&qrcode.ModuleStyle{
        Shape:        "rounded",
        Roundness:    0.3,
        Transparency: 1.0,
    }),
    qrcode.WithForegroundColor("#FF0000"),
    qrcode.WithBackgroundColor("#FFFFFF"),
)
buf.Write(data)
```

### Module Styles

The PNG renderer supports custom module shapes:

```go
// Rounded modules
pngR.Render(ctx, qr, &buf,
    qrcode.WithModuleStyle(&qrcode.ModuleStyle{
        Shape:        "rounded",
        Roundness:    0.5,
        Transparency: 1.0,
    }),
)

// Circle modules (convenience helper)
pngR.Render(ctx, qr, &buf,
    qrcode.WithCircleModules(),
    qrcode.WithForegroundColor("#DC2626"),
)

// Diamond modules
pngR.Render(ctx, qr, &buf,
    qrcode.WithModuleStyle(&qrcode.ModuleStyle{
        Shape:        "diamond",
        Roundness:    0,
        Transparency: 1.0,
    }),
    qrcode.WithForegroundColor("#7C3AED"),
)

// Gradient foreground
pngR.Render(ctx, qr, &buf,
    qrcode.WithGradient("#059669", "#0891B2", 135),
)
```

### Logo Overlay

```go
client, _ := qrcode.New(
    qrcode.WithDefaultSize(400),
    qrcode.WithErrorCorrection(qrcode.LevelH),
    qrcode.WithLogo("logo.png", 0.25),
)
defer client.Close()

data, _ := client.Render(ctx, &payload.URLPayload{URL: "https://example.com"}, qrcode.FormatPNG)
```

Or use the `logo` package for manual control:

```go
import "github.com/os-gomod/qrcode/logo"

logoProc := logo.New("logo.png", 0.25)
logoImg, _ := logoProc.Load()

// Render QR, then overlay:
resizedLogo := logo.ResizeLogo(logoImg, qr.Size, 0.25)
final := logo.OverlayLogo(qrImage, resizedLogo, 4)
```

### Structured Error Handling

```go
import qrerrors "github.com/os-gomod/qrcode/errors"

if err != nil {
    if qrerrors.IsCode(err, qrerrors.ErrCodeValidation) {
        // Handle validation errors
    }
    if qrerrors.IsCode(err, qrerrors.ErrCodeEncoding) {
        // Handle encoding errors
    }
}
```

## Configuration Options

| Option | Description | Default |
|--------|-------------|---------|
| `WithDefaultSize(int)` | Output image size in pixels (100–4000) | `300` |
| `WithVersion(int)` | QR version (1–40), 0 for auto | `0` |
| `WithMinVersion(int)` | Minimum QR version | `1` |
| `WithMaxVersion(int)` | Maximum QR version | `40` |
| `WithAutoSize(bool)` | Automatic version selection | `true` |
| `WithErrorCorrection(ECLevel)` | Error correction level (L/M/Q/H) | `M` |
| `WithECLevel(ECLevel)` | Canonical alias for `WithErrorCorrection` | `M` |
| `WithQuietZone(int)` | Quiet zone (margin) in modules (0–20) | `4` |
| `WithWorkerCount(int)` | Max concurrent workers for batch (1–64) | `4` |
| `WithQueueSize(int)` | Internal queue size | `1024` |
| `WithDefaultFormat(Format)` | Default output format | `FormatPNG` |
| `WithForegroundColor(string)` | Foreground hex color | `#000000` |
| `WithBackgroundColor(string)` | Background hex color | `#FFFFFF` |
| `WithMaskPattern(int)` | Mask pattern (-1 for auto, 0–7) | `-1` |
| `WithLogo(string, float64)` | Logo source path and size ratio (0.05–0.4) | disabled |
| `WithLogoOverlay(bool)` | Enable/disable logo overlay | `false` |
| `WithLogoTint(string)` | Tint color applied to the logo | disabled |
| `WithPrefix(string)` | Filename prefix for batch output | disabled |

## Client Interface

```go
type Client interface {
    Generate(ctx context.Context, p payload.Payload) (*encoding.QRCode, error)
    GenerateWithOptions(ctx context.Context, p payload.Payload, opts ...Option) (*encoding.QRCode, error)
    GenerateToWriter(ctx context.Context, p payload.Payload, w io.Writer, format Format) error
    Render(ctx context.Context, p payload.Payload, format Format) ([]byte, error)
    Save(ctx context.Context, p payload.Payload, path string) error
    Batch(ctx context.Context, payloads []payload.Payload, opts ...Option) ([]*encoding.QRCode, error)
    Close() error
    SetOptions(opts ...Option) error
    Closed() bool
}
```

Constructors:

```go
client, err := qrcode.New(opts ...)            // Returns error on invalid config
client := qrcode.MustNew(opts ...)             // Panics on invalid config
client, err := qrcode.NewClient(opts ...)      // Alias for New()
client := qrcode.MustNewClient(opts ...)       // Alias for MustNew()
```

## Benchmarks

Run benchmarks with:

```bash
go test -bench=. -benchmem ./...
```

See `benchmark_test.go` for the full benchmark suite, which covers:

- `BenchmarkQuick` — One-liner PNG generation
- `BenchmarkQuickSVG` — One-liner SVG generation
- `BenchmarkGenerate` — Client-based QR encoding
- `BenchmarkGenerateWithOptions` — Per-call option overrides
- `BenchmarkRender_PNG` — PNG rendering pipeline
- `BenchmarkRender_SVG` — SVG rendering pipeline
- `BenchmarkRender_Terminal` — Terminal output rendering
- `BenchmarkGenerateToWriter_SVG` — Stream-based SVG output
- `BenchmarkBatch` — Concurrent batch generation (50 items)
- `BenchmarkNew` — Client construction overhead
- `BenchmarkBuilder` — Builder pattern construction
- `BenchmarkEncode` — Low-level QR matrix encoding

## Testing

```bash
# Run all tests with race detection and coverage
go test ./... -race -cover

# Run a specific package
go test ./payload/... -race -cover

# Run with verbose output
go test ./... -race -cover -v

# Run benchmarks
go test -bench=. -benchmem ./...
```

## Migration from v1

See [docs/MIGRATION.md](docs/MIGRATION.md) for the complete v1 to v2 migration guide.

### Key Changes

- `Generator` is a type alias for `Client` — existing code compiles unchanged
- `ErrorCorrectionLevel` is a type alias for `ECLevel`
- `Close(context.Context)` changed to `Close() error`
- Package-level render functions removed — use `client.Render()` instead
- New canonical names: `NewClient`, `MustNewClient`, `WithECLevel`

## Architecture

See [docs/ADR.md](docs/ADR.md) for Architecture Decision Records covering:

- Layered architecture with `internal/` packages
- Payload interface design
- Storage abstraction for file I/O
- Worker pool for batch processing
- Singleflight deduplication
- Functional options pattern
- Renderer registry pattern

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

[MIT](LICENSE)
