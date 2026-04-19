package renderer

import (
	"context"
	"fmt"
	"io"

	"github.com/os-gomod/qrcode/encoding"
)

// PDFRenderer renders QR codes as PDF (Portable Document Format) documents.
//
// The renderer produces a minimal, self-contained PDF 1.4 file with a single
// page containing the QR code. Each dark module is drawn as a filled rectangle
// using the PDF content stream operator "re" (rectangle) with the fill operator "f".
// Colors are specified in DeviceRGB color space.
//
// The page dimensions are automatically computed to fit the QR code plus quiet
// zone and any configured border width, with an additional one-module margin.
//
// Example:
//
//	r := renderer.NewPDFRenderer()
//	err := r.Render(ctx, qr, os.Stdout,
//	    renderer.WithForegroundColor("#000000"),
//	    renderer.WithBorderWidth(10),
//	)
type PDFRenderer struct{}

// NewPDFRenderer creates a new PDFRenderer. The returned renderer is
// stateless and safe for concurrent use.
func NewPDFRenderer() *PDFRenderer {
	return &PDFRenderer{}
}

// Type returns the format identifier "pdf".
func (r *PDFRenderer) Type() string { return "pdf" }

// ContentType returns the MIME type "application/pdf".
func (r *PDFRenderer) ContentType() string { return "application/pdf" }

// Render writes the QR code as a PDF 1.4 document to w using the given
// render options.
//
// The generated PDF contains a catalog, page tree, single page, and a
// content stream object. Module size is fixed at 10pt. The page size is
// computed as (qrSize + 2*quietZone)*moduleSize + 2*(borderWidth + moduleSize).
// Both foreground and background colors must be valid "#RRGGBB" hex strings.
func (r *PDFRenderer) Render(_ context.Context, qr *encoding.QRCode, w io.Writer, opts ...RenderOption) error {
	cfg := ApplyOptions(opts...)
	fgR, fgG, fgB, err := ParseHexColor(cfg.ForegroundColor)
	if err != nil {
		return fmt.Errorf("invalid foreground color: %w", err)
	}
	bgR, bgG, bgB, err := ParseHexColor(cfg.BackgroundColor)
	if err != nil {
		return fmt.Errorf("invalid background color: %w", err)
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
	content := buildContentStream(qr, moduleSize, qz, border, fg, bg)
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
	_, err = io.WriteString(w, p.String())
	return err
}

func pdfColor(r, g, b uint8) string {
	return fmt.Sprintf("%.3f %.3f %.3f", float64(r)/255.0, float64(g)/255.0, float64(b)/255.0)
}

func buildContentStream(qr *encoding.QRCode, moduleSize, qz, border float64, fg, bg string) string {
	var s pdfStream
	qrPixelSize := (float64(qr.Size) + 2*qz) * moduleSize
	offsetX := border + moduleSize
	offsetY := border + moduleSize
	s.rectf(offsetX, offsetY, qrPixelSize, qrPixelSize, bg, "f")
	s.setFillColor(fg)
	for row := 0; row < qr.Size; row++ {
		for col := 0; col < qr.Size; col++ {
			if qr.Modules[row][col] {
				x := offsetX + (float64(col)+qz)*moduleSize
				y := offsetY + (float64(row)+qz)*moduleSize
				s.rectf(x, y, moduleSize, moduleSize, "", "f")
			}
		}
	}
	return s.String()
}

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
		p.write(keyValues[i+1]) //nolint:gosec // G602: index is always valid for even-length key-value pairs
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
func (p *pdfBuilder) String() string { return string(p.buf) }

type pdfStream struct {
	buf []byte
}

func (s *pdfStream) setFillColor(color string) {
	s.write(fmt.Sprintf("%s rg\n", color))
}

func (s *pdfStream) rectf(x, y, w, h float64, fillColor, op string) {
	if fillColor != "" {
		s.write(fmt.Sprintf("%s rg\n", fillColor))
	}
	s.write(fmt.Sprintf("%.2f %.2f %.2f %.2f re %s\n", x, y, w, h, op))
}

func (s *pdfStream) write(data string) {
	s.buf = append(s.buf, data...)
}
func (s *pdfStream) String() string { return string(s.buf) }
