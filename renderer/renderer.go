// Package renderer provides QR code rendering in multiple output formats
// including PNG, SVG, PDF, terminal Unicode, and base64 with support for
// custom module styles, gradients, and transparency.
//
// The package uses a functional-options pattern for configuration. Each
// renderer implements a common Render method that accepts a variadic list
// of RenderOption values, making it easy to compose complex rendering
// configurations.
//
// # Quick Start
//
//	r := renderer.NewPNGRenderer()
//	err := r.Render(ctx, qrCode, os.Stdout)
//
// # Custom Styling
//
//	err := r.Render(ctx, qrCode, os.Stdout,
//	    renderer.WithGradient("#FF0000", "#0000FF", 45),
//	    renderer.WithRoundedModules(0.5),
//	    renderer.WithTransparency(0.8),
//	    renderer.WithForegroundColor("#000000"),
//	    renderer.WithBackgroundColor("#FFFFFF"),
//	)
//
// # Module Shapes
//
// The renderer supports four module shapes: "square", "rounded", "circle",
// and "diamond". Shape can be set via WithRoundedModules, WithCircleModules,
// or by constructing a ModuleStyle directly and passing it to WithModuleStyle.
//
// # Gradient Fills
//
// Linear gradient fills are supported for PNG and SVG output formats.
// Use WithGradient to specify start color, end color, and angle:
//
//	err := r.Render(ctx, qrCode, os.Stdout,
//	    renderer.WithGradient("#4A00E0", "#8E2DE2", 135),
//	)
//
// # Supported Formats
//
//   - PNG: raster image with full module-style support
//   - SVG: vector graphic with gradient and shape support
//   - PDF: document output (PDF 1.4)
//   - Terminal: Unicode block characters with ANSI color support
//   - Base64: base64-encoded PNG data URI for embedding in HTML/CSS
package renderer

import (
	"fmt"
	"strconv"
)

// Format enumerates the supported rendering output formats.
//
// Each format has an associated renderer type that implements the
// Render(context.Context, *encoding.QRCode, io.Writer, ...RenderOption) method.
type Format int

const (
	// FormatPNG renders the QR code as a PNG image.
	FormatPNG Format = iota
	// FormatSVG renders the QR code as an SVG vector graphic.
	FormatSVG
	// FormatTerminal renders the QR code as Unicode block characters.
	FormatTerminal
	// FormatPDF renders the QR code as a PDF document.
	FormatPDF
)

// String returns the lowercase file-extension style label for the format.
// For example, FormatPNG returns "png" and FormatSVG returns "svg".
func (f Format) String() string {
	switch f {
	case FormatPNG:
		return "png"
	case FormatSVG:
		return "svg"
	case FormatTerminal:
		return "terminal"
	case FormatPDF:
		return "pdf"
	default:
		return "unknown"
	}
}

// RenderOption is a functional option that modifies a RenderConfig at render time.
//
// RenderOption functions are designed to be passed to a renderer's Render method.
// They are applied in order, so later options override earlier ones for the
// same field. Options that depend on ModuleStyle (such as WithGradient,
// WithRoundedModules, WithCircleModules, and WithTransparency) will
// auto-initialize a default ModuleStyle if one is not already set.
//
// Example usage:
//
//	opts := []renderer.RenderOption{
//	    renderer.WithWidth(512),
//	    renderer.WithHeight(512),
//	    renderer.WithQuietZone(2),
//	}
//	err := renderer.NewPNGRenderer().Render(ctx, qr, os.Stdout, opts...)
type RenderOption func(*RenderConfig)

// RenderConfig holds all parameters that control QR code rendering.
//
// Use ApplyOptions to construct a RenderConfig from defaults with
// functional options, or call DefaultRenderConfig and modify fields directly.
// Typical users should prefer the RenderOption pattern.
//
// Example:
//
//	cfg := renderer.ApplyOptions(
//	    renderer.WithWidth(512),
//	    renderer.WithForegroundColor("#1A1A2E"),
//	    renderer.WithBackgroundColor("#F0F0F0"),
//	)
type RenderConfig struct {
	// Width is the target output width in pixels.
	// For raster formats (PNG), this directly controls image dimensions.
	// A scale factor is computed from Width and the total module count
	// (including quiet zones) to determine the per-module pixel size.
	Width int
	// Height is the target output height in pixels.
	// When set to the same value as Width, the QR code is rendered square.
	Height int
	// ForegroundColor is the module (dark segment) color as a 7-character
	// hex string in "#RRGGBB" format. Defaults to "#000000" (black).
	ForegroundColor string
	// BackgroundColor is the background color as a 7-character hex string
	// in "#RRGGBB" format. Defaults to "#FFFFFF" (white).
	BackgroundColor string
	// QuietZone is the number of quiet-zone (margin) modules surrounding
	// the QR code on all sides. The QR code specification requires at
	// least 4 quiet-zone modules for reliable scanning. Defaults to 4.
	QuietZone int
	// BorderWidth is an extra border in pixels added outside the quiet zone.
	// This is used by some renderers (e.g., PDF) to add visual padding.
	BorderWidth int
	// ModuleStyle controls the visual appearance of individual QR code modules,
	// including shape (square, rounded, circle, diamond), gradient fills,
	// and transparency. A nil value means default square modules with no
	// gradient and full opacity.
	ModuleStyle *ModuleStyle
}

// WithWidth sets the target output width in pixels.
// The per-module scale is automatically computed from this value and the
// total module count (QR matrix size + 2 × quiet zone).
func WithWidth(w int) RenderOption {
	return func(c *RenderConfig) {
		c.Width = w
	}
}

// WithHeight sets the target output height in pixels.
// For square QR codes, this should typically match the Width.
func WithHeight(h int) RenderOption {
	return func(c *RenderConfig) {
		c.Height = h
	}
}

// WithForegroundColor sets the foreground (module) color as a hex string.
// The color must be in "#RRGGBB" format, for example "#FF5733".
func WithForegroundColor(color string) RenderOption {
	return func(c *RenderConfig) {
		c.ForegroundColor = color
	}
}

// WithBackgroundColor sets the background color as a hex string.
// The color must be in "#RRGGBB" format, for example "#F5F5F5".
func WithBackgroundColor(color string) RenderOption {
	return func(c *RenderConfig) {
		c.BackgroundColor = color
	}
}

// WithQuietZone sets the number of quiet-zone (margin) modules around the
// QR code. The quiet zone is required for reliable scanner recognition;
// the QR specification mandates at least 4 modules. Increasing this value
// adds more whitespace around the code.
func WithQuietZone(qz int) RenderOption {
	return func(c *RenderConfig) {
		c.QuietZone = qz
	}
}

// WithBorderWidth sets the extra border width in pixels outside the quiet
// zone. This is primarily used by PDF rendering to add visual padding.
func WithBorderWidth(bw int) RenderOption {
	return func(c *RenderConfig) {
		c.BorderWidth = bw
	}
}

// ModuleStyle controls the visual appearance of individual QR code modules.
//
// A ModuleStyle can customize the shape of each dark module, apply linear
// gradient fills, and set transparency (alpha). Use the convenience option
// functions WithGradient, WithRoundedModules, WithCircleModules, and
// WithTransparency to build a ModuleStyle incrementally, or construct one
// directly and pass it to WithModuleStyle.
//
// Example (direct construction):
//
//	style := &renderer.ModuleStyle{
//	    Shape:          "rounded",
//	    Roundness:      0.6,
//	    GradientEnabled: true,
//	    GradientStart:  "#FF0000",
//	    GradientEnd:    "#0000FF",
//	    GradientAngle:  90,
//	    Transparency:   0.85,
//	}
//	err := r.Render(ctx, qr, os.Stdout, renderer.WithModuleStyle(style))
type ModuleStyle struct {
	// Shape is the module shape. Accepted values are:
	//   - "square"  (default): standard rectangular modules
	//   - "rounded": rounded rectangles; corner radius controlled by Roundness
	//   - "circle":  each module rendered as a circle inscribed in the module cell
	//   - "diamond": each module rendered as a diamond (rotated square)
	Shape string
	// Roundness controls the corner radius for rounded modules.
	// The value is clamped to [0.0, 1.0] where 0.0 produces sharp corners
	// and 1.0 produces a circle. Only meaningful when Shape is "rounded".
	Roundness float64
	// GradientEnabled enables a linear gradient fill on modules.
	// When true, the GradientStart and GradientEnd colors are used instead
	// of the renderer's ForegroundColor. Supported by PNG and SVG renderers.
	GradientEnabled bool
	// GradientStart is the gradient start color as a hex string ("#RRGGBB").
	// Ignored unless GradientEnabled is true.
	GradientStart string
	// GradientEnd is the gradient end color as a hex string ("#RRGGBB").
	// Ignored unless GradientEnabled is true.
	GradientEnd string
	// GradientAngle is the linear gradient angle in degrees, measured
	// clockwise from the positive x-axis. For example, 0° is left-to-right,
	// 90° is top-to-bottom, and 135° is diagonal top-left to bottom-right.
	GradientAngle float64
	// Transparency controls the module opacity (alpha channel). A value of
	// 1.0 means fully opaque and 0.0 means fully transparent.
	Transparency float64
}

// DefaultModuleStyle returns a ModuleStyle with default square shape, no
// gradient, and full opacity (Transparency 1.0). This is the starting point
// for all convenience option functions (WithGradient, WithRoundedModules, etc.)
// that need to initialize a ModuleStyle when none is set.
func DefaultModuleStyle() *ModuleStyle {
	return &ModuleStyle{
		Shape:        "square",
		Roundness:    0.0,
		Transparency: 1.0,
	}
}

// WithModuleStyle sets the module style for rendering. Pass a fully
// constructed ModuleStyle to control shape, gradient, and transparency.
// This option overrides any previous ModuleStyle configuration.
//
// Example:
//
//	style := &renderer.ModuleStyle{Shape: "circle"}
//	err := r.Render(ctx, qr, os.Stdout, renderer.WithModuleStyle(style))
func WithModuleStyle(style *ModuleStyle) RenderOption {
	return func(c *RenderConfig) {
		c.ModuleStyle = style
	}
}

// WithGradient enables a linear gradient fill from startColor to endColor at
// the given angle. Colors must be in "#RRGGBB" format. The angle is measured
// in degrees clockwise from the positive x-axis (0° = left-to-right,
// 90° = top-to-bottom).
//
// If a ModuleStyle has not been set by a previous option, a default one is
// initialized automatically.
//
// Example:
//
//	err := r.Render(ctx, qr, os.Stdout,
//	    renderer.WithGradient("#4A00E0", "#8E2DE2", 135),
//	)
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

// WithRoundedModules sets the module shape to rounded rectangles with the
// given roundness. The roundness value controls the corner radius as a
// fraction of the module size (0.0 = sharp square, 1.0 = circle).
// If a ModuleStyle has not been set by a previous option, a default one
// is initialized automatically.
//
// Example:
//
//	err := r.Render(ctx, qr, os.Stdout,
//	    renderer.WithRoundedModules(0.5),
//	)
func WithRoundedModules(roundness float64) RenderOption {
	return func(c *RenderConfig) {
		if c.ModuleStyle == nil {
			c.ModuleStyle = DefaultModuleStyle()
		}
		c.ModuleStyle.Shape = "rounded"
		c.ModuleStyle.Roundness = roundness
	}
}

// WithCircleModules sets the module shape to circles. Each module is
// rendered as a circle inscribed within its cell. If a ModuleStyle has
// not been set by a previous option, a default one is initialized automatically.
func WithCircleModules() RenderOption {
	return func(c *RenderConfig) {
		if c.ModuleStyle == nil {
			c.ModuleStyle = DefaultModuleStyle()
		}
		c.ModuleStyle.Shape = "circle"
	}
}

// WithTransparency sets the module opacity (alpha) value. A value of 1.0
// is fully opaque and 0.0 is fully transparent. If a ModuleStyle has not
// been set by a previous option, a default one is initialized automatically.
func WithTransparency(alpha float64) RenderOption {
	return func(c *RenderConfig) {
		if c.ModuleStyle == nil {
			c.ModuleStyle = DefaultModuleStyle()
		}
		c.ModuleStyle.Transparency = alpha
	}
}

// IsGradientEnabled reports whether gradient fill is active.
// Returns false if the receiver is nil.
func (s *ModuleStyle) IsGradientEnabled() bool {
	return s != nil && s.GradientEnabled
}

// IsRounded reports whether the module shape is "rounded".
// Returns false if the receiver is nil.
func (s *ModuleStyle) IsRounded() bool {
	return s != nil && s.Shape == "rounded"
}

// IsCircle reports whether the module shape is "circle".
// Returns false if the receiver is nil.
func (s *ModuleStyle) IsCircle() bool {
	return s != nil && s.Shape == "circle"
}

// Validate checks that the module style parameters are valid.
// It returns an error if:
//   - Shape is not one of "square", "rounded", "circle", or "diamond";
//   - Roundness is outside [0.0, 1.0];
//   - Transparency is outside [0.0, 1.0];
//   - Gradient is enabled but start or end color is not a valid "#RRGGBB" hex string.
//
// A nil receiver is always valid and returns nil.
func (s *ModuleStyle) Validate() error {
	if s == nil {
		return nil
	}
	validShapes := map[string]bool{
		"square":  true,
		"rounded": true,
		"circle":  true,
		"diamond": true,
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

// DefaultRenderConfig returns a RenderConfig with sensible defaults:
//   - Width: 256, Height: 256
//   - ForegroundColor: "#000000"
//   - BackgroundColor: "#FFFFFF"
//   - QuietZone: 4 (per QR specification)
//   - BorderWidth: 0
//   - ModuleStyle: nil (plain square modules)
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

// ApplyOptions creates a RenderConfig from DefaultRenderConfig and applies
// each of the given RenderOption functions in order. This is the recommended
// way to build a RenderConfig.
//
// Example:
//
//	cfg := renderer.ApplyOptions(
//	    renderer.WithWidth(512),
//	    renderer.WithQuietZone(2),
//	    renderer.WithGradient("#FF0000", "#0000FF", 45),
//	)
func ApplyOptions(opts ...RenderOption) *RenderConfig {
	c := DefaultRenderConfig()
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// ParseHexColor parses a "#RRGGBB" hex color string into its red, green,
// and blue components (each 0–255). The input must be exactly 7 characters
// long and start with '#'. Returns an error for malformed inputs.
//
// Example:
//
//	r, g, b, err := renderer.ParseHexColor("#1A2B3C")
//	// r=26, g=43, b=60, err=nil
func ParseHexColor(hex string) (uint8, uint8, uint8, error) {
	if len(hex) != 7 || hex[0] != '#' {
		return 0, 0, 0, fmt.Errorf("invalid hex color format %q: expected \"#RRGGBB\"", hex)
	}
	r, err := strconv.ParseUint(hex[1:3], 16, 8)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid red component in %q: %w", hex, err)
	}
	g, err := strconv.ParseUint(hex[3:5], 16, 8)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid green component in %q: %w", hex, err)
	}
	b, err := strconv.ParseUint(hex[5:7], 16, 8)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid blue component in %q: %w", hex, err)
	}
	return uint8(r), uint8(g), uint8(b), nil
}

// ScaleSize calculates the integer pixel scale factor needed to fit a QR
// code of the given matrix size (with quiet zone) into targetPixels.
// The scale is computed as targetPixels / (matrixSize + 2*quietZone).
// Returns a minimum of 1 to avoid zero-sized modules.
func ScaleSize(matrixSize, quietZone, targetPixels int) int {
	totalModules := matrixSize + 2*quietZone
	if totalModules <= 0 {
		return 1
	}
	scale := targetPixels / totalModules
	if scale < 1 {
		return 1
	}
	return scale
}
