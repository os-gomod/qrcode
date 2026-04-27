package logo

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"testing"
)

func TestSupportedFormats(t *testing.T) {
	formats := SupportedFormats()
	if len(formats) == 0 {
		t.Error("should return supported formats")
	}
	for _, f := range formats {
		if f == "" {
			t.Error("format should not be empty")
		}
	}
}

func TestIsSupportedFormat(t *testing.T) {
	tests := []struct {
		ext  string
		want bool
	}{
		{".png", true},
		{".jpg", true},
		{".jpeg", true},
		{".gif", true},
		{".PNG", true},
		{".bmp", false},
		{"", false},
		{"png", false}, // no dot
	}
	for _, tt := range tests {
		got := IsSupportedFormat(tt.ext)
		if got != tt.want {
			t.Errorf("IsSupportedFormat(%q) = %v, want %v", tt.ext, got, tt.want)
		}
	}
}

func TestNew(t *testing.T) {
	p := New("logo.png", 0.25)
	if p == nil {
		t.Fatal("New() returned nil")
	}
	if p.Source() != "logo.png" {
		t.Errorf("Source() = %q, want 'logo.png'", p.Source())
	}
	if p.SizeRatio() != 0.25 {
		t.Errorf("SizeRatio() = %f, want 0.25", p.SizeRatio())
	}
}

func TestWithTint(t *testing.T) {
	p := New("logo.png", 0.25)
	p2 := p.WithTint(color.RGBA{R: 255, G: 0, B: 0, A: 255})
	if p2 != p {
		t.Error("WithTint should return same pointer for chaining")
	}
}

func TestLoad_Empty(t *testing.T) {
	p := New("", 0.25)
	_, err := p.Load()
	if err == nil {
		t.Error("expected error for empty source")
	}
}

func TestLoad_NonExistent(t *testing.T) {
	p := New("/nonexistent/file.png", 0.25)
	_, err := p.Load()
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestLoadFromBytes_Empty(t *testing.T) {
	_, err := LoadFromBytes(nil)
	if err == nil {
		t.Error("expected error for empty data")
	}
	_, err = LoadFromBytes([]byte{})
	if err == nil {
		t.Error("expected error for empty byte slice")
	}
}

func TestLoadFromBytes_Invalid(t *testing.T) {
	_, err := LoadFromBytes([]byte("not an image"))
	if err == nil {
		t.Error("expected error for invalid image data")
	}
}

func TestLoadFromReader_Invalid(t *testing.T) {
	_, err := LoadFromReader(bytes.NewReader([]byte("not an image")))
	if err == nil {
		t.Error("expected error for invalid reader data")
	}
}

func TestResizeLogo(t *testing.T) {
	img := createTestImage(100, 100)
	resized := ResizeLogo(img, 21, 0.5)
	if resized == nil {
		t.Fatal("ResizeLogo returned nil")
	}
	bounds := resized.Bounds()
	if bounds.Dx() < 1 || bounds.Dy() < 1 {
		t.Error("resized image should have positive dimensions")
	}
}

func TestResizeLogo_SmallRatio(t *testing.T) {
	img := createTestImage(100, 100)
	resized := ResizeLogo(img, 21, 0.01)
	if resized == nil {
		t.Fatal("ResizeLogo with tiny ratio returned nil")
	}
	bounds := resized.Bounds()
	if bounds.Dx() < 1 || bounds.Dy() < 1 {
		t.Error("minimum size should be 1x1")
	}
}

func TestResizeLogoToPixels(t *testing.T) {
	img := createTestImage(200, 100)
	resized := ResizeLogoToPixels(img, 50, 25)
	if resized == nil {
		t.Fatal("ResizeLogoToPixels returned nil")
	}
	bounds := resized.Bounds()
	if bounds.Dx() != 50 || bounds.Dy() != 25 {
		t.Errorf("got %dx%d, want 50x25", bounds.Dx(), bounds.Dy())
	}
}

func TestResizeLogoToPixels_MinSize(t *testing.T) {
	img := createTestImage(10, 10)
	resized := ResizeLogoToPixels(img, 0, 0)
	bounds := resized.Bounds()
	if bounds.Dx() != 1 || bounds.Dy() != 1 {
		t.Errorf("min size should be 1x1, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestTintLogo_NilTint(t *testing.T) {
	img := createTestImage(10, 10)
	result := TintLogo(img, nil)
	if result == nil {
		t.Fatal("TintLogo with nil tint returned nil")
	}
}

func TestTintLogo_Color(t *testing.T) {
	img := createTestImage(10, 10)
	tint := color.RGBA{R: 255, G: 0, B: 0, A: 255}
	result := TintLogo(img, tint)
	if result == nil {
		t.Fatal("TintLogo returned nil")
	}
	bounds := result.Bounds()
	if bounds.Dx() != 10 || bounds.Dy() != 10 {
		t.Errorf("tinted image should preserve size, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestOverlayLogo(t *testing.T) {
	qrImg := createTestImage(100, 100)
	logoImg := createTestImage(20, 20)
	result := OverlayLogo(qrImg, logoImg, 0)
	if result == nil {
		t.Fatal("OverlayLogo returned nil")
	}
	bounds := result.Bounds()
	if bounds.Dx() != 100 || bounds.Dy() != 100 {
		t.Errorf("overlay should preserve QR image size, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestCloneToRGBA(t *testing.T) {
	img := createTestImage(10, 10)
	clone := CloneToRGBA(img)
	if clone == nil {
		t.Fatal("CloneToRGBA returned nil")
	}
	bounds := clone.Bounds()
	if bounds.Dx() != 10 || bounds.Dy() != 10 {
		t.Errorf("clone should have same size, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestEncodePNG(t *testing.T) {
	img := createTestImage(10, 10)
	data, err := EncodePNG(img)
	if err != nil {
		t.Fatalf("EncodePNG error: %v", err)
	}
	if len(data) == 0 {
		t.Error("encoded PNG should not be empty")
	}
	// Verify PNG magic bytes.
	if len(data) >= 4 && data[0] != 0x89 {
		t.Error("output doesn't start with PNG magic bytes")
	}
}

func TestValidate(t *testing.T) {
	t.Run("empty path", func(t *testing.T) {
		if err := Validate(""); err == nil {
			t.Error("expected error for empty path")
		}
	})
	t.Run("unsupported format", func(t *testing.T) {
		if err := Validate("test.bmp"); err == nil {
			t.Error("expected error for unsupported format")
		}
	})
	t.Run("non-existent file", func(t *testing.T) {
		if err := Validate("/nonexistent/file.png"); err == nil {
			t.Error("expected error for non-existent file")
		}
	})
	t.Run("valid file", func(t *testing.T) {
		// Create a temp PNG file.
		img := createTestImage(5, 5)
		var buf bytes.Buffer
		_ = png.Encode(&buf, img)
		tmpFile := t.TempDir() + "/test.png"
		if err := os.WriteFile(tmpFile, buf.Bytes(), 0o644); err != nil {
			t.Fatalf("failed to write temp file: %v", err)
		}
		if err := Validate(tmpFile); err != nil {
			t.Errorf("valid PNG should pass: %v", err)
		}
	})
}

func TestLogoSize(t *testing.T) {
	size := LogoSize(300, 0.25)
	if size != 75 {
		t.Errorf("LogoSize(300, 0.25) = %d, want 75", size)
	}
	size2 := LogoSize(10, 0.01)
	if size2 != 1 {
		t.Errorf("LogoSize(10, 0.01) = %d, want 1 (min)", size2)
	}
}

// createTestImage creates a simple test image with the given dimensions.
func createTestImage(w, h int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x), G: uint8(y), B: 128, A: 255})
		}
	}
	return img
}
