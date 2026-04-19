package renderer

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/os-gomod/qrcode/encoding"
)

func TestPNGRenderer(t *testing.T) {
	r := NewPNGRenderer()
	if r.Type() != "png" {
		t.Errorf("Type() = %q, want %q", r.Type(), "png")
	}
	if r.ContentType() != "image/png" {
		t.Errorf("ContentType() = %q", r.ContentType())
	}

	qr, _ := encoding.Encode([]byte("test"), encoding.ECLevelM, 0)
	var buf bytes.Buffer
	err := r.Render(context.Background(), qr, &buf)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("Render should produce output")
	}
}

func TestPNGRendererWithOptions(t *testing.T) {
	r := NewPNGRenderer()
	qr, _ := encoding.Encode([]byte("test"), encoding.ECLevelM, 0)

	var buf bytes.Buffer
	err := r.Render(context.Background(), qr, &buf,
		WithWidth(100), WithHeight(100), WithQuietZone(2),
		WithForegroundColor("#FF0000"), WithBackgroundColor("#0000FF"))
	if err != nil {
		t.Fatalf("Render with options error: %v", err)
	}
}

func TestPNGRendererAdvancedModules(t *testing.T) {
	r := NewPNGRenderer()
	qr, _ := encoding.Encode([]byte("test"), encoding.ECLevelM, 0)
	style := DefaultModuleStyle()
	style.Shape = "circle"

	var buf bytes.Buffer
	err := r.Render(context.Background(), qr, &buf, WithModuleStyle(style))
	if err != nil {
		t.Fatalf("Render with circle modules error: %v", err)
	}
}

func TestPNGRendererGradient(t *testing.T) {
	r := NewPNGRenderer()
	qr, _ := encoding.Encode([]byte("test"), encoding.ECLevelM, 0)
	style := DefaultModuleStyle()
	style.GradientEnabled = true
	style.GradientStart = "#000000"
	style.GradientEnd = "#FFFFFF"

	var buf bytes.Buffer
	err := r.Render(context.Background(), qr, &buf, WithModuleStyle(style))
	if err != nil {
		t.Fatalf("Render with gradient error: %v", err)
	}
}

func TestPNGRendererInvalidColor(t *testing.T) {
	r := NewPNGRenderer()
	qr, _ := encoding.Encode([]byte("test"), encoding.ECLevelM, 0)

	var buf bytes.Buffer
	err := r.Render(context.Background(), qr, &buf, WithForegroundColor("invalid"))
	if err == nil {
		t.Error("expected error for invalid color")
	}
}

func TestSVGRenderer(t *testing.T) {
	r := NewSVGRenderer()
	if r.Type() != "svg" {
		t.Errorf("Type() = %q, want %q", r.Type(), "svg")
	}

	qr, _ := encoding.Encode([]byte("test"), encoding.ECLevelM, 0)
	var buf bytes.Buffer
	err := r.Render(context.Background(), qr, &buf)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	svg := buf.String()
	if !strings.Contains(svg, "<svg") {
		t.Error("SVG should contain <svg> tag")
	}
	if !strings.Contains(svg, "rect") {
		t.Error("SVG should contain rect elements")
	}
}

func TestTerminalRenderer(t *testing.T) {
	r := NewTerminalRenderer()
	if r.Type() != "terminal" {
		t.Errorf("Type() = %q, want %q", r.Type(), "terminal")
	}

	qr, _ := encoding.Encode([]byte("test"), encoding.ECLevelM, 0)
	var buf bytes.Buffer
	err := r.Render(context.Background(), qr, &buf)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
}

func TestPDFRenderer(t *testing.T) {
	r := NewPDFRenderer()
	if r.Type() != "pdf" {
		t.Errorf("Type() = %q, want %q", r.Type(), "pdf")
	}

	qr, _ := encoding.Encode([]byte("test"), encoding.ECLevelM, 0)
	var buf bytes.Buffer
	err := r.Render(context.Background(), qr, &buf)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	pdf := buf.String()
	if !strings.Contains(pdf, "%PDF") {
		t.Error("PDF should contain %PDF header")
	}
}

func TestBase64Renderer(t *testing.T) {
	r := NewBase64Renderer()
	if r.Type() != "base64" {
		t.Errorf("Type() = %q, want %q", r.Type(), "base64")
	}

	qr, _ := encoding.Encode([]byte("test"), encoding.ECLevelM, 0)
	b64, err := r.RenderToString(context.Background(), qr)
	if err != nil {
		t.Fatalf("RenderToString error: %v", err)
	}
	if !strings.HasPrefix(b64, "iVBOR") {
		t.Error("base64 output should start with PNG header")
	}

	dataURL, err := r.RenderDataURL(context.Background(), qr)
	if err != nil {
		t.Fatalf("RenderDataURL error: %v", err)
	}
	if !strings.HasPrefix(dataURL, "data:image/png;base64,") {
		t.Error("data URL should have correct prefix")
	}
}

func TestColorComponents(t *testing.T) {
	tests := []struct {
		hex   string
		wantR int
		wantG int
		wantB int
	}{
		{"#FF0000", 255, 0, 0},
		{"#000000", 0, 0, 0},
		{"#123456", 18, 52, 86},
		{"invalid", 0, 0, 0},
	}
	for _, tt := range tests {
		r, g, b := colorComponents(tt.hex)
		if tt.hex == "invalid" {
			continue
		}
		if r != tt.wantR || g != tt.wantG || b != tt.wantB {
			t.Errorf("colorComponents(%q) = %d,%d,%d, want %d,%d,%d", tt.hex, r, g, b, tt.wantR, tt.wantG, tt.wantB)
		}
	}
}
