package renderer

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/png"

	"github.com/os-gomod/qrcode/v2/internal/encoding"
)

// PNGRenderer renders QR codes as PNG images.
// It is safe for concurrent use.
type PNGRenderer struct{}

// NewPNGRenderer creates a new PNGRenderer.
func NewPNGRenderer() *PNGRenderer {
	return &PNGRenderer{}
}

// Render encodes the QR matrix as PNG image bytes.
//
//nolint:gocyclo,cyclop // rendering pipeline has multiple configuration paths
func (*PNGRenderer) Render(ctx context.Context, qr *encoding.QRCode, opts ...RenderOption) ([]byte, error) {
	cfg := ApplyOptions(opts...)

	fgR, fgG, fgB, err := ParseHexColor(cfg.ForegroundColor)
	if err != nil {
		return nil, err
	}
	bgR, bgG, bgB, err := ParseHexColor(cfg.BackgroundColor)
	if err != nil {
		return nil, err
	}

	scale := ScaleSize(qr.Size, cfg.QuietZone, cfg.Width)
	totalModules := qr.Size + 2*cfg.QuietZone
	imgSize := totalModules * scale

	img := image.NewRGBA(image.Rect(0, 0, imgSize, imgSize))

	// Fill background.
	bg := color.RGBA{R: bgR, G: bgG, B: bgB, A: 255}
	for y := range imgSize {
		for x := range imgSize {
			img.Set(x, y, bg)
		}
	}

	fg := color.RGBA{R: fgR, G: fgG, B: fgB, A: 255}

	if cfg.ModuleStyle.UseAdvanced() {
		if validateErr := cfg.ModuleStyle.Validate(); validateErr != nil {
			return nil, validateErr
		}
		for row := range qr.Size {
			for col := range qr.Size {
				if qr.Modules[row][col] {
					originX := (col + cfg.QuietZone) * scale
					originY := (row + cfg.QuietZone) * scale
					DrawModule(img, originX, originY, scale, cfg.ModuleStyle, fg)
				}
			}
		}
	} else {
		for row := range qr.Size {
			for col := range qr.Size {
				if qr.Modules[row][col] {
					originX := (col + cfg.QuietZone) * scale
					originY := (row + cfg.QuietZone) * scale
					for py := originY; py < originY+scale; py++ {
						for px := originX; px < originX+scale; px++ {
							img.Set(px, py, fg)
						}
					}
				}
			}
		}
	}

	var buf bytes.Buffer
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	if encodeErr := png.Encode(&buf, img); encodeErr != nil {
		return nil, fmt.Errorf("png encode failed: %w", encodeErr)
	}
	return buf.Bytes(), nil
}
