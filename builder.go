package qrcode

import (
	"bytes"
	"context"
	"os"
	"time"

	"github.com/os-gomod/qrcode/payload"
)

// Builder provides a fluent API for constructing a Generator with chained configuration.
// Each setter method returns the Builder itself, allowing calls to be chained.
// The accumulated options are applied when Build or MustBuild is called.
//
// Builder also provides one-shot convenience methods (Quick, QuickURL, QuickWiFi, etc.)
// that build a generator, render a single QR code, and immediately close the generator.
//
//	png, err := qrcode.NewBuilder().
//	    ErrorCorrection(qrcode.LevelQ).
//	    ForegroundColor("#1a1a2e").
//	    BackgroundColor("#e0e0e0").
//	    Size(512).
//	    Quick("https://example.com")
//
// For repeated use, call Build once and reuse the Generator:
//
//	gen, err := qrcode.NewBuilder().
//	    ErrorCorrection(qrcode.LevelH).
//	    QuietZone(4).
//	    Build()
type Builder struct {
	opts []Option
}

// NewBuilder creates a new Builder with default settings.
func NewBuilder() *Builder {
	return &Builder{}
}

// Size sets the default image size in pixels.
func (b *Builder) Size(size int) *Builder {
	b.opts = append(b.opts, WithDefaultSize(size))
	return b
}

// Margin sets the quiet-zone (margin) module count.
func (b *Builder) Margin(margin int) *Builder {
	b.opts = append(b.opts, WithQuietZone(margin))
	return b
}

// ErrorCorrection sets the default error correction level.
func (b *Builder) ErrorCorrection(level ErrorCorrectionLevel) *Builder {
	b.opts = append(b.opts, WithErrorCorrection(level))
	return b
}

// Version sets the QR code version (1–40).
func (b *Builder) Version(v int) *Builder {
	b.opts = append(b.opts, WithVersion(v))
	return b
}

// MinVersion sets the minimum QR code version.
func (b *Builder) MinVersion(v int) *Builder {
	b.opts = append(b.opts, WithMinVersion(v))
	return b
}

// MaxVersion sets the maximum QR code version.
func (b *Builder) MaxVersion(v int) *Builder {
	b.opts = append(b.opts, WithMaxVersion(v))
	return b
}

// MaskPattern sets the mask pattern (0–7).
func (b *Builder) MaskPattern(pattern int) *Builder {
	b.opts = append(b.opts, WithMaskPattern(pattern))
	return b
}

// Format sets the default output format.
func (b *Builder) Format(f Format) *Builder {
	b.opts = append(b.opts, WithDefaultFormat(f))
	return b
}

// ForegroundColor sets the foreground (module) color.
func (b *Builder) ForegroundColor(color string) *Builder {
	b.opts = append(b.opts, WithForegroundColor(color))
	return b
}

// BackgroundColor sets the background color.
func (b *Builder) BackgroundColor(color string) *Builder {
	b.opts = append(b.opts, WithBackgroundColor(color))
	return b
}

// Logo configures a centered logo overlay with the given image source and size ratio.
func (b *Builder) Logo(source string, sizeRatio float64) *Builder {
	b.opts = append(b.opts, WithLogo(source, sizeRatio))
	return b
}

// LogoOverlay enables or disables the logo overlay.
func (b *Builder) LogoOverlay(enabled bool) *Builder {
	b.opts = append(b.opts, WithLogoOverlay(enabled))
	return b
}

// LogoTint applies a tint color to the overlaid logo.
func (b *Builder) LogoTint(color string) *Builder {
	b.opts = append(b.opts, WithLogoTint(color))
	return b
}

// WorkerCount sets the number of concurrent batch workers.
func (b *Builder) WorkerCount(n int) *Builder {
	b.opts = append(b.opts, WithWorkerCount(n))
	return b
}

// QueueSize sets the internal work queue capacity.
func (b *Builder) QueueSize(n int) *Builder {
	b.opts = append(b.opts, WithQueueSize(n))
	return b
}

// Prefix sets a URI prefix prepended to encoded data.
func (b *Builder) Prefix(prefix string) *Builder {
	b.opts = append(b.opts, WithPrefix(prefix))
	return b
}

// AutoSize enables or disables automatic version selection.
func (b *Builder) AutoSize(auto bool) *Builder {
	b.opts = append(b.opts, WithAutoSize(auto))
	return b
}

// Options appends arbitrary Option values to the builder.
func (b *Builder) Options(opts ...Option) *Builder {
	b.opts = append(b.opts, opts...)
	return b
}

// Build creates a Generator from the accumulated configuration. Returns an error
// if the combined options produce an invalid configuration.
func (b *Builder) Build() (Generator, error) {
	return New(b.opts...)
}

// MustBuild is like Build but panics on error. Useful in short-lived programs
// where configuration errors should halt immediately.
//
//	gen := qrcode.NewBuilder().Size(512).MustBuild()
func (b *Builder) MustBuild() Generator {
	return MustNew(b.opts...)
}

// Clone returns a shallow copy of the builder with the same options.
// The returned Builder can be modified independently. Used internally by
// convenience methods to avoid mutating the caller's Builder.
func (b *Builder) Clone() *Builder {
	clone := &Builder{}
	clone.opts = make([]Option, len(b.opts))
	copy(clone.opts, b.opts)
	return clone
}

// buildQuick creates a generator from the builder with the given size and renders
// the payload as PNG bytes. This is the shared implementation for all Builder.Quick*
// convenience methods.
func (b *Builder) buildQuick(p payload.Payload, size ...int) ([]byte, error) {
	gen, err := b.Clone().Size(quickSize(size...)).Build()
	if err != nil {
		return nil, err
	}
	defer gen.Close(context.Background()) //nolint:errcheck // Close error intentionally ignored in fire-and-forget convenience method
	return generatePNG(gen, p)
}

// buildQuickSVG creates a generator and renders the payload as SVG string.
func (b *Builder) buildQuickSVG(p payload.Payload, size ...int) (string, error) {
	gen, err := b.Clone().Size(quickSize(size...)).Build()
	if err != nil {
		return "", err
	}
	defer gen.Close(context.Background()) //nolint:errcheck // Close error intentionally ignored in fire-and-forget convenience method
	return generateSVG(gen, p)
}

// Quick generates a PNG QR code from the given text data. The optional size
// argument specifies the image width/height in pixels; defaults to 256.
func (b *Builder) Quick(data string, size ...int) ([]byte, error) {
	return b.buildQuick(&payload.TextPayload{Text: data}, size...)
}

// QuickSVG generates an SVG QR code from the given text data. The optional size
// argument specifies the image width/height in pixels; defaults to 256.
func (b *Builder) QuickSVG(data string, size ...int) (string, error) {
	return b.buildQuickSVG(&payload.TextPayload{Text: data}, size...)
}

// QuickFile generates a QR code from the given text data and writes it to the
// specified file path. The output format is inferred from the file extension
// (.png, .svg, or .pdf).
func (b *Builder) QuickFile(data, path string, size ...int) error {
	gen, err := b.Clone().Size(quickSize(size...)).Build()
	if err != nil {
		return err
	}
	defer gen.Close(context.Background()) //nolint:errcheck // Close error intentionally ignored in fire-and-forget convenience method
	return saveFile(gen, &payload.TextPayload{Text: data}, path)
}

// QuickURL generates a PNG QR code encoding the given URL.
func (b *Builder) QuickURL(url string, size ...int) ([]byte, error) {
	return b.buildQuick(&payload.URLPayload{URL: url}, size...)
}

// QuickWiFi generates a PNG QR code encoding a WiFi network configuration.
// The encryption parameter should be one of the payload.Encryption* constants.
func (b *Builder) QuickWiFi(ssid, password, encryption string, size ...int) ([]byte, error) {
	return b.buildQuick(&payload.WiFiPayload{
		SSID:       ssid,
		Password:   password,
		Encryption: encryption,
	}, size...)
}

// BuildAndGeneratePNG builds a generator and produces a PNG byte slice from the
// payload. The generator is closed after rendering. For repeated use, call Build
// once and use Generator.GenerateToWriter directly.
func (b *Builder) BuildAndGeneratePNG(ctx context.Context, p payload.Payload) ([]byte, error) {
	gen, err := b.Build()
	if err != nil {
		return nil, err
	}
	defer gen.Close(ctx)       //nolint:errcheck // Close error intentionally ignored; generator lifecycle is short-lived
	return generatePNG(gen, p) //nolint:contextcheck // internal helper intentionally uses its own context
}

// BuildAndGenerateSVG builds a generator and produces an SVG string from the payload.
// The generator is closed after rendering.
func (b *Builder) BuildAndGenerateSVG(ctx context.Context, p payload.Payload) (string, error) {
	gen, err := b.Build()
	if err != nil {
		return "", err
	}
	defer gen.Close(ctx)       //nolint:errcheck // Close error intentionally ignored; generator lifecycle is short-lived
	return generateSVG(gen, p) //nolint:contextcheck // internal helper intentionally uses its own context
}

// BuildAndSave builds a generator and saves the rendered QR code to a file.
// The format is inferred from the file extension (.png, .svg, or .pdf).
// The generator is closed after saving.
func (b *Builder) BuildAndSave(ctx context.Context, p payload.Payload, path string) error {
	gen, err := b.Build()
	if err != nil {
		return err
	}
	defer gen.Close(ctx)          //nolint:errcheck // Close error intentionally ignored; generator lifecycle is short-lived
	return saveFile(gen, p, path) //nolint:contextcheck // internal helper intentionally uses its own context
}

// QuickContact generates a PNG QR code encoding a vCard contact with the
// given name, phone, and email. Additional vCard fields can be set by using
// the payload.Contact builder and Builder.BuildAndGeneratePNG.
func (b *Builder) QuickContact(firstName, lastName, phone, email string, size ...int) ([]byte, error) {
	return b.buildQuick(&payload.VCardPayload{
		FirstName: firstName,
		LastName:  lastName,
		Phone:     phone,
		Email:     email,
	}, size...)
}

// QuickSMS generates a PNG QR code encoding an SMS message.
func (b *Builder) QuickSMS(phone, message string, size ...int) ([]byte, error) {
	return b.buildQuick(&payload.SMSPayload{
		Phone:   phone,
		Message: message,
	}, size...)
}

// QuickEmail generates a PNG QR code encoding an email message.
func (b *Builder) QuickEmail(to, subject, body string, size ...int) ([]byte, error) {
	return b.buildQuick(&payload.EmailPayload{
		To:      to,
		Subject: subject,
		Body:    body,
	}, size...)
}

// QuickGeo generates a PNG QR code encoding a geographic location (geo: URI).
func (b *Builder) QuickGeo(lat, lng float64, size ...int) ([]byte, error) {
	return b.buildQuick(&payload.GeoPayload{
		Latitude:  lat,
		Longitude: lng,
	}, size...)
}

// QuickEvent generates a PNG QR code encoding a calendar event (iCalendar VEVENT).
func (b *Builder) QuickEvent(title, location string, start, end time.Time, size ...int) ([]byte, error) {
	return b.buildQuick(&payload.CalendarPayload{
		Title:    title,
		Location: location,
		Start:    start,
		End:      end,
	}, size...)
}

// generatePNG renders a QR code to PNG and returns the raw bytes.
//
//nolint:contextcheck // internal helper intentionally uses Background context
func generatePNG(gen Generator, p payload.Payload) ([]byte, error) {
	var buf bytes.Buffer
	if err := gen.GenerateToWriter(context.Background(), p, &buf, FormatPNG); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// generateSVG renders a QR code to SVG and returns the string content.
//
//nolint:contextcheck // internal helper intentionally uses Background context
func generateSVG(gen Generator, p payload.Payload) (string, error) {
	var buf bytes.Buffer
	if err := gen.GenerateToWriter(context.Background(), p, &buf, FormatSVG); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// saveFile renders a QR code and writes it to disk, inferring format from the file extension.
//
//nolint:contextcheck // internal helper intentionally uses Background context
func saveFile(gen Generator, p payload.Payload, path string) error {
	ext := extensionFromPath(path)
	var format Format
	switch ext {
	case ".svg":
		format = FormatSVG
	case ".pdf":
		format = FormatPDF
	default:
		format = FormatPNG
	}
	var buf bytes.Buffer
	if err := gen.GenerateToWriter(context.Background(), p, &buf, format); err != nil {
		return err
	}
	return os.WriteFile(path, buf.Bytes(), 0o644) //nolint:gosec // G306: output files are intentionally world-readable
}
