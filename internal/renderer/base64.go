package renderer

import (
	"context"
	"encoding/base64"

	"github.com/os-gomod/qrcode/v2/internal/encoding"
)

// Base64Renderer renders QR codes as Base64-encoded data URLs (PNG-based).
// It is safe for concurrent use.
type Base64Renderer struct {
	pngRenderer *PNGRenderer
}

// NewBase64Renderer creates a new Base64Renderer.
func NewBase64Renderer() *Base64Renderer {
	return &Base64Renderer{
		pngRenderer: NewPNGRenderer(),
	}
}

// Render encodes the QR matrix as a Base64 data URL with the PNG MIME type prefix.
func (r *Base64Renderer) Render(ctx context.Context, qr *encoding.QRCode, opts ...RenderOption) ([]byte, error) {
	encoded, err := r.RenderToString(ctx, qr, opts...)
	if err != nil {
		return nil, err
	}
	return []byte("data:image/png;base64," + encoded), nil
}

// RenderToString returns raw Base64-encoded PNG data (without the data URL prefix).
func (r *Base64Renderer) RenderToString(ctx context.Context, qr *encoding.QRCode, opts ...RenderOption) (string, error) {
	pngBytes, err := r.pngRenderer.Render(ctx, qr, opts...)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(pngBytes), nil
}

// RenderDataURL returns the complete data URL string.
func (r *Base64Renderer) RenderDataURL(ctx context.Context, qr *encoding.QRCode, opts ...RenderOption) (string, error) {
	encoded, err := r.RenderToString(ctx, qr, opts...)
	if err != nil {
		return "", err
	}
	return "data:image/png;base64," + encoded, nil
}

// Encode is a convenience function that encodes a QR code as a Base64 data URL
// using default settings. It is equivalent to:
//
//	r := NewBase64Renderer()
//	return r.RenderDataURL(context.Background(), qr)
func Encode(qr *encoding.QRCode) (string, error) {
	r := NewBase64Renderer()
	return r.RenderDataURL(context.Background(), qr)
}

// EncodeRaw is a convenience function that returns raw Base64-encoded PNG data
// without the data URL prefix.
func EncodeRaw(qr *encoding.QRCode) (string, error) {
	r := NewBase64Renderer()
	return r.RenderToString(context.Background(), qr)
}
