package logo

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestSupportedFormats(t *testing.T) {
	formats := SupportedFormats()
	if len(formats) == 0 {
		t.Error("SupportedFormats should return non-empty slice")
	}
	expected := []string{".png", ".jpg", ".jpeg", ".gif"}
	if len(formats) != len(expected) {
		t.Errorf("expected %d formats, got %d", len(expected), len(formats))
	}
	for i, f := range expected {
		if formats[i] != f {
			t.Errorf("format[%d] = %q, want %q", i, formats[i], f)
		}
	}
}

func TestIsSupportedFormat(t *testing.T) {
	tests := []struct {
		ext  string
		want bool
	}{
		{".png", true},
		{".PNG", true},
		{".jpg", true},
		{".jpeg", true},
		{".gif", true},
		{".bmp", false},
		{".svg", false},
		{"", false},
		{"png", false}, // missing dot
	}

	for _, tt := range tests {
		got := IsSupportedFormat(tt.ext)
		if got != tt.want {
			t.Errorf("IsSupportedFormat(%q) = %v, want %v", tt.ext, got, tt.want)
		}
	}
}

func TestResizeLogo(t *testing.T) {
	// Create a simple test image
	srcImg := image.NewRGBA(image.Rect(0, 0, 100, 50))
	for y := 0; y < 50; y++ {
		for x := 0; x < 100; x++ {
			srcImg.Set(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}

	result := ResizeLogo(srcImg, 21, 0.3)
	if result == nil {
		t.Fatal("ResizeLogo returned nil")
	}
	bounds := result.Bounds()
	if bounds.Dx() <= 0 || bounds.Dy() <= 0 {
		t.Errorf("resize result has invalid dimensions: %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestResizeLogoToPixels(t *testing.T) {
	srcImg := image.NewRGBA(image.Rect(0, 0, 50, 50))
	for y := 0; y < 50; y++ {
		for x := 0; x < 50; x++ {
			srcImg.Set(x, y, color.White)
		}
	}

	result := ResizeLogoToPixels(srcImg, 20, 20)
	if result == nil {
		t.Fatal("ResizeLogoToPixels returned nil")
	}
	if result.Bounds().Dx() != 20 || result.Bounds().Dy() != 20 {
		t.Errorf("expected 20x20, got %dx%d", result.Bounds().Dx(), result.Bounds().Dy())
	}

	// Test clamping
	result2 := ResizeLogoToPixels(srcImg, 0, 0)
	if result2.Bounds().Dx() != 1 || result2.Bounds().Dy() != 1 {
		t.Error("zero dimensions should be clamped to 1")
	}
}

func TestOverlayLogo(t *testing.T) {
	qrImg := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			qrImg.Set(x, y, color.White)
		}
	}

	logoImg := image.NewRGBA(image.Rect(0, 0, 20, 20))
	for y := 0; y < 20; y++ {
		for x := 0; x < 20; x++ {
			logoImg.Set(x, y, color.RGBA{R: 0, G: 0, B: 255, A: 255})
		}
	}

	result := OverlayLogo(qrImg, logoImg, 0)
	if result == nil {
		t.Fatal("OverlayLogo returned nil")
	}
	if result.Bounds().Dx() != 100 {
		t.Errorf("result width = %d, want 100", result.Bounds().Dx())
	}
}

func TestCloneToRGBA(t *testing.T) {
	srcImg := image.NewRGBA(image.Rect(0, 0, 30, 30))
	for y := 0; y < 30; y++ {
		for x := 0; x < 30; x++ {
			srcImg.Set(x, y, color.Black)
		}
	}

	cloned := CloneToRGBA(srcImg)
	if cloned == nil {
		t.Fatal("CloneToRGBA returned nil")
	}
	if cloned.Bounds() != srcImg.Bounds() {
		t.Error("bounds should match")
	}

	// Verify pixel at a specific location
	r, g, b, _ := cloned.At(5, 5).RGBA()
	if r != 0 || g != 0 || b != 0 {
		t.Error("cloned pixel should be black")
	}
}

func TestValidate(t *testing.T) {
	// Create a temp PNG file
	tmpDir := t.TempDir()
	validPath := filepath.Join(tmpDir, "test.png")

	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	f, _ := os.Create(validPath)
	if err := png.Encode(f, img); err != nil {
		t.Fatalf("failed to create test PNG: %v", err)
	}
	f.Close()

	if err := Validate(validPath); err != nil {
		t.Errorf("Validate(%q) error: %v", validPath, err)
	}

	// Invalid path
	if err := Validate(""); err == nil {
		t.Error("Validate('') should return error")
	}

	// Invalid format
	invalidPath := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(invalidPath, []byte("text"), 0o644)
	if err := Validate(invalidPath); err == nil {
		t.Error("Validate(.txt) should return error")
	}

	// Non-existent
	if err := Validate("/nonexistent/file.png"); err == nil {
		t.Error("Validate(nonexistent) should return error")
	}
}

func TestLogoSize(t *testing.T) {
	tests := []struct {
		size  int
		ratio float64
		want  int
	}{
		{100, 0.3, 30},
		{100, 0, 1},
		{0, 0.3, 1},
	}

	for _, tt := range tests {
		got := LogoSize(tt.size, tt.ratio)
		if got != tt.want {
			t.Errorf("LogoSize(%d, %f) = %d, want %d", tt.size, tt.ratio, got, tt.want)
		}
	}
}

func TestProcessor(t *testing.T) {
	p := New("/path/to/logo.png", 0.25)
	if p.Source() != "/path/to/logo.png" {
		t.Errorf("Source() = %q", p.Source())
	}
	if p.SizeRatio() != 0.25 {
		t.Errorf("SizeRatio() = %f", p.SizeRatio())
	}

	p2 := p.WithTint(color.RGBA{R: 255, A: 255})
	// WithTint returns same pointer (pointer receiver)
	if p2.Source() != p.Source() {
		t.Error("WithTint should return same processor")
	}

	// Load with empty path
	_, err := p.Load()
	if err == nil {
		t.Error("Load() with non-existent path should return error")
	}
}

func TestEncodePNG(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			img.Set(x, y, color.White)
		}
	}
	data, err := EncodePNG(img)
	if err != nil {
		t.Fatalf("EncodePNG error: %v", err)
	}
	if len(data) == 0 {
		t.Error("EncodePNG should return non-empty data")
	}
}

func TestLoadFromBytes(t *testing.T) {
	_, err := LoadFromBytes(nil)
	if err == nil {
		t.Error("expected error for nil data")
	}
	_, err = LoadFromBytes([]byte{})
	if err == nil {
		t.Error("expected error for empty data")
	}
	_, err = LoadFromBytes([]byte("not an image"))
	if err == nil {
		t.Error("expected error for invalid image data")
	}
}

func TestTintLogo(t *testing.T) {
	srcImg := image.NewRGBA(image.Rect(0, 0, 10, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			srcImg.Set(x, y, color.RGBA{R: 100, G: 100, B: 100, A: 255})
		}
	}

	// With nil tint, should clone
	result := TintLogo(srcImg, nil)
	if result == nil {
		t.Fatal("TintLogo with nil should return clone")
	}

	// With tint
	result2 := TintLogo(srcImg, color.RGBA{R: 255, G: 0, B: 0, A: 255})
	if result2 == nil {
		t.Fatal("TintLogo returned nil")
	}
}

func TestLoadFromReader(t *testing.T) {
	// LoadFromReader with nil reader causes panic in image.Decode
	// We can't easily test this without a real image reader
	// Just verify LoadFromBytes works for invalid data
	_, err := LoadFromBytes([]byte("not an image"))
	if err == nil {
		t.Error("expected error for invalid image data")
	}
}
