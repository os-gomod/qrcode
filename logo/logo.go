// Package logo provides image loading, resizing, tinting, and overlay compositing
// for QR code center logo images.
//
// The package supports PNG, JPEG, and GIF formats and exposes both a
// [Processor] type for fluent logo manipulation and standalone functions
// for individual operations.
//
// # Typical Pipeline
//
//	proc := logo.New("logo.png", 0.25).WithTint(color.RGBA{0, 0, 0, 255})
//	img, err := proc.Load()
//	resized := logo.ResizeLogo(img, qrModules, 0.25)
//	tinted := logo.TintLogo(resized, color.RGBA{0, 0, 0, 255})
//	final := logo.OverlayLogo(qrImage, tinted, qrModules)
package logo

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"

	// Image format decoders registered via init().
	_ "image/gif"
	_ "image/jpeg"
)

// SupportedFormats returns the list of supported image file extensions
// (including the leading dot).
func SupportedFormats() []string {
	return []string{".png", ".jpg", ".jpeg", ".gif"}
}

// IsSupportedFormat reports whether the file extension (with or without a
// leading dot) is a supported image format. The check is case-insensitive.
func IsSupportedFormat(ext string) bool {
	lower := strings.ToLower(ext)
	for _, f := range SupportedFormats() {
		if lower == f {
			return true
		}
	}
	return false
}

// Processor loads and manipulates a logo image for QR code overlay.
//
// Create a Processor with [New] and optionally chain [Processor.WithTint].
// Use [Processor.Load] to read the image from disk.
type Processor struct {
	source    string
	sizeRatio float64
	tint      color.Color
}

// New creates a new Processor for the logo at source with the given size
// ratio. The sizeRatio determines how much of the QR code the logo will
// occupy (e.g., 0.25 = 25%).
func New(source string, sizeRatio float64) *Processor {
	return &Processor{
		source:    source,
		sizeRatio: sizeRatio,
	}
}

// WithTint sets a color tint to be applied during processing. Returns the
// receiver for method chaining. Pass nil to clear any previously set tint.
func (p *Processor) WithTint(c color.Color) *Processor {
	p.tint = c
	return p
}

// Source returns the logo source file path.
func (p *Processor) Source() string { return p.source }

// SizeRatio returns the logo size ratio as a fraction of the QR code size.
func (p *Processor) SizeRatio() float64 { return p.sizeRatio }

// Load reads and decodes the logo image from the source path.
// Returns an error if the source is empty, the file cannot be opened,
// or the image format cannot be decoded.
func (p *Processor) Load() (image.Image, error) {
	if p.source == "" {
		return nil, fmt.Errorf("logo: source path is empty")
	}
	f, err := os.Open(p.source)
	if err != nil {
		return nil, fmt.Errorf("logo: failed to open %q: %w", p.source, err)
	}
	defer f.Close() //nolint:errcheck // file.Close error is not actionable in this context
	img, _, err := image.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("logo: failed to decode %q: %w", p.source, err)
	}
	return img, nil
}

// LoadFromBytes decodes an image from raw bytes. Supports PNG, JPEG,
// and GIF formats via Go's standard image decoders.
func LoadFromBytes(data []byte) (image.Image, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("logo: empty data")
	}
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("logo: failed to decode image data: %w", err)
	}
	return img, nil
}

// LoadFromReader decodes an image from an [io.Reader]. Supports PNG,
// JPEG, and GIF formats via Go's standard image decoders.
func LoadFromReader(r io.Reader) (image.Image, error) {
	img, _, err := image.Decode(r)
	if err != nil {
		return nil, fmt.Errorf("logo: failed to decode image: %w", err)
	}
	return img, nil
}

// ResizeLogo resizes the logo to fit within the QR code based on the
// ratio of the QR module count. The aspect ratio of the original image
// is preserved. The result is an *image.RGBA.
func ResizeLogo(img image.Image, qrModules int, ratio float64) *image.RGBA {
	bounds := img.Bounds()
	origW := bounds.Dx()
	origH := bounds.Dy()
	targetSize := int(float64(qrModules) * ratio)
	if targetSize < 1 {
		targetSize = 1
	}
	aspect := float64(origW) / float64(origH)
	var w, h int
	if aspect >= 1 {
		w = targetSize
		h = int(float64(targetSize) / aspect)
	} else {
		h = targetSize
		w = int(float64(targetSize) * aspect)
	}
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	return bilinearResize(img, bounds, origW, origH, w, h)
}

// bilinearResize performs nearest-neighbor resampling of img into a new
// w×h *image.RGBA. It is the shared implementation for ResizeLogo and
// ResizeLogoToPixels.
func bilinearResize(img image.Image, bounds image.Rectangle, origW, origH, w, h int) *image.RGBA {
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			srcX := bounds.Min.X + int(float64(x)*float64(origW)/float64(w))
			srcY := bounds.Min.Y + int(float64(y)*float64(origH)/float64(h))
			dst.Set(x, y, img.At(srcX, srcY))
		}
	}
	return dst
}

// ResizeLogoToPixels resizes the logo to the exact pixel dimensions
// specified by width and height. Values less than 1 are clamped to 1.
func ResizeLogoToPixels(img image.Image, width, height int) *image.RGBA {
	if width < 1 {
		width = 1
	}
	if height < 1 {
		height = 1
	}
	bounds := img.Bounds()
	origW := bounds.Dx()
	origH := bounds.Dy()
	return bilinearResize(img, bounds, origW, origH, width, height)
}

// TintLogo applies a multiplicative color tint to the logo image.
// Each pixel's RGBA components are multiplied with the corresponding
// tint components. Pass nil for tint to simply clone the image.
func TintLogo(img image.Image, tint color.Color) *image.RGBA {
	if tint == nil {
		return CloneToRGBA(img)
	}
	bounds := img.Bounds()
	dst := image.NewRGBA(bounds)
	tr, tg, tb, ta := tint.RGBA()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			dstR := (r * tr) >> 16
			dstG := (g * tg) >> 16
			dstB := (b * tb) >> 16
			dstA := (a * ta) >> 16
			dst.SetRGBA(x, y, color.RGBA{
				R: uint8(dstR),
				G: uint8(dstG),
				B: uint8(dstB),
				A: uint8(dstA),
			})
		}
	}
	return dst
}

// OverlayLogo composites the logo onto the center of the QR code image
// using alpha blending. The QR code image is used as the base and the
// logo is drawn on top. Returns a new *image.RGBA with the composited result.
func OverlayLogo(qrImage, logoImg image.Image, _ int) *image.RGBA {
	qrBounds := qrImage.Bounds()
	dst := image.NewRGBA(qrBounds)
	draw.Draw(dst, qrBounds, qrImage, image.Point{}, draw.Src)
	logoBounds := logoImg.Bounds()
	logoW := logoBounds.Dx()
	logoH := logoBounds.Dy()
	offsetX := (qrBounds.Dx() - logoW) / 2
	offsetY := (qrBounds.Dy() - logoH) / 2
	logoRect := image.Rect(offsetX, offsetY, offsetX+logoW, offsetY+logoH)
	draw.Draw(dst, logoRect, logoImg, image.Point{}, draw.Over)
	return dst
}

// CloneToRGBA copies any image into a new *image.RGBA using source-over
// compositing. This is useful for normalizing image types before further
// processing.
func CloneToRGBA(img image.Image) *image.RGBA {
	bounds := img.Bounds()
	dst := image.NewRGBA(bounds)
	draw.Draw(dst, bounds, img, image.Point{}, draw.Src)
	return dst
}

// EncodePNG encodes an image as PNG bytes suitable for writing to
// a file or sending in an HTTP response.
func EncodePNG(img image.Image) ([]byte, error) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("logo: failed to encode PNG: %w", err)
	}
	return buf.Bytes(), nil
}

// Validate checks that the logo file at path exists, is a regular file
// (not a directory), and has a supported image format extension.
func Validate(path string) error {
	if path == "" {
		return fmt.Errorf("logo: path is empty")
	}
	ext := filepath.Ext(path)
	if !IsSupportedFormat(ext) {
		return fmt.Errorf("logo: unsupported format %q, supported: %v", ext, SupportedFormats())
	}
	f, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("logo: cannot access %q: %w", path, err)
	}
	if f.IsDir() {
		return fmt.Errorf("logo: %q is a directory, not a file", path)
	}
	return nil
}

// LogoSize calculates the logo pixel size from the QR code pixel size
// and the desired ratio. The result is clamped to a minimum of 1.
//
//nolint:revive // stutter: LogoSize is the canonical name for this function
func LogoSize(qrPixelSize int, ratio float64) int {
	size := int(float64(qrPixelSize) * ratio)
	if size < 1 {
		size = 1
	}
	return size
}
