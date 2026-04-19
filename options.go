package qrcode

import (
	"time"
)

// Option configures a Generator at construction time. Options are applied to a
// default Config in the order they are provided, allowing later options to
// override earlier ones.
//
// Options can also be applied after construction via Generator.SetOptions or
// Generator.GenerateWithOptions for per-call customization.
//
//	gen, err := qrcode.New(
//	    qrcode.WithErrorCorrection(qrcode.LevelH),
//	    qrcode.WithDefaultSize(400),
//	    qrcode.WithLogo("logo.png", 0.25),
//	)
type Option func(*Config)

// WithVersion sets the default QR code version (1–40). When set to a non-zero
// value, automatic version selection is disabled and the specified version is
// used for all generations. Pair with WithAutoSize(false) to enforce the version.
func WithVersion(v int) Option {
	return func(c *Config) {
		c.DefaultVersion = v
	}
}

// WithMinVersion sets the minimum QR code version the auto-sizer may select (1–40).
// When AutoSize is enabled, the encoder will not choose a version smaller than this.
func WithMinVersion(v int) Option {
	return func(c *Config) {
		c.MinVersion = v
	}
}

// WithMaxVersion sets the maximum QR code version the auto-sizer may select (1–40).
// Useful to limit the physical size of the generated QR code.
func WithMaxVersion(v int) Option {
	return func(c *Config) {
		c.MaxVersion = v
	}
}

// WithErrorCorrection sets the default error correction level. Higher levels
// increase damage tolerance but reduce data capacity. Use LevelH when a logo
// overlay is applied.
func WithErrorCorrection(level ErrorCorrectionLevel) Option {
	return func(c *Config) {
		c.DefaultECLevel = level.String()
	}
}

// WithAutoSize enables or disables automatic version selection based on data length.
// When enabled (the default), the encoder automatically picks the smallest QR
// version that fits the data at the configured error correction level.
func WithAutoSize(auto bool) Option {
	return func(c *Config) {
		c.AutoSize = auto
	}
}

// WithWorkerCount sets the number of worker goroutines for batch generation (1–64).
// This option affects the Generator.Batch method and the batch.Processor type.
func WithWorkerCount(n int) Option {
	return func(c *Config) {
		c.WorkerCount = n
	}
}

// WithQueueSize sets the capacity of the internal work queue (>=1). This option
// affects the batch.Processor pipeline.
func WithQueueSize(n int) Option {
	return func(c *Config) {
		c.QueueSize = n
	}
}

// WithDefaultFormat sets the default output format for rendering. Accepted values
// are FormatPNG, FormatSVG, FormatTerminal, and FormatPDF.
func WithDefaultFormat(f Format) Option {
	return func(c *Config) {
		c.DefaultFormat = f.String()
	}
}

// WithDefaultSize sets the default image size in pixels for both width and height.
// Valid range is 100–4000. The default is 300 pixels.
func WithDefaultSize(size int) Option {
	return func(c *Config) {
		c.DefaultSize = size
	}
}

// WithQuietZone sets the number of quiet-zone (margin) modules around the QR code.
// The quiet zone is the blank border required by QR readers for reliable scanning.
// Valid range is 0–20; the default is 4.
func WithQuietZone(zone int) Option {
	return func(c *Config) {
		c.QuietZone = zone
	}
}

// WithForegroundColor sets the foreground (module) color as a hex string (e.g. "#000000").
// This color is used for the dark modules of the QR code. Supports gradients when
// used together with renderer.WithGradient.
func WithForegroundColor(color string) Option {
	return func(c *Config) {
		c.ForegroundColor = color
	}
}

// WithBackgroundColor sets the background color as a hex string (e.g. "#FFFFFF").
// This color is used for the light modules and quiet zone of the QR code.
func WithBackgroundColor(color string) Option {
	return func(c *Config) {
		c.BackgroundColor = color
	}
}

// WithMaskPattern sets the mask pattern to use (0–7). Each mask pattern applies a
// different XOR formula to the data modules to optimize readability. Set to -1
// (the default) for automatic selection, which evaluates all eight patterns and
// picks the one with the lowest penalty score.
func WithMaskPattern(pattern int) Option {
	return func(c *Config) {
		c.MaskPattern = pattern
	}
}

// WithLogo configures a logo to overlay at the center of the QR code.
// sizeRatio is the fraction of the QR code image the logo occupies (e.g. 0.25
// means the logo width is 25% of the QR image width). Accepted range is 0.05–0.4.
// When a logo is used, LevelH error correction is strongly recommended.
func WithLogo(logoSource string, sizeRatio float64) Option {
	return func(c *Config) {
		c.LogoSource = logoSource
		c.LogoSizeRatio = sizeRatio
		c.LogoOverlay = true
	}
}

// WithLogoOverlay enables or disables logo overlay rendering. When enabled, the
// generator composites the configured logo image onto the center of the QR code.
// A logo source must be configured via WithLogo or by setting Config.LogoSource.
func WithLogoOverlay(enabled bool) Option {
	return func(c *Config) {
		c.LogoOverlay = enabled
	}
}

// WithLogoTint applies a tint color to the overlaid logo. The tint is specified as
// a hex string (e.g. "#1a1a2e") and is applied by multiplying the logo's pixel
// colors with the tint, producing a colorized overlay that matches the QR theme.
func WithLogoTint(color string) Option {
	return func(c *Config) {
		c.LogoTint = color
	}
}

// WithPrefix sets a URI prefix prepended to the encoded payload data. This is
// useful when the raw encoded output needs to be wrapped in a URI scheme or
// namespace (e.g. "https://example.com/qr?").
func WithPrefix(prefix string) Option {
	return func(c *Config) {
		c.Prefix = prefix
	}
}

// WithConcurrency sets the number of concurrent workers. This is an alias for
// WithWorkerCount provided for API clarity in batch-processing contexts.
func WithConcurrency(n int) Option {
	return func(c *Config) {
		c.WorkerCount = n
	}
}

// WithSlowOperation sets the threshold above which a generation is logged as slow.
// The default is 100ms. Set to 0 to disable slow-operation warnings.
func WithSlowOperation(d time.Duration) Option {
	return func(c *Config) {
		c.SlowOperation = d
	}
}
