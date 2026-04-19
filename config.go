package qrcode

import (
	"fmt"
	"time"
)

// Config holds all configuration parameters for the QR code generator.
// Config is typically constructed implicitly through functional Option values
// passed to New or NewBuilder, but can also be manipulated directly for
// advanced use cases such as merging presets.
//
// The zero-value Config is not valid; use defaultConfig (via New) to obtain
// a Config with sensible defaults, then apply Option overrides.
//
//	cfg := qrcode.defaultConfig() // internal; use qrcode.New() instead
//	qrcode.WithErrorCorrection(qrcode.LevelH)(cfg)
type Config struct {
	DefaultVersion  int           `json:"default_version"`  // QR version (1–40); 0 means auto-select
	DefaultECLevel  string        `json:"default_ec_level"` // Error correction: "L", "M", "Q", "H"
	MinVersion      int           `json:"min_version"`      // Minimum version for auto-selection (1–40)
	MaxVersion      int           `json:"max_version"`      // Maximum version for auto-selection (1–40)
	AutoSize        bool          `json:"auto_size"`        // Enable automatic version sizing by data length
	WorkerCount     int           `json:"worker_count"`     // Concurrent goroutines for batch operations (1–64)
	QueueSize       int           `json:"queue_size"`       // Internal work queue capacity (>=1)
	DefaultFormat   string        `json:"default_format"`   // Default render format: "png", "svg", "terminal", "pdf"
	DefaultSize     int           `json:"default_size"`     // Default image width/height in pixels (100–4000)
	QuietZone       int           `json:"quiet_zone"`       // Quiet-zone margin modules around the QR code (0–20)
	ForegroundColor string        `json:"foreground_color"` // Hex color for dark modules (e.g. "#000000")
	BackgroundColor string        `json:"background_color"` // Hex color for light modules (e.g. "#FFFFFF")
	MaskPattern     int           `json:"mask_pattern"`     // Data mask pattern (0–7, or -1 for auto)
	LogoSource      string        `json:"logo_source"`      // File path or URL of the overlay logo image
	LogoSizeRatio   float64       `json:"logo_size_ratio"`  // Fraction of QR image occupied by logo (0.05–0.4)
	LogoOverlay     bool          `json:"logo_overlay"`     // Whether to render the logo overlay
	LogoTint        string        `json:"logo_tint"`        // Hex color to tint the logo image
	Prefix          string        `json:"prefix"`           // URI prefix prepended to encoded data
	SlowOperation   time.Duration `json:"slow_operation"`   // Threshold for logging slow generations
}

// defaultConfig creates a new Config with all defaults applied.
func defaultConfig() *Config {
	return &Config{
		DefaultECLevel:  "M",
		MinVersion:      1,
		MaxVersion:      40,
		AutoSize:        true,
		WorkerCount:     4,
		QueueSize:       1024,
		DefaultFormat:   "png",
		DefaultSize:     300,
		QuietZone:       4,
		ForegroundColor: "#000000",
		BackgroundColor: "#FFFFFF",
		MaskPattern:     -1,
		LogoSizeRatio:   0.25,
		SlowOperation:   100 * time.Millisecond,
	}
}

// Clone returns a shallow copy of the configuration. The returned Config can be
// modified independently without affecting the original. This is used internally
// to snapshot the configuration under a read lock before applying per-call options.
func (c *Config) Clone() *Config {
	clone := *c
	return &clone
}

// Merge overlays non-zero fields from other into c. Fields that are their
// zero value in other are left unchanged in c. This is useful for applying
// a partial configuration preset on top of a base configuration.
//
//	base := defaultConfig()
//	preset := &Config{DefaultSize: 512, ForegroundColor: "#1a1a2e"}
//	base.Merge(preset)
func (c *Config) Merge(other *Config) {
	if other == nil {
		return
	}
	if other.DefaultVersion != 0 {
		c.DefaultVersion = other.DefaultVersion
	}
	mergeString(&c.DefaultECLevel, other.DefaultECLevel)
	if other.MinVersion != 0 {
		c.MinVersion = other.MinVersion
	}
	if other.MaxVersion != 0 {
		c.MaxVersion = other.MaxVersion
	}
	if other.AutoSize {
		c.AutoSize = other.AutoSize
	}
	if other.WorkerCount != 0 {
		c.WorkerCount = other.WorkerCount
	}
	if other.QueueSize != 0 {
		c.QueueSize = other.QueueSize
	}
	mergeString(&c.DefaultFormat, other.DefaultFormat)
	if other.DefaultSize != 0 {
		c.DefaultSize = other.DefaultSize
	}
	if other.QuietZone != 0 {
		c.QuietZone = other.QuietZone
	}
	mergeString(&c.ForegroundColor, other.ForegroundColor)
	mergeString(&c.BackgroundColor, other.BackgroundColor)
	if other.MaskPattern != 0 {
		c.MaskPattern = other.MaskPattern
	}
	mergeString(&c.LogoSource, other.LogoSource)
	if other.LogoSizeRatio != 0 {
		c.LogoSizeRatio = other.LogoSizeRatio
	}
	if other.LogoOverlay {
		c.LogoOverlay = other.LogoOverlay
	}
	mergeString(&c.LogoTint, other.LogoTint)
	mergeString(&c.Prefix, other.Prefix)
	if other.SlowOperation != 0 {
		c.SlowOperation = other.SlowOperation
	}
}

// mergeString sets *dst to src when src is non-empty.
// This helper reduces cyclomatic complexity in [Config.Merge].
func mergeString(dst *string, src string) {
	if src != "" {
		*dst = src
	}
}

// Validate checks that all configuration fields are within valid ranges and
// returns an error describing the first violation found. Checked constraints
// include: MinVersion <= MaxVersion, DefaultVersion within range, WorkerCount
// (1–64), QueueSize (>=1), DefaultSize (100–4000), QuietZone (0–20),
// LogoSource required when LogoOverlay is true, LogoSizeRatio (0.05–0.4),
// and MaskPattern (-1–7).
func (c *Config) Validate() error {
	if c.MinVersion > c.MaxVersion {
		return fmt.Errorf("config: min_version (%d) must be <= max_version (%d)", c.MinVersion, c.MaxVersion)
	}
	if c.DefaultVersion != 0 && (c.DefaultVersion < c.MinVersion || c.DefaultVersion > c.MaxVersion) {
		return fmt.Errorf("config: default_version (%d) must be between %d and %d", c.DefaultVersion, c.MinVersion, c.MaxVersion)
	}
	if c.WorkerCount < 1 || c.WorkerCount > 64 {
		return fmt.Errorf("config: worker_count must be between 1 and 64, got %d", c.WorkerCount)
	}
	if c.QueueSize < 1 {
		return fmt.Errorf("config: queue_size must be >= 1, got %d", c.QueueSize)
	}
	if c.DefaultSize < 100 || c.DefaultSize > 4000 {
		return fmt.Errorf("config: default_size must be between 100 and 4000, got %d", c.DefaultSize)
	}
	if c.QuietZone < 0 || c.QuietZone > 20 {
		return fmt.Errorf("config: quiet_zone must be between 0 and 20, got %d", c.QuietZone)
	}
	if c.LogoOverlay && c.LogoSource == "" {
		return fmt.Errorf("config: logo_source must be specified when logo_overlay is enabled")
	}
	if c.LogoSizeRatio > 0 && (c.LogoSizeRatio < 0.05 || c.LogoSizeRatio > 0.4) {
		return fmt.Errorf("config: logo_size_ratio must be between 0.05 and 0.4, got %.2f", c.LogoSizeRatio)
	}
	if c.MaskPattern < -1 || c.MaskPattern > 7 {
		return fmt.Errorf("config: mask_pattern must be between -1 and 7, got %d", c.MaskPattern)
	}
	return nil
}

// parseECLevel converts an EC level string ("L", "M", "Q", "H") to its numeric value.
func parseECLevel(level string) int {
	switch level {
	case "L":
		return 0
	case "M":
		return 1
	case "Q":
		return 2
	case "H":
		return 3
	default:
		return -1
	}
}
