package renderer

import "fmt"

// RenderOption applies a configuration override to a RenderConfig.
type RenderOption func(*RenderConfig)

// RenderConfig holds per-call rendering parameters.
type RenderConfig struct {
	Width           int
	Height          int
	ForegroundColor string
	BackgroundColor string
	QuietZone       int
	BorderWidth     int
	ModuleStyle     *ModuleStyle
}

// ModuleStyle configures advanced module rendering options.
type ModuleStyle struct {
	Shape           string
	Roundness       float64
	GradientEnabled bool
	GradientStart   string
	GradientEnd     string
	GradientAngle   float64
	Transparency    float64
}

// DefaultModuleStyle returns a ModuleStyle with sensible defaults (square, opaque).
func DefaultModuleStyle() *ModuleStyle {
	return &ModuleStyle{
		Shape:        "square",
		Roundness:    0.0,
		Transparency: 1.0,
	}
}

// WithWidth sets the target width in pixels.
func WithWidth(w int) RenderOption {
	return func(c *RenderConfig) { c.Width = w }
}

// WithHeight sets the target height in pixels.
func WithHeight(h int) RenderOption {
	return func(c *RenderConfig) { c.Height = h }
}

// WithForegroundColor sets the foreground (dark module) color in "#RRGGBB" format.
func WithForegroundColor(color string) RenderOption {
	return func(c *RenderConfig) { c.ForegroundColor = color }
}

// WithBackgroundColor sets the background color in "#RRGGBB" format.
func WithBackgroundColor(color string) RenderOption {
	return func(c *RenderConfig) { c.BackgroundColor = color }
}

// WithQuietZone sets the quiet zone (margin) in modules.
func WithQuietZone(qz int) RenderOption {
	return func(c *RenderConfig) { c.QuietZone = qz }
}

// WithBorderWidth sets the border width in pixels (used by PDF).
func WithBorderWidth(bw int) RenderOption {
	return func(c *RenderConfig) { c.BorderWidth = bw }
}

// WithModuleStyle sets the complete module style configuration.
func WithModuleStyle(style *ModuleStyle) RenderOption {
	return func(c *RenderConfig) { c.ModuleStyle = style }
}

// WithGradient enables a gradient fill from startColor to endColor at the given angle.
func WithGradient(startColor, endColor string, angle float64) RenderOption {
	return func(c *RenderConfig) {
		if c.ModuleStyle == nil {
			c.ModuleStyle = DefaultModuleStyle()
		}
		c.ModuleStyle.GradientEnabled = true
		c.ModuleStyle.GradientStart = startColor
		c.ModuleStyle.GradientEnd = endColor
		c.ModuleStyle.GradientAngle = angle
	}
}

// WithRoundedModules enables rounded rectangle modules with the given roundness (0.0–1.0).
func WithRoundedModules(roundness float64) RenderOption {
	return func(c *RenderConfig) {
		if c.ModuleStyle == nil {
			c.ModuleStyle = DefaultModuleStyle()
		}
		c.ModuleStyle.Shape = "rounded"
		c.ModuleStyle.Roundness = roundness
	}
}

// WithCircleModules enables circle-shaped modules.
func WithCircleModules() RenderOption {
	return func(c *RenderConfig) {
		if c.ModuleStyle == nil {
			c.ModuleStyle = DefaultModuleStyle()
		}
		c.ModuleStyle.Shape = "circle"
	}
}

// WithTransparency sets the module transparency (alpha), 0.0 = fully transparent, 1.0 = opaque.
func WithTransparency(alpha float64) RenderOption {
	return func(c *RenderConfig) {
		if c.ModuleStyle == nil {
			c.ModuleStyle = DefaultModuleStyle()
		}
		c.ModuleStyle.Transparency = alpha
	}
}

// DefaultRenderConfig returns a RenderConfig with sensible defaults.
func DefaultRenderConfig() *RenderConfig {
	return &RenderConfig{
		Width:           256,
		Height:          256,
		ForegroundColor: "#000000",
		BackgroundColor: "#FFFFFF",
		QuietZone:       4,
		BorderWidth:     0,
	}
}

// ApplyOptions applies a sequence of RenderOption overrides to default config.
func ApplyOptions(opts ...RenderOption) *RenderConfig {
	c := DefaultRenderConfig()
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// IsGradientEnabled reports whether the style has an active gradient.
func (s *ModuleStyle) IsGradientEnabled() bool {
	return s != nil && s.GradientEnabled
}

// IsRounded reports whether the style uses rounded rectangles.
func (s *ModuleStyle) IsRounded() bool {
	return s != nil && s.Shape == "rounded"
}

// IsCircle reports whether the style uses circle modules.
func (s *ModuleStyle) IsCircle() bool {
	return s != nil && s.Shape == "circle"
}

// IsDiamond reports whether the style uses diamond modules.
func (s *ModuleStyle) IsDiamond() bool {
	return s != nil && s.Shape == "diamond"
}

// UseAdvanced returns true when the style requires non-standard rendering paths.
func (s *ModuleStyle) UseAdvanced() bool {
	return s != nil && (s.Shape != "square" || s.GradientEnabled || s.Transparency < 1.0)
}

// Validate checks the ModuleStyle fields and returns an error if any are invalid.
func (s *ModuleStyle) Validate() error {
	if s == nil {
		return nil
	}
	validShapes := map[string]bool{
		"square": true, "rounded": true, "circle": true, "diamond": true,
	}
	if !validShapes[s.Shape] {
		return fmt.Errorf("invalid module shape %q: must be one of square, rounded, circle, diamond", s.Shape)
	}
	if s.Roundness < 0.0 || s.Roundness > 1.0 {
		return fmt.Errorf("invalid roundness %f: must be between 0.0 and 1.0", s.Roundness)
	}
	if s.Transparency < 0.0 || s.Transparency > 1.0 {
		return fmt.Errorf("invalid transparency %f: must be between 0.0 and 1.0", s.Transparency)
	}
	if s.GradientEnabled {
		if _, _, _, err := ParseHexColor(s.GradientStart); err != nil {
			return fmt.Errorf("invalid gradient start color: %w", err)
		}
		if _, _, _, err := ParseHexColor(s.GradientEnd); err != nil {
			return fmt.Errorf("invalid gradient end color: %w", err)
		}
	}
	return nil
}
