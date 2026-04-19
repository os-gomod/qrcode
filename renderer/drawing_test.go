package renderer

import (
	"image"
	"image/color"
	"testing"
)

func TestDrawCircle(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 20, 20))
	DrawCircle(img, 10, 10, 5, color.RGBA{R: 255, A: 255})

	// Center should be drawn
	r, _, _, _ := img.At(10, 10).RGBA()
	if r == 0 {
		t.Error("center pixel should be drawn")
	}

	// Far corner should not be drawn
	r, _, _, _ = img.At(0, 0).RGBA()
	if r != 0 {
		t.Error("corner should not be drawn")
	}
}

func TestDrawDiamond(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 20, 20))
	DrawDiamond(img, 10, 10, 5, 5, color.RGBA{R: 255, A: 255})

	// Center should be drawn
	r, _, _, _ := img.At(10, 10).RGBA()
	if r == 0 {
		t.Error("center pixel should be drawn")
	}

	// Zero size should not draw
	img2 := image.NewRGBA(image.Rect(0, 0, 20, 20))
	DrawDiamond(img2, 10, 10, 0, 5, color.White) // no-op
	DrawDiamond(img2, 10, 10, 5, 0, color.White) // no-op
}

func TestInterpolateColor(t *testing.T) {
	black := color.RGBA{R: 0, G: 0, B: 0, A: 255}
	white := color.RGBA{R: 255, G: 255, B: 255, A: 255}

	// t=0 should return start
	result := InterpolateColor(black, white, 0.0)
	if result.R != 0 {
		t.Errorf("InterpolateColor t=0: R = %d, want 0", result.R)
	}

	// t=1 should return end
	result = InterpolateColor(black, white, 1.0)
	if result.R != 255 {
		t.Errorf("InterpolateColor t=1: R = %d, want 255", result.R)
	}

	// t=0.5 should be midpoint
	result = InterpolateColor(black, white, 0.5)
	if result.R != 127 && result.R != 128 {
		t.Errorf("InterpolateColor t=0.5: R = %d, want ~128", result.R)
	}

	// Clamp below 0
	result = InterpolateColor(black, white, -1.0)
	if result.R != 0 {
		t.Errorf("InterpolateColor t=-1: R = %d, want 0", result.R)
	}

	// Clamp above 1
	result = InterpolateColor(black, white, 2.0)
	if result.R != 255 {
		t.Errorf("InterpolateColor t=2: R = %d, want 255", result.R)
	}
}

func TestApplyTransparency(t *testing.T) {
	white := color.RGBA{R: 255, G: 255, B: 255, A: 255}

	result := ApplyTransparency(white, 0.5)
	if result.A != 127 && result.A != 128 {
		t.Errorf("ApplyTransparency alpha = %d, want ~128", result.A)
	}
	if result.R != 255 {
		t.Errorf("ApplyTransparency R = %d, want 255", result.R)
	}

	// Full transparency
	result = ApplyTransparency(white, 0.0)
	if result.A != 0 {
		t.Errorf("ApplyTransparency alpha 0.0: A = %d, want 0", result.A)
	}

	// Full opaque
	result = ApplyTransparency(white, 1.0)
	if result.A != 255 {
		t.Errorf("ApplyTransparency alpha 1.0: A = %d, want 255", result.A)
	}

	// Clamp
	result = ApplyTransparency(white, -1.0)
	if result.A != 0 {
		t.Errorf("ApplyTransparency alpha -1: A = %d", result.A)
	}
	result = ApplyTransparency(white, 2.0)
	if result.A != 255 {
		t.Errorf("ApplyTransparency alpha 2: A = %d", result.A)
	}
}

func TestParseGradient(t *testing.T) {
	start, end, err := ParseGradient("#FF0000", "#0000FF")
	if err != nil {
		t.Fatalf("ParseGradient error: %v", err)
	}
	if start.R != 255 {
		t.Errorf("start R = %d, want 255", start.R)
	}
	if end.B != 255 {
		t.Errorf("end B = %d, want 255", end.B)
	}

	// Invalid start
	_, _, err = ParseGradient("invalid", "#0000FF")
	if err == nil {
		t.Error("expected error for invalid start color")
	}

	// Invalid end
	_, _, err = ParseGradient("#FF0000", "invalid")
	if err == nil {
		t.Error("expected error for invalid end color")
	}
}

func TestGradientColor(t *testing.T) {
	start := color.RGBA{R: 255, G: 0, B: 0, A: 255}
	end := color.RGBA{R: 0, G: 0, B: 255, A: 255}

	// Zero dimensions should return start
	result := GradientColor(0, 0, 0, 0, 0, start, end)
	if result.R != 255 {
		t.Errorf("zero dim: R = %d, want 255", result.R)
	}

	// Normal dimensions
	result = GradientColor(50, 50, 100, 100, 0, start, end)
	// Just verify no panic and we get something
	if result.R == 0 && result.G == 0 && result.B == 0 {
		t.Error("GradientColor returned unexpected black")
	}
}

func TestGradientColorAt(t *testing.T) {
	start := color.RGBA{R: 0, G: 0, B: 0, A: 255}
	end := color.RGBA{R: 255, G: 255, B: 255, A: 255}

	// Zero dimensions should return start
	result := GradientColorAt(0, 0, 0, 0, 0, start, end)
	if result.R != 0 {
		t.Errorf("zero dim: R = %d, want 0", result.R)
	}

	// Normal
	result = GradientColorAt(50, 50, 100, 100, 45, start, end)
	_ = result
}

func TestClamp01(t *testing.T) {
	tests := []struct {
		in   float64
		want float64
	}{
		{0.5, 0.5},
		{0.0, 0.0},
		{1.0, 1.0},
		{-1.0, 0.0},
		{2.0, 1.0},
	}

	for _, tt := range tests {
		got := clamp01(tt.in)
		if got != tt.want {
			t.Errorf("clamp01(%f) = %f, want %f", tt.in, got, tt.want)
		}
	}
}

func TestAbsInt(t *testing.T) {
	if absInt(5) != 5 {
		t.Error("absInt(5) != 5")
	}
	if absInt(-5) != 5 {
		t.Error("absInt(-5) != 5")
	}
	if absInt(0) != 0 {
		t.Error("absInt(0) != 0")
	}
}
