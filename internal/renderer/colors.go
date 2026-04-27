package renderer

import (
	"fmt"
	"strconv"
)

//nolint:revive // RGB components follow Go convention (R, G, B, err)
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

// ScaleSize computes the pixel size of each module given the matrix dimensions,
// quiet zone, and target pixel size.
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

//nolint:revive // RGB components follow Go convention (R, G, B)
func colorToInt(hex string) (int, int, int) {
	r, g, b, err := ParseHexColor(hex)
	if err != nil {
		return 0, 0, 0
	}
	return int(r), int(g), int(b)
}
