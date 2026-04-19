package renderer

import (
	"image/color"
	"math"
)

// GradientColorAt computes the linear gradient color at pixel position (px, py)
// for an image of the given dimensions and gradient angle.
//
// The interpolation is based on projecting the point onto the gradient
// direction vector (defined by angle in degrees) and normalising against
// the maximum possible projection distance across the image. The result
// is clamped to [0.0, 1.0] before interpolating between start and end colors.
//
// This function differs from GradientColor in that it uses a centered
// projection for more uniform gradient distribution across the image.
//
// Parameters:
//   - px, py:   pixel coordinates
//   - width, height: image dimensions
//   - angle:    gradient angle in degrees (0° = right, 90° = down)
//   - start:    gradient start color
//   - end:      gradient end color
func GradientColorAt(px, py, width, height int, angle float64, start, end color.RGBA) color.RGBA {
	angleRad := angle * math.Pi / 180
	cx := float64(width) / 2
	cy := float64(height) / 2
	dx := float64(px) - cx
	dy := float64(py) - cy
	projection := dx*math.Cos(angleRad) + dy*math.Sin(angleRad)
	maxDist := math.Abs(float64(width)*math.Cos(angleRad)) + math.Abs(float64(height)*math.Sin(angleRad))
	if maxDist == 0 {
		return start
	}
	t := clamp01(projection/maxDist + 0.5)
	return InterpolateColor(start, end, t)
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
