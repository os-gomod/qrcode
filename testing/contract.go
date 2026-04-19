package testing

import (
	"context"
	"testing"

	qrcode "github.com/os-gomod/qrcode"
	"github.com/os-gomod/qrcode/encoding"
	"github.com/os-gomod/qrcode/payload"
)

// GeneratorContractTest runs a suite of contract tests on a QR code Generator.
//
// It validates TextPayload and WiFiPayload generation, error handling for
// invalid payloads, options passthrough (error correction level), writer
// output, and proper Close lifecycle behavior.
func GeneratorContractTest(t *testing.T, gen qrcode.Generator) {
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
		AssertFalse(t, gen.Closed(), "generator should not be closed before Close()")
	})
	t.Run("Close", func(t *testing.T) {
		t.Helper()
		err := gen.Close(ctx)
		AssertNoError(t, err)
		AssertTrue(t, gen.Closed(), "generator should be closed after Close()")
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

// QRCodeIsValid checks that the QR code matrix is well-formed, verifying
// that Version (1–40), Size (≥ 21), ECLevel (0–3), MaskPattern (0–7),
// and the Modules matrix dimensions are all valid.
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
