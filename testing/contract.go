package testing

import (
	"context"
	"testing"

	qrcode "github.com/os-gomod/qrcode/v2"
	"github.com/os-gomod/qrcode/v2/internal/encoding"
	"github.com/os-gomod/qrcode/v2/payload"
)

func ClientContractTest(t *testing.T, gen qrcode.Client) {
	t.Helper()
	ctx := context.Background()
	t.Run("TextPayload", func(t *testing.T) {
		t.Helper()
		p := &payload.TextPayload{Text: "https://example.com"}
		qr, err := gen.Generate(ctx, p)
		AssertNoError(t, err)
		if qr == nil {
			t.Fatal("expected non-nil QRCode, got nil")
		}
		AssertTrue(t, qr.Size > 0, "QR code Size should be positive")
		AssertTrue(t, qr.Version >= 1, "QR code Version should be >= 1")
	})
	t.Run("WiFiPayload", func(t *testing.T) {
		t.Helper()
		p := &payload.WiFiPayload{
			SSID:       "TestNetwork",
			Password:   "testpass123",
			Encryption: payload.EncryptionWPA2,
		}
		qr, err := gen.Generate(ctx, p)
		AssertNoError(t, err)
		if qr == nil {
			t.Fatal("expected non-nil QRCode, got nil")
		}
		AssertTrue(t, qr.Size > 0, "WiFi QR code Size should be positive")
	})
	t.Run("InvalidPayload_NilText", func(t *testing.T) {
		t.Helper()
		p := &payload.TextPayload{Text: ""}
		_, err := gen.Generate(ctx, p)
		AssertTrue(t, err != nil, "expected error for empty text payload, got nil")
	})
	t.Run("InvalidPayload_BadWiFi", func(t *testing.T) {
		t.Helper()
		p := &payload.WiFiPayload{
			SSID:       "",
			Password:   "",
			Encryption: "INVALID",
		}
		_, err := gen.Generate(ctx, p)
		AssertTrue(t, err != nil, "expected error for invalid WiFi payload, got nil")
	})
	t.Run("GenerateWithOptions", func(t *testing.T) {
		t.Helper()
		p := &payload.TextPayload{Text: "option-test"}
		qr, err := gen.GenerateWithOptions(ctx, p, qrcode.WithErrorCorrection(qrcode.LevelH))
		AssertNoError(t, err)
		if qr == nil {
			t.Fatal("expected non-nil QRCode, got nil")
		}
		AssertEquals(t, 3, qr.ECLevel)
	})
	t.Run("GenerateToWriter", func(t *testing.T) {
		t.Helper()
		p := &payload.TextPayload{Text: "writer-test"}
		var buf []byte
		err := gen.GenerateToWriter(ctx, p, &byteSliceWriter{buf: &buf}, qrcode.FormatSVG)
		AssertNoError(t, err)
		AssertTrue(t, len(buf) > 0, "expected non-empty SVG output")
	})
	t.Run("Closed", func(t *testing.T) {
		t.Helper()
		AssertFalse(t, gen.Closed(), "client should not be closed before Close()")
	})
	t.Run("Close", func(t *testing.T) {
		t.Helper()
		err := gen.Close()
		AssertNoError(t, err)
		AssertTrue(t, gen.Closed(), "client should be closed after Close()")
		p := &payload.TextPayload{Text: "after-close"}
		_, err = gen.Generate(ctx, p)
		AssertTrue(t, err != nil, "expected error when generating after Close()")
	})
}

type byteSliceWriter struct {
	buf *[]byte
}

func (w *byteSliceWriter) Write(p []byte) (int, error) {
	*w.buf = append(*w.buf, p...)
	return len(p), nil
}

func QRCodeIsValid(t *testing.T, qr *encoding.QRCode) {
	t.Helper()
	AssertTrue(t, qr != nil, "QRCode must not be nil")
	AssertTrue(t, qr.Version >= 1 && qr.Version <= 40, "Version must be 1-40")
	AssertTrue(t, qr.Size >= 21, "Size must be >= 21 (version 1)")
	AssertTrue(t, qr.ECLevel >= 0 && qr.ECLevel <= 3, "ECLevel must be 0-3")
	AssertTrue(t, qr.MaskPattern >= 0 && qr.MaskPattern <= 7, "MaskPattern must be 0-7")
	AssertTrue(t, len(qr.Modules) == qr.Size, "Modules matrix must have Size rows")
	if len(qr.Modules) > 0 {
		AssertTrue(t, len(qr.Modules[0]) == qr.Size, "Modules matrix must have Size columns")
	}
}
