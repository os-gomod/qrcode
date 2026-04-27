package renderer

import (
	"bytes"
	"context"
	"fmt"

	"github.com/os-gomod/qrcode/v2/internal/encoding"
)

// PDFRenderer renders QR codes as PDF documents.
// It is safe for concurrent use.
type PDFRenderer struct{}

// NewPDFRenderer creates a new PDFRenderer.
func NewPDFRenderer() *PDFRenderer {
	return &PDFRenderer{}
}

// Render encodes the QR matrix as a minimal PDF document.
func (*PDFRenderer) Render(_ context.Context, qr *encoding.QRCode, opts ...RenderOption) ([]byte, error) {
	cfg := ApplyOptions(opts...)

	fgR, fgG, fgB, err := ParseHexColor(cfg.ForegroundColor)
	if err != nil {
		return nil, fmt.Errorf("invalid foreground color: %w", err)
	}
	bgR, bgG, bgB, err := ParseHexColor(cfg.BackgroundColor)
	if err != nil {
		return nil, fmt.Errorf("invalid background color: %w", err)
	}

	moduleSize := 10.0
	qz := float64(cfg.QuietZone)
	border := float64(cfg.BorderWidth)
	qrSize := float64(qr.Size)
	contentWidth := (qrSize + 2*qz) * moduleSize
	pageMargin := border + moduleSize
	pageWidth := contentWidth + 2*pageMargin
	pageHeight := contentWidth + 2*pageMargin

	fg := pdfColor(fgR, fgG, fgB)
	bg := pdfColor(bgR, bgG, bgB)
	content := buildPDFContentStream(qr, moduleSize, qz, border, fg, bg)

	p := &pdfBuilder{}
	p.header()
	p.dictObj(1, "/Type", "/Catalog", "/Pages", "2 0 R")
	p.dictObj(2, "/Type", "/Pages", "/Kids", "[3 0 R]", "/Count", "1")
	p.dictObj(3,
		"/Type", "/Page",
		"/Parent", "2 0 R",
		"/MediaBox", fmt.Sprintf("[0 0 %.2f %.2f]", pageWidth, pageHeight),
		"/Contents", "4 0 R",
		"/Resources", "<< /ColorSpace << /DeviceRGB /DeviceRGB >> >>",
	)
	p.streamObj(4, content)
	p.finish()

	var buf bytes.Buffer
	buf.Write(p.Bytes())
	return buf.Bytes(), nil
}

// ---------------------------------------------------------------------------
// PDF builder helpers
// ---------------------------------------------------------------------------

func pdfColor(r, g, b uint8) string {
	return fmt.Sprintf("%.3f %.3f %.3f", float64(r)/255.0, float64(g)/255.0, float64(b)/255.0)
}

func buildPDFContentStream(qr *encoding.QRCode, moduleSize, qz, border float64, fg, bg string) string {
	var s pdfStream
	qrPixelSize := (float64(qr.Size) + 2*qz) * moduleSize
	offsetX := border + moduleSize
	offsetY := border + moduleSize
	s.rectf(offsetX, offsetY, qrPixelSize, qrPixelSize, bg, "f")
	s.setFillColor(fg)
	for row := range qr.Size {
		for col := range qr.Size {
			if qr.Modules[row][col] {
				x := offsetX + (float64(col)+qz)*moduleSize
				y := offsetY + (float64(row)+qz)*moduleSize
				s.rectf(x, y, moduleSize, moduleSize, "", "f")
			}
		}
	}
	return s.String()
}

// ---------------------------------------------------------------------------
// pdfBuilder — incremental PDF document construction
// ---------------------------------------------------------------------------

type pdfBuilder struct {
	buf     []byte
	offsets []int
	lastObj int
}

func (p *pdfBuilder) header() {
	p.write("%PDF-1.4\n%\xe2\xe3\xcf\xd3\n")
}

func (p *pdfBuilder) dictObj(num int, keyValues ...string) {
	p.reserveOffset(num)
	p.write(fmt.Sprintf("%d 0 obj\n<< ", num))
	for i := 0; i < len(keyValues); i += 2 {
		if i > 0 {
			p.write(" ")
		}
		p.write(keyValues[i])
		p.write(" ")
		p.write(keyValues[i+1]) //nolint:gosec //G602: keyValues always has even length (key-value pairs)
	}
	p.write(" >>\nendobj\n")
}

func (p *pdfBuilder) streamObj(num int, data string) {
	p.reserveOffset(num)
	p.write(fmt.Sprintf("%d 0 obj\n<< /Length %d >>\nstream\n", num, len(data)))
	p.write(data)
	p.write("\nendstream\nendobj\n")
}

func (p *pdfBuilder) reserveOffset(num int) {
	for len(p.offsets) <= num {
		p.offsets = append(p.offsets, 0)
	}
	p.offsets[num] = len(p.buf)
	if num > p.lastObj {
		p.lastObj = num
	}
}

func (p *pdfBuilder) finish() {
	xrefOffset := len(p.buf)
	p.write("xref\n")
	p.write(fmt.Sprintf("0 %d\n", p.lastObj+1))
	p.write("0000000000 65535 f \n")
	for i := 1; i <= p.lastObj; i++ {
		p.write(fmt.Sprintf("%010d 00000 n \n", p.offsets[i]))
	}
	p.write("trailer\n")
	p.write(fmt.Sprintf("<< /Size %d /Root 1 0 R >>\n", p.lastObj+1))
	p.write("startxref\n")
	p.write(fmt.Sprintf("%d\n", xrefOffset))
	p.write("%%EOF\n")
}

func (p *pdfBuilder) write(s string) {
	p.buf = append(p.buf, s...)
}

func (p *pdfBuilder) Bytes() []byte {
	return p.buf
}

// ---------------------------------------------------------------------------
// pdfStream — content stream builder
// ---------------------------------------------------------------------------

type pdfStream struct {
	buf []byte
}

func (s *pdfStream) setFillColor(color string) {
	s.write(color + " rg\n")
}

func (s *pdfStream) rectf(x, y, w, h float64, fillColor, op string) {
	if fillColor != "" {
		s.write(fillColor + " rg\n")
	}
	s.write(fmt.Sprintf("%.2f %.2f %.2f %.2f re %s\n", x, y, w, h, op))
}

func (s *pdfStream) write(data string) {
	s.buf = append(s.buf, data...)
}

func (s *pdfStream) String() string {
	return string(s.buf)
}
