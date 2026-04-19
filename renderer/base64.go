package renderer

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"

	"github.com/os-gomod/qrcode/encoding"
)

// Base64Renderer renders QR codes as base64-encoded PNG data URIs.
//
// This renderer wraps a PNGRenderer and encodes its output into base64.
// It is designed for scenarios where the QR code needs to be embedded
// directly in HTML <img> tags or CSS background-image properties.
//
// The renderer delegates all rendering options (including module styles,
// gradients, and transparency) to the underlying PNGRenderer.
//
// Example:
//
//	r := renderer.NewBase64Renderer()
//	dataURI, err := r.RenderDataURL(ctx, qr)
//	// dataURI = "data:image/png;base64,iVBORw0KGgo..."
type Base64Renderer struct {
	pngRenderer *PNGRenderer
}

// NewBase64Renderer creates a new Base64Renderer with an underlying
// PNGRenderer. The returned renderer is stateless and safe for concurrent use.
func NewBase64Renderer() *Base64Renderer {
	return &Base64Renderer{
		pngRenderer: NewPNGRenderer(),
	}
}

// Type returns the format identifier "base64".
func (r *Base64Renderer) Type() string { return "base64" }

// ContentType returns the MIME type "text/plain".
func (r *Base64Renderer) ContentType() string { return "text/plain" }

// Render writes the QR code as a base64 PNG data URI to w.
// The output format is: data:image/png;base64,<base64-encoded-png-bytes>.
// All standard RenderOptions (including module styles) are supported.
func (r *Base64Renderer) Render(ctx context.Context, qr *encoding.QRCode, w io.Writer, opts ...RenderOption) error {
	encoded, err := r.RenderToString(ctx, qr, opts...)
	if err != nil {
		return err
	}
	_, err = io.WriteString(w, "data:image/png;base64,"+encoded)
	return err
}

// RenderToString renders the QR code and returns the raw base64-encoded
// PNG bytes as a string, without the data URI prefix. This is useful when
// you need only the base64 payload for custom URI schemes or API payloads.
func (r *Base64Renderer) RenderToString(ctx context.Context, qr *encoding.QRCode, opts ...RenderOption) (string, error) {
	var buf bytes.Buffer
	if err := r.pngRenderer.Render(ctx, qr, &buf, opts...); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

// RenderDataURL renders the QR code and returns a complete data URI string
// in the form "data:image/png;base64,<base64>". This value can be used
// directly as the src attribute of an HTML <img> tag.
func (r *Base64Renderer) RenderDataURL(ctx context.Context, qr *encoding.QRCode, opts ...RenderOption) (string, error) {
	b64, err := r.RenderToString(ctx, qr, opts...)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("data:image/png;base64,%s", b64), nil
}

// Encode is a convenience function that renders a QR code as a base64
// PNG data URI string using default render options. It is equivalent to:
//
//	renderer.NewBase64Renderer().RenderDataURL(context.Background(), qr)
func Encode(qr *encoding.QRCode) (string, error) {
	r := NewBase64Renderer()
	return r.RenderDataURL(context.Background(), qr)
}

// EncodeRaw is a convenience function that renders a QR code as raw
// base64-encoded PNG bytes (without the data URI prefix) using default
// render options. It is equivalent to:
//
//	renderer.NewBase64Renderer().RenderToString(context.Background(), qr)
func EncodeRaw(qr *encoding.QRCode) (string, error) {
	r := NewBase64Renderer()
	return r.RenderToString(context.Background(), qr)
}
