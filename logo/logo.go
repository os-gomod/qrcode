package logo

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"

	_ "image/gif"  // register GIF decoder
	_ "image/jpeg" // register JPEG decoder
)

func SupportedFormats() []string {
	return []string{".png", ".jpg", ".jpeg", ".gif"}
}

func IsSupportedFormat(ext string) bool {
	lower := strings.ToLower(ext)
	for _, f := range SupportedFormats() {
		if lower == f {
			return true
		}
	}
	return false
}

type Processor struct {
	source    string
	sizeRatio float64
	tint      color.Color
}

func New(source string, sizeRatio float64) *Processor {
	return &Processor{
		source:    source,
		sizeRatio: sizeRatio,
	}
}

func (p *Processor) WithTint(c color.Color) *Processor {
	p.tint = c
	return p
}
func (p *Processor) Source() string     { return p.source }
func (p *Processor) SizeRatio() float64 { return p.sizeRatio }
func (p *Processor) Load() (image.Image, error) {
	if p.source == "" {
		return nil, errors.New("logo: source path is empty")
	}
	f, err := os.Open(p.source)
	if err != nil {
		return nil, fmt.Errorf("logo: failed to open %q: %w", p.source, err)
	}
	defer func() { _ = f.Close() }()
	img, _, err := image.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("logo: failed to decode %q: %w", p.source, err)
	}
	return img, nil
}

func LoadFromBytes(data []byte) (image.Image, error) {
	if len(data) == 0 {
		return nil, errors.New("logo: empty data")
	}
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("logo: failed to decode image data: %w", err)
	}
	return img, nil
}

func LoadFromReader(r io.Reader) (image.Image, error) {
	img, _, err := image.Decode(r)
	if err != nil {
		return nil, fmt.Errorf("logo: failed to decode image: %w", err)
	}
	return img, nil
}

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

func bilinearResize(img image.Image, bounds image.Rectangle, origW, origH, w, h int) *image.RGBA {
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			srcX := bounds.Min.X + int(float64(x)*float64(origW)/float64(w))
			srcY := bounds.Min.Y + int(float64(y)*float64(origH)/float64(h))
			dst.Set(x, y, img.At(srcX, srcY))
		}
	}
	return dst
}

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

func CloneToRGBA(img image.Image) *image.RGBA {
	bounds := img.Bounds()
	dst := image.NewRGBA(bounds)
	draw.Draw(dst, bounds, img, image.Point{}, draw.Src)
	return dst
}

func EncodePNG(img image.Image) ([]byte, error) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("logo: failed to encode PNG: %w", err)
	}
	return buf.Bytes(), nil
}

func Validate(path string) error {
	if path == "" {
		return errors.New("logo: path is empty")
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

func LogoSize(qrPixelSize int, ratio float64) int {
	size := int(float64(qrPixelSize) * ratio)
	if size < 1 {
		size = 1
	}
	return size
}
