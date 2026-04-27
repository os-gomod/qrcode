package renderer

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"strings"
	"testing"

	"github.com/os-gomod/qrcode/v2/internal/encoding"
)

// qr is a shared test QR code used throughout this file.
var qr *encoding.QRCode

func init() {
	var err error
	qr, err = encoding.Encode([]byte("test-render"), 1, 0)
	if err != nil {
		panic("failed to create test QR code: " + err.Error())
	}
}

// ---------------------------------------------------------------------------
// Renderer interface & registry
// ---------------------------------------------------------------------------

func TestGetRenderer(t *testing.T) {
	tests := []struct {
		f    Format
		want bool
	}{
		{FormatPNG, true},
		{FormatSVG, true},
		{FormatTerminal, true},
		{FormatPDF, true},
		{FormatBase64, true},
		{Format(99), false},
	}
	for _, tt := range tests {
		r, err := GetRenderer(tt.f)
		if tt.want {
			if err != nil {
				t.Errorf("GetRenderer(%d) error: %v", tt.f, err)
			}
			if r == nil {
				t.Errorf("GetRenderer(%d) returned nil", tt.f)
			}
		} else {
			if err == nil {
				t.Errorf("GetRenderer(%d) should error", tt.f)
			}
		}
	}
}

func TestRegisterRenderer(t *testing.T) {
	// Save original registry
	orig := registry[Format(99)]
	defer func() { registry[Format(99)] = orig }()

	RegisterRenderer(Format(99), &mockRenderer{data: []byte("mock")})
	r, err := GetRenderer(Format(99))
	if err != nil {
		t.Fatalf("expected no error after registration, got: %v", err)
	}
	data, err := r.Render(context.Background(), qr)
	if err != nil {
		t.Fatalf("mock render error: %v", err)
	}
	if string(data) != "mock" {
		t.Errorf("expected mock data, got %q", string(data))
	}
}

type mockRenderer struct {
	data []byte
}

func (m *mockRenderer) Render(_ context.Context, _ *encoding.QRCode, _ ...RenderOption) ([]byte, error) {
	return m.data, nil
}

// ---------------------------------------------------------------------------
// Format tests
// ---------------------------------------------------------------------------

func TestFormatString(t *testing.T) {
	tests := []struct {
		f    Format
		want string
	}{
		{FormatPNG, "png"},
		{FormatSVG, "svg"},
		{FormatTerminal, "terminal"},
		{FormatPDF, "pdf"},
		{FormatBase64, "base64"},
		{Format(99), "unknown"},
	}
	for _, tt := range tests {
		if got := tt.f.String(); got != tt.want {
			t.Errorf("Format(%d).String() = %q, want %q", tt.f, got, tt.want)
		}
	}
}

func TestFormatContentType(t *testing.T) {
	tests := []struct {
		f    Format
		want string
	}{
		{FormatPNG, "image/png"},
		{FormatSVG, "image/svg+xml"},
		{FormatTerminal, "text/plain"},
		{FormatPDF, "application/pdf"},
		{FormatBase64, "text/plain"},
		{Format(99), "application/octet-stream"},
	}
	for _, tt := range tests {
		if got := tt.f.ContentType(); got != tt.want {
			t.Errorf("Format(%d).ContentType() = %q, want %q", tt.f, got, tt.want)
		}
	}
}

// ---------------------------------------------------------------------------
// Color utilities
// ---------------------------------------------------------------------------

func TestParseHexColor(t *testing.T) {
	tests := []struct {
		input   string
		wantR   uint8
		wantG   uint8
		wantB   uint8
		wantErr bool
	}{
		{"#000000", 0, 0, 0, false},
		{"#FFFFFF", 255, 255, 255, false},
		{"#FF0000", 255, 0, 0, false},
		{"#00FF00", 0, 255, 0, false},
		{"#0000FF", 0, 0, 255, false},
		{"#1a2b3c", 0x1a, 0x2b, 0x3c, false},
		{"invalid", 0, 0, 0, true},
		{"#12345", 0, 0, 0, true},
		{"#1234567", 0, 0, 0, true},
		{"123456", 0, 0, 0, true},
		{"", 0, 0, 0, true},
	}
	for _, tt := range tests {
		r, g, b, err := ParseHexColor(tt.input)
		if tt.wantErr {
			if err == nil {
				t.Errorf("ParseHexColor(%q): expected error", tt.input)
			}
		} else {
			if err != nil {
				t.Errorf("ParseHexColor(%q): unexpected error: %v", tt.input, err)
			}
			if r != tt.wantR || g != tt.wantG || b != tt.wantB {
				t.Errorf("ParseHexColor(%q) = (%d,%d,%d), want (%d,%d,%d)", tt.input, r, g, b, tt.wantR, tt.wantG, tt.wantB)
			}
		}
	}
}

func TestScaleSize(t *testing.T) {
	tests := []struct {
		matrix    int
		quietZone int
		target    int
		want      int
	}{
		{21, 4, 256, 8},
		{25, 2, 300, 10},
		{21, 4, 100, 3},
		{1, 4, 100, 11},
	}
	for _, tt := range tests {
		got := ScaleSize(tt.matrix, tt.quietZone, tt.target)
		if got != tt.want {
			t.Errorf("ScaleSize(%d, %d, %d) = %d, want %d", tt.matrix, tt.quietZone, tt.target, got, tt.want)
		}
	}
}

// ---------------------------------------------------------------------------
// Options & ModuleStyle
// ---------------------------------------------------------------------------

func TestDefaultRenderConfig(t *testing.T) {
	cfg := DefaultRenderConfig()
	if cfg.Width != 256 || cfg.Height != 256 {
		t.Errorf("default size = %dx%d, want 256x256", cfg.Width, cfg.Height)
	}
	if cfg.QuietZone != 4 {
		t.Errorf("default QuietZone = %d, want 4", cfg.QuietZone)
	}
	if cfg.ForegroundColor != "#000000" || cfg.BackgroundColor != "#FFFFFF" {
		t.Errorf("default colors = %s/%s", cfg.ForegroundColor, cfg.BackgroundColor)
	}
}

func TestApplyOptions(t *testing.T) {
	cfg := ApplyOptions(
		WithWidth(512),
		WithHeight(512),
		WithQuietZone(8),
		WithForegroundColor("#FF0000"),
		WithBackgroundColor("#00FF00"),
		WithBorderWidth(2),
	)
	if cfg.Width != 512 {
		t.Errorf("Width = %d, want 512", cfg.Width)
	}
	if cfg.Height != 512 {
		t.Errorf("Height = %d, want 512", cfg.Height)
	}
	if cfg.QuietZone != 8 {
		t.Errorf("QuietZone = %d, want 8", cfg.QuietZone)
	}
	if cfg.ForegroundColor != "#FF0000" {
		t.Errorf("ForegroundColor = %s", cfg.ForegroundColor)
	}
	if cfg.BackgroundColor != "#00FF00" {
		t.Errorf("BackgroundColor = %s", cfg.BackgroundColor)
	}
	if cfg.BorderWidth != 2 {
		t.Errorf("BorderWidth = %d, want 2", cfg.BorderWidth)
	}
}

func TestModuleStyle(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		ms := DefaultModuleStyle()
		if ms.Shape != "square" || ms.Roundness != 0.0 || ms.Transparency != 1.0 {
			t.Errorf("default module style = %+v", ms)
		}
	})
	t.Run("validate", func(t *testing.T) {
		ms := DefaultModuleStyle()
		if err := ms.Validate(); err != nil {
			t.Errorf("default should be valid: %v", err)
		}
	})
	t.Run("validate invalid shape", func(t *testing.T) {
		ms := &ModuleStyle{Shape: "star"}
		if err := ms.Validate(); err == nil {
			t.Error("invalid shape should fail validation")
		}
	})
	t.Run("validate roundness", func(t *testing.T) {
		ms := &ModuleStyle{Shape: "rounded", Roundness: 1.5}
		if err := ms.Validate(); err == nil {
			t.Error("roundness > 1.0 should fail")
		}
	})
	t.Run("validate transparency", func(t *testing.T) {
		ms := &ModuleStyle{Shape: "square", Transparency: -0.5}
		if err := ms.Validate(); err == nil {
			t.Error("negative transparency should fail")
		}
	})
	t.Run("validate nil", func(t *testing.T) {
		var ms *ModuleStyle
		if err := ms.Validate(); err != nil {
			t.Error("nil style should be valid")
		}
	})
	t.Run("UseAdvanced nil", func(t *testing.T) {
		var ms *ModuleStyle
		if ms.UseAdvanced() {
			t.Error("nil should not use advanced")
		}
	})
	t.Run("UseAdvanced square", func(t *testing.T) {
		ms := &ModuleStyle{Shape: "square", Transparency: 1.0}
		if ms.UseAdvanced() {
			t.Error("plain square should not use advanced")
		}
	})
	t.Run("UseAdvanced rounded", func(t *testing.T) {
		ms := &ModuleStyle{Shape: "rounded"}
		if !ms.UseAdvanced() {
			t.Error("rounded should use advanced")
		}
	})
	t.Run("UseAdvanced gradient", func(t *testing.T) {
		ms := &ModuleStyle{Shape: "square", GradientEnabled: true}
		if !ms.UseAdvanced() {
			t.Error("gradient should use advanced")
		}
	})
	t.Run("IsDiamond", func(t *testing.T) {
		ms := &ModuleStyle{Shape: "diamond"}
		if !ms.IsDiamond() {
			t.Error("diamond shape should return true")
		}
	})
	t.Run("nil checks", func(t *testing.T) {
		var ms *ModuleStyle
		if ms.IsGradientEnabled() || ms.IsRounded() || ms.IsCircle() || ms.IsDiamond() {
			t.Error("nil style should return false for all")
		}
	})
}

func TestWithGradient(t *testing.T) {
	cfg := ApplyOptions(WithGradient("#FF0000", "#0000FF", 45.0))
	if cfg.ModuleStyle == nil {
		t.Fatal("ModuleStyle should not be nil")
	}
	if !cfg.ModuleStyle.GradientEnabled {
		t.Error("gradient should be enabled")
	}
	if cfg.ModuleStyle.GradientAngle != 45.0 {
		t.Errorf("gradient angle = %f, want 45.0", cfg.ModuleStyle.GradientAngle)
	}
}

func TestWithRoundedModules(t *testing.T) {
	cfg := ApplyOptions(WithRoundedModules(0.5))
	if cfg.ModuleStyle == nil || cfg.ModuleStyle.Shape != "rounded" {
		t.Error("should set rounded shape")
	}
	if cfg.ModuleStyle.Roundness != 0.5 {
		t.Errorf("roundness = %f, want 0.5", cfg.ModuleStyle.Roundness)
	}
}

func TestWithCircleModules(t *testing.T) {
	cfg := ApplyOptions(WithCircleModules())
	if cfg.ModuleStyle == nil || cfg.ModuleStyle.Shape != "circle" {
		t.Error("should set circle shape")
	}
}

func TestWithTransparency(t *testing.T) {
	cfg := ApplyOptions(WithTransparency(0.5))
	if cfg.ModuleStyle == nil || cfg.ModuleStyle.Transparency != 0.5 {
		t.Error("should set transparency")
	}
}

// ---------------------------------------------------------------------------
// PNG renderer
// ---------------------------------------------------------------------------

func TestPNGRenderer_Render(t *testing.T) {
	ctx := context.Background()
	r := NewPNGRenderer()
	data, err := r.Render(ctx, qr, WithWidth(256), WithHeight(256))
	if err != nil {
		t.Fatalf("PNG render error: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("PNG output should not be empty")
	}
	// PNG magic bytes
	if len(data) >= 4 && (data[0] != 0x89 || data[1] != 0x50) {
		t.Error("output doesn't look like PNG")
	}
}

func TestPNGRenderer_InvalidColor(t *testing.T) {
	ctx := context.Background()
	r := NewPNGRenderer()
	_, err := r.Render(ctx, qr, WithForegroundColor("bad"))
	if err == nil {
		t.Error("expected error for invalid foreground color")
	}
}

func TestPNGRenderer_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	r := NewPNGRenderer()
	_, err := r.Render(ctx, qr)
	if err == nil {
		t.Error("expected error for cancelled context")
	}
}

func TestPNGRenderer_AdvancedStyles(t *testing.T) {
	ctx := context.Background()
	r := NewPNGRenderer()
	tests := []struct {
		name string
		opts []RenderOption
	}{
		{"rounded", []RenderOption{WithRoundedModules(0.5), WithWidth(128)}},
		{"circle", []RenderOption{WithCircleModules(), WithWidth(128)}},
		{"diamond", []RenderOption{WithModuleStyle(&ModuleStyle{Shape: "diamond", Transparency: 1.0}), WithWidth(128)}},
		{"gradient", []RenderOption{WithGradient("#FF0000", "#0000FF", 45), WithWidth(128)}},
		{"transparency", []RenderOption{WithTransparency(0.5), WithWidth(128)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := r.Render(ctx, qr, tt.opts...)
			if err != nil {
				t.Fatalf("%s: render error: %v", tt.name, err)
			}
			if len(data) == 0 {
				t.Fatalf("%s: output is empty", tt.name)
			}
		})
	}
}

func TestPNGRenderer_EdgeSizes(t *testing.T) {
	ctx := context.Background()
	r := NewPNGRenderer()
	sizes := []int{100, 200, 400, 1000}
	for _, size := range sizes {
		t.Run(fmtSize(size), func(t *testing.T) {
			data, err := r.Render(ctx, qr, WithWidth(size), WithHeight(size))
			if err != nil {
				t.Fatalf("size %d: render error: %v", size, err)
			}
			if len(data) == 0 {
				t.Fatalf("size %d: output is empty", size)
			}
		})
	}
}

func fmtSize(n int) string {
	return fmt.Sprintf("%dpx", n)
}

// ---------------------------------------------------------------------------
// SVG renderer
// ---------------------------------------------------------------------------

func TestSVGRenderer_Render(t *testing.T) {
	ctx := context.Background()
	r := NewSVGRenderer()
	data, err := r.Render(ctx, qr, WithWidth(256), WithHeight(256))
	if err != nil {
		t.Fatalf("SVG render error: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("SVG output should not be empty")
	}
	if !strings.Contains(string(data), "<svg") {
		t.Error("SVG output should contain <svg tag")
	}
	if !strings.Contains(string(data), "</svg>") {
		t.Error("SVG output should contain closing </svg>")
	}
}

func TestSVGRenderer_InvalidColor(t *testing.T) {
	ctx := context.Background()
	r := NewSVGRenderer()
	_, err := r.Render(ctx, qr, WithForegroundColor("bad"))
	if err == nil {
		t.Error("expected error for invalid color")
	}
}

func TestSVGRenderer_AdvancedStyles(t *testing.T) {
	ctx := context.Background()
	r := NewSVGRenderer()
	tests := []struct {
		name string
		opts []RenderOption
	}{
		{"rounded", []RenderOption{WithRoundedModules(0.5), WithWidth(128)}},
		{"circle", []RenderOption{WithCircleModules(), WithWidth(128)}},
		{"gradient", []RenderOption{WithGradient("#FF0000", "#0000FF", 45), WithWidth(128)}},
		{"transparency", []RenderOption{WithTransparency(0.6), WithWidth(128)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := r.Render(ctx, qr, tt.opts...)
			if err != nil {
				t.Fatalf("%s: render error: %v", tt.name, err)
			}
			if len(data) == 0 {
				t.Fatalf("%s: output is empty", tt.name)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Terminal renderer
// ---------------------------------------------------------------------------

func TestTerminalRenderer_Render(t *testing.T) {
	ctx := context.Background()
	r := NewTerminalRenderer()
	data, err := r.Render(ctx, qr, WithWidth(256))
	if err != nil {
		t.Fatalf("Terminal render error: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("Terminal output should not be empty")
	}
}

func TestTerminalRenderer_ColorModes(t *testing.T) {
	ctx := context.Background()
	r := NewTerminalRenderer()

	t.Run("custom colors produce ANSI escapes", func(t *testing.T) {
		data, err := r.Render(ctx, qr,
			WithForegroundColor("#FF0000"),
			WithBackgroundColor("#00FF00"),
		)
		if err != nil {
			t.Fatalf("render error: %v", err)
		}
		output := string(data)
		if !strings.Contains(output, "\x1b[") {
			t.Error("custom colors should produce ANSI escape sequences")
		}
	})

	t.Run("default colors no ANSI escapes", func(t *testing.T) {
		data, err := r.Render(ctx, qr)
		if err != nil {
			t.Fatalf("render error: %v", err)
		}
		output := string(data)
		if strings.Contains(output, "\x1b[") {
			t.Error("default colors should not produce ANSI escapes")
		}
	})
}

// ---------------------------------------------------------------------------
// PDF renderer
// ---------------------------------------------------------------------------

func TestPDFRenderer_Render(t *testing.T) {
	ctx := context.Background()
	r := NewPDFRenderer()
	data, err := r.Render(ctx, qr, WithWidth(256))
	if err != nil {
		t.Fatalf("PDF render error: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("PDF output should not be empty")
	}
	pdf := string(data)
	required := []string{"%PDF-1.4", "xref", "trailer", "%%EOF"}
	for _, marker := range required {
		if !strings.Contains(pdf, marker) {
			t.Errorf("PDF output missing required marker %q", marker)
		}
	}
	if !strings.Contains(pdf, "stream\n") || !strings.Contains(pdf, "\nendstream") {
		t.Error("PDF should contain a non-empty content stream")
	}
}

func TestPDFRenderer_InvalidColor(t *testing.T) {
	ctx := context.Background()
	r := NewPDFRenderer()
	_, err := r.Render(ctx, qr, WithForegroundColor("bad"))
	if err == nil {
		t.Error("expected error for invalid color")
	}
}

// ---------------------------------------------------------------------------
// Base64 renderer
// ---------------------------------------------------------------------------

func TestBase64Renderer_Render(t *testing.T) {
	ctx := context.Background()
	r := NewBase64Renderer()
	data, err := r.Render(ctx, qr, WithWidth(256))
	if err != nil {
		t.Fatalf("Base64 render error: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("Base64 output should not be empty")
	}
	if !strings.HasPrefix(string(data), "data:image/png;base64,") {
		t.Errorf("Base64 output should have data URI prefix, got: %s", string(data[:min(60, len(data))]))
	}
}

func TestBase64Renderer_RenderToString(t *testing.T) {
	ctx := context.Background()
	r := NewBase64Renderer()
	raw, err := r.RenderToString(ctx, qr)
	if err != nil {
		t.Fatalf("RenderToString error: %v", err)
	}
	if raw == "" {
		t.Fatal("empty string")
	}
	if strings.Contains(raw, "data:image/png;base64,") {
		t.Error("RenderToString should NOT contain data URI prefix")
	}
}

func TestBase64Renderer_RenderDataURL(t *testing.T) {
	ctx := context.Background()
	r := NewBase64Renderer()
	url, err := r.RenderDataURL(ctx, qr)
	if err != nil {
		t.Fatalf("RenderDataURL error: %v", err)
	}
	if !strings.HasPrefix(url, "data:image/png;base64,") {
		t.Errorf("expected data URI prefix, got: %s", url[:min(60, len(url))])
	}
}

func TestBase64Renderer_PackageLevelEncode(t *testing.T) {
	encoded, err := Encode(qr)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}
	raw, err := EncodeRaw(qr)
	if err != nil {
		t.Fatalf("EncodeRaw error: %v", err)
	}
	if !strings.HasPrefix(encoded, "data:image/png;base64,") {
		t.Error("Encode should return data URL")
	}
	if strings.Contains(raw, "data:") {
		t.Error("EncodeRaw should not contain data URL prefix")
	}
}

// ---------------------------------------------------------------------------
// Drawing helpers
// ---------------------------------------------------------------------------

func TestDrawModule(t *testing.T) {
	newImg := func(scale int) *image.RGBA {
		total := scale * 10
		return image.NewRGBA(image.Rect(0, 0, total, total))
	}

	fg := color.RGBA{R: 0x11, G: 0x22, B: 0x33, A: 0xFF}
	expand := func(v uint8) uint32 { return uint32(v) | uint32(v)<<8 }
	const scale = 10

	t.Run("nil style defaults to square", func(t *testing.T) {
		img := newImg(scale)
		DrawModule(img, 0, 0, scale, nil, fg)
		got := img.At(scale/2, scale/2)
		r, _, _, _ := got.RGBA()
		if r != expand(0x11) {
			t.Fatalf("center pixel not painted, got=%v", got)
		}
	})

	t.Run("rounded shape non-crash", func(t *testing.T) {
		style := DefaultModuleStyle()
		style.Shape = "rounded"
		style.Roundness = 0.5
		img := newImg(scale)
		DrawModule(img, 5, 5, scale, style, fg)
		painted := false
		for py := 0; py < img.Bounds().Dy(); py++ {
			for px := 0; px < img.Bounds().Dx(); px++ {
				_, _, _, a := img.At(px, py).RGBA()
				if a != 0 {
					painted = true
				}
			}
		}
		if !painted {
			t.Error("rounded module produced an entirely empty image")
		}
	})

	t.Run("circle shape non-crash", func(t *testing.T) {
		style := DefaultModuleStyle()
		style.Shape = "circle"
		img := newImg(scale)
		DrawModule(img, 5, 5, scale, style, fg)
		painted := false
		for py := 0; py < img.Bounds().Dy(); py++ {
			for px := 0; px < img.Bounds().Dx(); px++ {
				_, _, _, a := img.At(px, py).RGBA()
				if a != 0 {
					painted = true
				}
			}
		}
		if !painted {
			t.Error("circle module produced an entirely empty image")
		}
	})

	t.Run("diamond shape non-crash", func(t *testing.T) {
		style := DefaultModuleStyle()
		style.Shape = "diamond"
		img := newImg(scale)
		DrawModule(img, 5, 5, scale, style, fg)
		painted := false
		for py := 0; py < img.Bounds().Dy(); py++ {
			for px := 0; px < img.Bounds().Dx(); px++ {
				_, _, _, a := img.At(px, py).RGBA()
				if a != 0 {
					painted = true
				}
			}
		}
		if !painted {
			t.Error("diamond module produced an entirely empty image")
		}
	})

	t.Run("transparency applied", func(t *testing.T) {
		style := &ModuleStyle{Shape: "square", Transparency: 0.5}
		img := newImg(scale)
		DrawModule(img, 0, 0, scale, style, fg)
		_, _, _, a := img.At(0, 0).RGBA()
		if a == uint32(0xFF<<8) || a == 0 {
			t.Errorf("expected semi-transparent alpha, got a=%d", a)
		}
	})
}

func TestDrawRoundedRect(t *testing.T) {
	c := color.RGBA{R: 0xFF, G: 0x00, B: 0x00, A: 0xFF}
	const s = 20

	t.Run("radius 0 full square", func(t *testing.T) {
		img := image.NewRGBA(image.Rect(0, 0, s, s))
		DrawRoundedRect(img, 0, 0, s, s, 0, c)
		for py := 0; py < s; py++ {
			for px := 0; px < s; px++ {
				_, _, _, a := img.At(px, py).RGBA()
				if a == 0 {
					t.Fatalf("radius-0: pixel (%d,%d) not painted", px, py)
				}
			}
		}
	})

	t.Run("max radius full circle", func(t *testing.T) {
		img := image.NewRGBA(image.Rect(0, 0, s, s))
		DrawRoundedRect(img, 0, 0, s, s, s, c)
		_, _, _, a := img.At(s/2, s/2).RGBA()
		if a == 0 {
			t.Fatal("max-radius: center pixel not painted")
		}
		_, _, _, a = img.At(0, 0).RGBA()
		if a != 0 {
			t.Fatal("max-radius: corner (0,0) should be empty")
		}
	})
}

func TestDrawCircle(t *testing.T) {
	c := color.RGBA{R: 0x00, G: 0xFF, B: 0x00, A: 0xFF}

	t.Run("zero radius", func(t *testing.T) {
		img := image.NewRGBA(image.Rect(0, 0, 10, 10))
		DrawCircle(img, 5, 5, 0, c)
		_, _, _, a := img.At(5, 5).RGBA()
		if a == 0 {
			t.Fatal("zero-radius circle should paint the center pixel")
		}
		_, _, _, a = img.At(6, 5).RGBA()
		if a != 0 {
			t.Fatal("zero-radius circle should not paint adjacent pixels")
		}
	})

	t.Run("normal radius", func(t *testing.T) {
		img := image.NewRGBA(image.Rect(0, 0, 20, 20))
		DrawCircle(img, 10, 10, 5, c)
		_, _, _, a := img.At(10, 10).RGBA()
		if a == 0 {
			t.Fatal("normal circle: center not painted")
		}
	})
}

func TestDrawDiamond(t *testing.T) {
	c := color.RGBA{R: 0x00, G: 0x00, B: 0xFF, A: 0xFF}

	t.Run("zero dimensions early return", func(t *testing.T) {
		img := image.NewRGBA(image.Rect(0, 0, 10, 10))
		DrawDiamond(img, 5, 5, 0, 0, c)
		for py := 0; py < 10; py++ {
			for px := 0; px < 10; px++ {
				_, _, _, a := img.At(px, py).RGBA()
				if a != 0 {
					t.Fatalf("zero-dimension diamond painted pixel (%d,%d)", px, py)
				}
			}
		}
	})

	t.Run("normal dimensions", func(t *testing.T) {
		img := image.NewRGBA(image.Rect(0, 0, 20, 20))
		DrawDiamond(img, 10, 10, 5, 5, c)
		_, _, _, a := img.At(10, 10).RGBA()
		if a == 0 {
			t.Fatal("diamond: center not painted")
		}
	})
}

// ---------------------------------------------------------------------------
// Color interpolation
// ---------------------------------------------------------------------------

func TestInterpolateColor(t *testing.T) {
	start := color.RGBA{R: 0, G: 0, B: 0, A: 255}
	end := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	tests := []struct {
		name string
		t    float64
		want color.RGBA
	}{
		{"t=0 returns start", 0.0, color.RGBA{R: 0, G: 0, B: 0, A: 255}},
		{"t=1 returns end", 1.0, color.RGBA{R: 255, G: 255, B: 255, A: 255}},
		{"t=0.5 midpoint", 0.5, color.RGBA{R: 127, G: 127, B: 127, A: 255}},
		{"t<0 clamped", -1.0, color.RGBA{R: 0, G: 0, B: 0, A: 255}},
		{"t>1 clamped", 2.0, color.RGBA{R: 255, G: 255, B: 255, A: 255}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := InterpolateColor(start, end, tt.t)
			if got != tt.want {
				t.Errorf("InterpolateColor(t=%v) = %v, want %v", tt.t, got, tt.want)
			}
		})
	}
}

func TestGradientColor(t *testing.T) {
	start := color.RGBA{R: 255, G: 0, B: 0, A: 255}
	end := color.RGBA{R: 0, G: 0, B: 255, A: 255}

	t.Run("zero dimensions returns start", func(t *testing.T) {
		got := GradientColor(0, 0, 0, 0, 0, start, end)
		if got != start {
			t.Errorf("zero-dim = %v, want %v", got, start)
		}
	})
	t.Run("angle 0 left redder", func(t *testing.T) {
		left := GradientColor(0, 50, 100, 100, 0, start, end)
		right := GradientColor(99, 50, 100, 100, 0, start, end)
		if left.R <= right.R {
			t.Error("angle 0: left should be redder than right")
		}
	})
}

func TestApplyTransparency(t *testing.T) {
	c := color.RGBA{R: 100, G: 150, B: 200, A: 255}
	tests := []struct {
		name  string
		alpha float64
		wantA uint8
	}{
		{"alpha 0.0", 0.0, 0},
		{"alpha 0.5", 0.5, 127},
		{"alpha 1.0", 1.0, 255},
		{"negative clamped", -0.5, 0},
		{"greater than 1 clamped", 1.5, 255},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ApplyTransparency(c, tt.alpha)
			if got.A != tt.wantA {
				t.Errorf("ApplyTransparency(%v) A=%d, want %d", tt.alpha, got.A, tt.wantA)
			}
		})
	}
}

func TestParseGradient(t *testing.T) {
	t.Run("valid colors", func(t *testing.T) {
		start, end, err := ParseGradient("#FF0000", "#0000FF")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if start != (color.RGBA{R: 255, G: 0, B: 0, A: 255}) {
			t.Errorf("start = %v", start)
		}
		if end != (color.RGBA{R: 0, G: 0, B: 255, A: 255}) {
			t.Errorf("end = %v", end)
		}
	})
	t.Run("invalid start", func(t *testing.T) {
		_, _, err := ParseGradient("bad", "#000000")
		if err == nil {
			t.Error("expected error for invalid start color")
		}
	})
}

// ---------------------------------------------------------------------------
// SVG helpers
// ---------------------------------------------------------------------------

func TestSVGModuleElement(t *testing.T) {
	t.Run("nil style square rect", func(t *testing.T) {
		out := SVGModuleElement(1, 2, 10, nil, "#000000")
		if !strings.Contains(out, "<rect") {
			t.Error("nil style should produce rect tag")
		}
	})
	t.Run("rounded path", func(t *testing.T) {
		style := &ModuleStyle{Shape: "rounded", Roundness: 0.5}
		out := SVGModuleElement(1, 2, 10, style, "#222222")
		if !strings.Contains(out, "<path") {
			t.Error("rounded should produce path tag")
		}
	})
	t.Run("circle", func(t *testing.T) {
		style := &ModuleStyle{Shape: "circle"}
		out := SVGModuleElement(1, 2, 10, style, "#333333")
		if !strings.Contains(out, "<circle") {
			t.Error("circle should produce circle tag")
		}
	})
	t.Run("diamond polygon", func(t *testing.T) {
		style := &ModuleStyle{Shape: "diamond"}
		out := SVGModuleElement(1, 2, 10, style, "#444444")
		if !strings.Contains(out, "<polygon") {
			t.Error("diamond should produce polygon tag")
		}
	})
	t.Run("gradient url fill", func(t *testing.T) {
		style := &ModuleStyle{Shape: "square", GradientEnabled: true}
		out := SVGModuleElement(1, 2, 10, style, "#000000")
		if !strings.Contains(out, "url(#qrgradient)") {
			t.Errorf("expected url(#qrgradient), got: %s", out)
		}
	})
	t.Run("transparency opacity", func(t *testing.T) {
		style := &ModuleStyle{Shape: "square", Transparency: 0.6}
		out := SVGModuleElement(1, 2, 10, style, "#555555")
		if !strings.Contains(out, `opacity="0.60"`) {
			t.Errorf("expected opacity, got: %s", out)
		}
	})
}

func TestSVGGradientDefinition(t *testing.T) {
	out := SVGGradientDefinition("#FF0000", "#00FF00", 135, "my-grad")
	if !strings.Contains(out, `<defs><linearGradient id="my-grad"`) {
		t.Error("missing linearGradient tag with correct id")
	}
	if !strings.Contains(out, `gradientUnits="objectBoundingBox"`) {
		t.Error("missing gradientUnits")
	}
	if !strings.Contains(out, `stop-color="#FF0000"`) {
		t.Error("missing start stop-color")
	}
	if !strings.Contains(out, `stop-color="#00FF00"`) {
		t.Error("missing end stop-color")
	}
}

func TestSVGRoundedRectPath(t *testing.T) {
	t.Run("r=0 no arcs", func(t *testing.T) {
		out := SVGRoundedRectPath(10, 10, 100, 100, 0)
		if strings.Contains(out, "A") {
			t.Error("radius 0 should not contain arc commands")
		}
	})
	t.Run("normal radius 4 arcs", func(t *testing.T) {
		out := SVGRoundedRectPath(0, 0, 100, 100, 20)
		if !strings.Contains(out, "A") {
			t.Error("rounded path should contain arc commands")
		}
		arcCount := strings.Count(out, " A")
		if arcCount != 4 {
			t.Errorf("expected 4 arcs, got %d", arcCount)
		}
	})
	t.Run("negative radius clamped", func(t *testing.T) {
		out := SVGRoundedRectPath(0, 0, 100, 100, -10)
		if strings.Contains(out, "A") {
			t.Error("negative radius should be clamped to 0")
		}
	})
}

// ---------------------------------------------------------------------------
// All renderers — cancelled context safety
// ---------------------------------------------------------------------------

func TestRenderers_CancelledContextNoPanic(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	for _, f := range []Format{FormatPNG, FormatSVG, FormatTerminal, FormatPDF, FormatBase64} {
		r, err := GetRenderer(f)
		if err != nil {
			continue
		}
		t.Run(f.String(), func(t *testing.T) {
			// Should not panic — may or may not return error depending on context check timing
			data, _ := r.Render(ctx, qr)
			_ = data
		})
	}
}
