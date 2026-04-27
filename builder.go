package qrcode

import (
	"context"
	"time"

	"github.com/os-gomod/qrcode/v2/payload"
)

// Builder provides a fluent API for constructing a Client.
// It is a convenience wrapper — all actual generation is delegated to the Client.
type Builder struct {
	opts []Option
}

// NewBuilder creates a new Builder with default configuration.
func NewBuilder() *Builder {
	return &Builder{}
}

// Size sets the default output image size in pixels.
func (b *Builder) Size(size int) *Builder {
	b.opts = append(b.opts, WithDefaultSize(size))
	return b
}

// Margin sets the quiet zone (margin) around the QR code.
func (b *Builder) Margin(margin int) *Builder {
	b.opts = append(b.opts, WithQuietZone(margin))
	return b
}

// ErrorCorrection sets the error correction level.
func (b *Builder) ErrorCorrection(level ECLevel) *Builder {
	b.opts = append(b.opts, WithErrorCorrection(level))
	return b
}

// Version sets the QR code version (1-40), or 0 for auto.
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

// MaskPattern sets the mask pattern (-1 for auto, 0-7).
func (b *Builder) MaskPattern(pattern int) *Builder {
	b.opts = append(b.opts, WithMaskPattern(pattern))
	return b
}

// Format sets the default output format.
func (b *Builder) Format(f Format) *Builder {
	b.opts = append(b.opts, WithDefaultFormat(f))
	return b
}

// ForegroundColor sets the foreground color (hex, e.g., "#000000").
func (b *Builder) ForegroundColor(color string) *Builder {
	b.opts = append(b.opts, WithForegroundColor(color))
	return b
}

// BackgroundColor sets the background color (hex, e.g., "#FFFFFF").
func (b *Builder) BackgroundColor(color string) *Builder {
	b.opts = append(b.opts, WithBackgroundColor(color))
	return b
}

// Logo sets a logo image to overlay on the QR code.
func (b *Builder) Logo(source string, sizeRatio float64) *Builder {
	b.opts = append(b.opts, WithLogo(source, sizeRatio))
	return b
}

// LogoOverlay enables or disables logo overlay.
func (b *Builder) LogoOverlay(enabled bool) *Builder {
	b.opts = append(b.opts, WithLogoOverlay(enabled))
	return b
}

// LogoTint sets the tint color applied to the logo.
func (b *Builder) LogoTint(color string) *Builder {
	b.opts = append(b.opts, WithLogoTint(color))
	return b
}

// WorkerCount sets the maximum concurrent workers for batch operations.
func (b *Builder) WorkerCount(n int) *Builder {
	b.opts = append(b.opts, WithWorkerCount(n))
	return b
}

// QueueSize sets the internal queue size.
func (b *Builder) QueueSize(n int) *Builder {
	b.opts = append(b.opts, WithQueueSize(n))
	return b
}

// Prefix sets a filename prefix for batch output.
func (b *Builder) Prefix(prefix string) *Builder {
	b.opts = append(b.opts, WithPrefix(prefix))
	return b
}

// AutoSize enables or disables automatic version selection.
func (b *Builder) AutoSize(auto bool) *Builder {
	b.opts = append(b.opts, WithAutoSize(auto))
	return b
}

// Options appends raw Option functions to the builder.
func (b *Builder) Options(opts ...Option) *Builder {
	b.opts = append(b.opts, opts...)
	return b
}

// Build creates a new Client from the accumulated options.
func (b *Builder) Build() (Client, error) {
	return New(b.opts...)
}

// MustBuild creates a new Client, panicking on error.
func (b *Builder) MustBuild() Client {
	return MustNew(b.opts...)
}

// Clone creates a copy of the builder with the same options.
func (b *Builder) Clone() *Builder {
	clone := &Builder{}
	clone.opts = make([]Option, len(b.opts))
	copy(clone.opts, b.opts)
	return clone
}

// ---------------------------------------------------------------------------
// Builder Quick helpers — identical signatures to the package-level Quick*
// functions, but using the Builder's accumulated options (logo, colors, etc.)
// instead of defaults.
// ---------------------------------------------------------------------------

// Quick generates a QR code as PNG bytes for the given text data.
func (b *Builder) Quick(data string, size ...int) ([]byte, error) {
	return b.quickRender(&payload.TextPayload{Text: data}, FormatPNG, size...)
}

// QuickSVG generates a QR code as an SVG string for the given text data.
func (b *Builder) QuickSVG(data string, size ...int) (string, error) {
	raw, err := b.quickRender(&payload.TextPayload{Text: data}, FormatSVG, size...)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

// QuickFile generates a QR code and saves it to a file.
func (b *Builder) QuickFile(data, path string, size ...int) error {
	ctx := context.Background()
	gen, err := b.Clone().Size(quickSize(size...)).Build()
	if err != nil {
		return err
	}
	defer func() { _ = gen.Close() }()
	return gen.Save(ctx, &payload.TextPayload{Text: data}, path)
}

// QuickURL generates a URL QR code as PNG bytes.
func (b *Builder) QuickURL(url string, size ...int) ([]byte, error) {
	return b.quickRender(&payload.URLPayload{URL: url}, FormatPNG, size...)
}

// QuickWiFi generates a WiFi QR code as PNG bytes.
func (b *Builder) QuickWiFi(ssid, password, encryption string, size ...int) ([]byte, error) {
	return b.quickRender(&payload.WiFiPayload{
		SSID:       ssid,
		Password:   password,
		Encryption: encryption,
	}, FormatPNG, size...)
}

// QuickContact generates a vCard QR code as PNG bytes.
func (b *Builder) QuickContact(firstName, lastName, phone, email string, size ...int) ([]byte, error) {
	return b.quickRender(&payload.VCardPayload{
		FirstName: firstName,
		LastName:  lastName,
		Phone:     phone,
		Email:     email,
	}, FormatPNG, size...)
}

// QuickSMS generates an SMS QR code as PNG bytes.
func (b *Builder) QuickSMS(phone, message string, size ...int) ([]byte, error) {
	return b.quickRender(&payload.SMSPayload{
		Phone:   phone,
		Message: message,
	}, FormatPNG, size...)
}

// QuickEmail generates an email QR code as PNG bytes.
func (b *Builder) QuickEmail(to, subject, body string, size ...int) ([]byte, error) {
	return b.quickRender(&payload.EmailPayload{
		To:      to,
		Subject: subject,
		Body:    body,
	}, FormatPNG, size...)
}

// QuickGeo generates a geo-location QR code as PNG bytes.
func (b *Builder) QuickGeo(lat, lng float64, size ...int) ([]byte, error) {
	return b.quickRender(&payload.GeoPayload{
		Latitude:  lat,
		Longitude: lng,
	}, FormatPNG, size...)
}

// QuickEvent generates a calendar event QR code as PNG bytes.
func (b *Builder) QuickEvent(title, location string, start, end time.Time, size ...int) ([]byte, error) {
	return b.quickRender(&payload.CalendarPayload{
		Title:    title,
		Location: location,
		Start:    start,
		End:      end,
	}, FormatPNG, size...)
}

// QuickPayment generates a PayPal QR code as PNG bytes.
func (b *Builder) QuickPayment(username, amount string, size ...int) ([]byte, error) {
	return b.quickRender(&payload.PayPalPayload{
		Username: username,
		Amount:   amount,
	}, FormatPNG, size...)
}

// quickRender is the single shared implementation for all Builder Quick methods.
// It builds a temporary client from the builder's options (with size override),
// renders to the specified format, and closes the client.
func (b *Builder) quickRender(p payload.Payload, format Format, size ...int) ([]byte, error) {
	ctx := context.Background()
	gen, err := b.Clone().Size(quickSize(size...)).Build()
	if err != nil {
		return nil, err
	}
	defer func() { _ = gen.Close() }()
	return gen.Render(ctx, p, format)
}
