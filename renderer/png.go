package renderer

import (
	"context"
	"image"
	"image/color"
	"image/png"
	"io"

	"github.com/os-gomod/qrcode/encoding"
)

// PNGRenderer renders QR codes as PNG raster images.
//
// The renderer supports the full range of module styles including
// rounded rectangles, circles, diamonds, linear gradient fills, and
// transparency (alpha blending). When a non-square ModuleStyle or
// gradient is configured, the renderer delegates to the advanced
// drawing functions in advanced_png.go (DrawModule, DrawRoundedRect,
// DrawCircle, DrawDiamond, etc.).
//
// For plain square modules with no gradient or transparency, a fast
// pixel-fill path is used instead.
//
// Example:
//
//	r := renderer.NewPNGRenderer()
//	err := r.Render(ctx, qr, os.Stdout,
//	    renderer.WithWidth(512),
//	    renderer.WithGradient("#4A00E0", "#8E2DE2", 135),
//	    renderer.WithRoundedModules(0.5),
//	)
type PNGRenderer struct{}

// NewPNGRenderer creates a new PNGRenderer. The returned renderer is
// stateless and safe for concurrent use.
func NewPNGRenderer() *PNGRenderer {
	return &PNGRenderer{}
}

// Type returns the format identifier "png".
func (r *PNGRenderer) Type() string { return "png" }

// ContentType returns the MIME type "image/png".
func (r *PNGRenderer) ContentType() string { return "image/png" }

// Render writes the QR code as a PNG image to w using the given render options.
//
// The output dimensions are determined by the Width option: a pixel scale
// factor is computed as Width / (qr.Size + 2*QuietZone) and applied uniformly
// to all modules. The resulting image is always square.
//
// When ModuleStyle is set with a non-square shape, gradient, or transparency,
// the advanced drawing path is activated. Otherwise, a fast path fills square
// modules directly.
func (r *PNGRenderer) Render(_ context.Context, qr *encoding.QRCode, w io.Writer, opts ...RenderOption) error {
	cfg := ApplyOptions(opts...)
	fgR, fgG, fgB, err := ParseHexColor(cfg.ForegroundColor)
	if err != nil {
		return err
	}
	bgR, bgG, bgB, err := ParseHexColor(cfg.BackgroundColor)
	if err != nil {
		return err
	}
	scale := ScaleSize(qr.Size, cfg.QuietZone, cfg.Width)
	totalModules := qr.Size + 2*cfg.QuietZone
	imgSize := totalModules * scale
	img := image.NewRGBA(image.Rect(0, 0, imgSize, imgSize))
	bg := color.RGBA{R: bgR, G: bgG, B: bgB, A: 255}
	for y := 0; y < imgSize; y++ {
		for x := 0; x < imgSize; x++ {
			img.Set(x, y, bg)
		}
	}
	fg := color.RGBA{R: fgR, G: fgG, B: fgB, A: 255}
	useAdvanced := cfg.ModuleStyle != nil &&
		(cfg.ModuleStyle.Shape != "square" || cfg.ModuleStyle.GradientEnabled || cfg.ModuleStyle.Transparency < 1.0)
	if useAdvanced {
		if validateErr := cfg.ModuleStyle.Validate(); validateErr != nil {
			return err
		}
		for row := 0; row < qr.Size; row++ {
			for col := 0; col < qr.Size; col++ {
				if qr.Modules[row][col] {
					originX := (col + cfg.QuietZone) * scale
					originY := (row + cfg.QuietZone) * scale
					DrawModule(img, originX, originY, scale, cfg.ModuleStyle, fg)
				}
			}
		}
	} else {
		for row := 0; row < qr.Size; row++ {
			for col := 0; col < qr.Size; col++ {
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
	if encErr := png.Encode(w, img); encErr != nil {
		return err
	}
	return nil
}
