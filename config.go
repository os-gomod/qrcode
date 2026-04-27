package qrcode

import (
	"errors"
	"fmt"
	"time"
)

// Config holds all configuration for QR code generation.
// Fields use value types (not pointers) for straightforward usage.
// To selectively override specific fields without zero-value ambiguity,
// use ConfigPatch with ApplyPatch.
type Config struct {
	DefaultVersion  int           `json:"default_version"`
	DefaultECLevel  string        `json:"default_ec_level"`
	MinVersion      int           `json:"min_version"`
	MaxVersion      int           `json:"max_version"`
	AutoSize        bool          `json:"auto_size"`
	WorkerCount     int           `json:"worker_count"`
	QueueSize       int           `json:"queue_size"`
	DefaultFormat   string        `json:"default_format"`
	DefaultSize     int           `json:"default_size"`
	QuietZone       int           `json:"quiet_zone"`
	ForegroundColor string        `json:"foreground_color"`
	BackgroundColor string        `json:"background_color"`
	MaskPattern     int           `json:"mask_pattern"`
	LogoSource      string        `json:"logo_source"`
	LogoSizeRatio   float64       `json:"logo_size_ratio"`
	LogoOverlay     bool          `json:"logo_overlay"`
	LogoTint        string        `json:"logo_tint"`
	Prefix          string        `json:"prefix"`
	SlowOperation   time.Duration `json:"slow_operation"`
}

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

// Clone returns a shallow copy of the config.
func (c *Config) Clone() *Config {
	clone := *c
	return &clone
}

// ConfigPatch provides explicit, zero-value-safe overrides for Config.
// Only fields that are non-nil are applied — nil fields are treated as
// "no change", eliminating the ambiguity of the old Merge method where
// zero values (0, "", false) could not be distinguished from "not set".
//
// Example:
//
//	patch := ConfigPatch{
//	    WorkerCount: intPtr(8),
//	    AutoSize:    boolPtr(false),
//	}
//	cfg := ApplyPatch(defaultConfig(), &patch)
type ConfigPatch struct {
	DefaultVersion  *int
	DefaultECLevel  *string
	MinVersion      *int
	MaxVersion      *int
	AutoSize        *bool
	WorkerCount     *int
	QueueSize       *int
	DefaultFormat   *string
	DefaultSize     *int
	QuietZone       *int
	ForegroundColor *string
	BackgroundColor *string
	MaskPattern     *int
	LogoSource      *string
	LogoSizeRatio   *float64
	LogoOverlay     *bool
	LogoTint        *string
	Prefix          *string
	SlowOperation   *time.Duration
}

// ApplyPatch applies all non-nil fields from patch to base, returning a new
// Config. The original base is not modified.
//
//nolint:gocyclo,cyclop // field-by-field patch application is inherently linear
func ApplyPatch(base *Config, patch *ConfigPatch) *Config {
	out := base.Clone()

	if patch.DefaultVersion != nil {
		out.DefaultVersion = *patch.DefaultVersion
	}
	if patch.DefaultECLevel != nil {
		out.DefaultECLevel = *patch.DefaultECLevel
	}
	if patch.MinVersion != nil {
		out.MinVersion = *patch.MinVersion
	}
	if patch.MaxVersion != nil {
		out.MaxVersion = *patch.MaxVersion
	}
	if patch.AutoSize != nil {
		out.AutoSize = *patch.AutoSize
	}
	if patch.WorkerCount != nil {
		out.WorkerCount = *patch.WorkerCount
	}
	if patch.QueueSize != nil {
		out.QueueSize = *patch.QueueSize
	}
	if patch.DefaultFormat != nil {
		out.DefaultFormat = *patch.DefaultFormat
	}
	if patch.DefaultSize != nil {
		out.DefaultSize = *patch.DefaultSize
	}
	if patch.QuietZone != nil {
		out.QuietZone = *patch.QuietZone
	}
	if patch.ForegroundColor != nil {
		out.ForegroundColor = *patch.ForegroundColor
	}
	if patch.BackgroundColor != nil {
		out.BackgroundColor = *patch.BackgroundColor
	}
	if patch.MaskPattern != nil {
		out.MaskPattern = *patch.MaskPattern
	}
	if patch.LogoSource != nil {
		out.LogoSource = *patch.LogoSource
	}
	if patch.LogoSizeRatio != nil {
		out.LogoSizeRatio = *patch.LogoSizeRatio
	}
	if patch.LogoOverlay != nil {
		out.LogoOverlay = *patch.LogoOverlay
	}
	if patch.LogoTint != nil {
		out.LogoTint = *patch.LogoTint
	}
	if patch.Prefix != nil {
		out.Prefix = *patch.Prefix
	}
	if patch.SlowOperation != nil {
		out.SlowOperation = *patch.SlowOperation
	}

	return out
}

// ConfigToPatch creates a ConfigPatch from a Config, where every field is
// set. This is useful when you want to serialize or compare a full config
// as a patch.
func ConfigToPatch(c *Config) ConfigPatch {
	return ConfigPatch{
		DefaultVersion:  &c.DefaultVersion,
		DefaultECLevel:  &c.DefaultECLevel,
		MinVersion:      &c.MinVersion,
		MaxVersion:      &c.MaxVersion,
		AutoSize:        &c.AutoSize,
		WorkerCount:     &c.WorkerCount,
		QueueSize:       &c.QueueSize,
		DefaultFormat:   &c.DefaultFormat,
		DefaultSize:     &c.DefaultSize,
		QuietZone:       &c.QuietZone,
		ForegroundColor: &c.ForegroundColor,
		BackgroundColor: &c.BackgroundColor,
		MaskPattern:     &c.MaskPattern,
		LogoSource:      &c.LogoSource,
		LogoSizeRatio:   &c.LogoSizeRatio,
		LogoOverlay:     &c.LogoOverlay,
		LogoTint:        &c.LogoTint,
		Prefix:          &c.Prefix,
		SlowOperation:   &c.SlowOperation,
	}
}

// Validate checks the config for invalid values and returns an error
// describing the first problem found.
//
//nolint:gocyclo,cyclop // multi-field validation requires sequential checks
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
		return errors.New("config: logo_source must be specified when logo_overlay is enabled")
	}
	if c.LogoSizeRatio > 0 && (c.LogoSizeRatio < 0.05 || c.LogoSizeRatio > 0.4) {
		return fmt.Errorf("config: logo_size_ratio must be between 0.05 and 0.4, got %.2f", c.LogoSizeRatio)
	}
	if c.MaskPattern < -1 || c.MaskPattern > 7 {
		return fmt.Errorf("config: mask_pattern must be between -1 and 7, got %d", c.MaskPattern)
	}
	return nil
}

// ValidatePatch checks a ConfigPatch for invalid values before applying.
// Only non-nil fields are validated. Returns nil if all set fields are valid.
//
//nolint:gocyclo,cyclop // mirrors Validate with nil checks
func ValidatePatch(patch *ConfigPatch) error {
	if patch.MinVersion != nil && patch.MaxVersion != nil {
		if *patch.MinVersion > *patch.MaxVersion {
			return fmt.Errorf("config patch: min_version (%d) must be <= max_version (%d)", *patch.MinVersion, *patch.MaxVersion)
		}
	}
	if patch.WorkerCount != nil {
		if *patch.WorkerCount < 1 || *patch.WorkerCount > 64 {
			return fmt.Errorf("config patch: worker_count must be between 1 and 64, got %d", *patch.WorkerCount)
		}
	}
	if patch.QueueSize != nil {
		if *patch.QueueSize < 1 {
			return fmt.Errorf("config patch: queue_size must be >= 1, got %d", *patch.QueueSize)
		}
	}
	if patch.DefaultSize != nil {
		if *patch.DefaultSize < 100 || *patch.DefaultSize > 4000 {
			return fmt.Errorf("config patch: default_size must be between 100 and 4000, got %d", *patch.DefaultSize)
		}
	}
	if patch.QuietZone != nil {
		if *patch.QuietZone < 0 || *patch.QuietZone > 20 {
			return fmt.Errorf("config patch: quiet_zone must be between 0 and 20, got %d", *patch.QuietZone)
		}
	}
	if patch.LogoOverlay != nil && *patch.LogoOverlay && patch.LogoSource != nil && *patch.LogoSource == "" {
		return errors.New("config patch: logo_source must be specified when logo_overlay is enabled")
	}
	if patch.LogoSizeRatio != nil && *patch.LogoSizeRatio > 0 {
		if *patch.LogoSizeRatio < 0.05 || *patch.LogoSizeRatio > 0.4 {
			return fmt.Errorf("config patch: logo_size_ratio must be between 0.05 and 0.4, got %.2f", *patch.LogoSizeRatio)
		}
	}
	if patch.MaskPattern != nil {
		if *patch.MaskPattern < -1 || *patch.MaskPattern > 7 {
			return fmt.Errorf("config patch: mask_pattern must be between -1 and 7, got %d", *patch.MaskPattern)
		}
	}
	return nil
}

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

// ---------------------------------------------------------------------------
// Pointer helpers for constructing ConfigPatch values.
// These are convenience functions to avoid verbose &value expressions.
// ---------------------------------------------------------------------------

// IntP returns a pointer to an int value.
func IntP(v int) *int { return &v }

// StringP returns a pointer to a string value.
func StringP(v string) *string { return &v }

// BoolP returns a pointer to a bool value.
func BoolP(v bool) *bool { return &v }

// Float64P returns a pointer to a float64 value.
func Float64P(v float64) *float64 { return &v }

// DurationP returns a pointer to a time.Duration value.
func DurationP(v time.Duration) *time.Duration { return &v }
