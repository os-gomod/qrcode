package qrcode

import (
	"bytes"
	"context"
	"os"
	"strings"
	"time"

	"github.com/os-gomod/qrcode/payload"
)

// quickSize returns the pixel size, defaulting to 256 when no size argument is provided.
func quickSize(size ...int) int {
	if len(size) > 0 && size[0] > 0 {
		return size[0]
	}
	return 256
}

// quickGen creates a short-lived generator with the given options, renders the payload
// as PNG bytes, and closes the generator. This is the shared implementation for all
// package-level Quick* convenience functions.
func quickGen(p payload.Payload, size ...int) ([]byte, error) {
	gen, err := New(WithDefaultSize(quickSize(size...)))
	if err != nil {
		return nil, err
	}
	defer gen.Close(context.Background()) //nolint:errcheck // Close error intentionally ignored in fire-and-forget convenience method
	return GeneratePNG(context.Background(), gen, p)
}

// quickGenSVG is like quickGen but renders to SVG format and returns a string.
func quickGenSVG(p payload.Payload, size ...int) (string, error) {
	gen, err := New(WithDefaultSize(quickSize(size...)))
	if err != nil {
		return "", err
	}
	defer gen.Close(context.Background()) //nolint:errcheck // Close error intentionally ignored in fire-and-forget convenience method
	return GenerateSVG(context.Background(), gen, p)
}

// Quick generates a PNG QR code from the given text data with an optional image size.
func Quick(data string, size ...int) ([]byte, error) {
	return quickGen(&payload.TextPayload{Text: data}, size...)
}

// QuickSVG generates an SVG QR code from the given text data with an optional image size.
func QuickSVG(data string, size ...int) (string, error) {
	return quickGenSVG(&payload.TextPayload{Text: data}, size...)
}

// QuickFile generates a QR code from the given text data and writes it to path.
// The output format is inferred from the file extension (.png, .svg, or .pdf).
func QuickFile(data, path string, size ...int) error {
	gen, err := New(WithDefaultSize(quickSize(size...)))
	if err != nil {
		return err
	}
	defer gen.Close(context.Background()) //nolint:errcheck // Close error intentionally ignored in fire-and-forget convenience method
	return Save(context.Background(), gen, &payload.TextPayload{Text: data}, path)
}

// QuickWiFi generates a PNG QR code encoding a WiFi network configuration.
func QuickWiFi(ssid, password, encryption string, size ...int) ([]byte, error) {
	return quickGen(&payload.WiFiPayload{
		SSID:       ssid,
		Password:   password,
		Encryption: encryption,
	}, size...)
}

// QuickURL generates a PNG QR code encoding the given URL.
func QuickURL(url string, size ...int) ([]byte, error) {
	return quickGen(&payload.URLPayload{URL: url}, size...)
}

// QuickContact generates a PNG QR code encoding a vCard contact.
func QuickContact(firstName, lastName, phone, email string, size ...int) ([]byte, error) {
	return quickGen(&payload.VCardPayload{
		FirstName: firstName,
		LastName:  lastName,
		Phone:     phone,
		Email:     email,
	}, size...)
}

// QuickSMS generates a PNG QR code encoding an SMS message.
func QuickSMS(phone, message string, size ...int) ([]byte, error) {
	return quickGen(&payload.SMSPayload{
		Phone:   phone,
		Message: message,
	}, size...)
}

// QuickEmail generates a PNG QR code encoding an email message.
func QuickEmail(to, subject, body string, size ...int) ([]byte, error) {
	return quickGen(&payload.EmailPayload{
		To:      to,
		Subject: subject,
		Body:    body,
	}, size...)
}

// QuickGeo generates a PNG QR code encoding a geographic location.
func QuickGeo(lat, lng float64, size ...int) ([]byte, error) {
	return quickGen(&payload.GeoPayload{
		Latitude:  lat,
		Longitude: lng,
	}, size...)
}

// QuickEvent generates a PNG QR code encoding a calendar event.
func QuickEvent(title, location string, start, end time.Time, size ...int) ([]byte, error) {
	return quickGen(&payload.CalendarPayload{
		Title:    title,
		Location: location,
		Start:    start,
		End:      end,
	}, size...)
}

// QuickPayment generates a PNG QR code encoding a PayPal payment link.
func QuickPayment(username, amount string, size ...int) ([]byte, error) {
	return quickGen(&payload.PayPalPayload{
		Username: username,
		Amount:   amount,
	}, size...)
}

// GeneratePNG renders a QR code from the payload as PNG bytes.
func GeneratePNG(ctx context.Context, gen Generator, p payload.Payload) ([]byte, error) {
	var buf bytes.Buffer
	if err := gen.GenerateToWriter(ctx, p, &buf, FormatPNG); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// GenerateSVG renders a QR code from the payload as an SVG string.
func GenerateSVG(ctx context.Context, gen Generator, p payload.Payload) (string, error) {
	var buf bytes.Buffer
	if err := gen.GenerateToWriter(ctx, p, &buf, FormatSVG); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// GenerateASCII renders a QR code from the payload as Unicode block characters.
func GenerateASCII(ctx context.Context, gen Generator, p payload.Payload) (string, error) {
	var buf bytes.Buffer
	if err := gen.GenerateToWriter(ctx, p, &buf, FormatTerminal); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// GenerateBase64 renders a QR code from the payload as a base64-encoded PNG data URI.
func GenerateBase64(ctx context.Context, gen Generator, p payload.Payload) (string, error) {
	var buf bytes.Buffer
	err := gen.GenerateToWriter(ctx, p, &buf, FormatBase64)
	return buf.String(), err
}

// SavePNG renders a QR code from the payload and writes the PNG to the given file path.
func SavePNG(ctx context.Context, gen Generator, p payload.Payload, path string) error {
	var buf bytes.Buffer
	if err := gen.GenerateToWriter(ctx, p, &buf, FormatPNG); err != nil {
		return err
	}
	return os.WriteFile(path, buf.Bytes(), 0o644) //nolint:gosec // G306: output files are intentionally world-readable
}

// SaveSVG renders a QR code from the payload and writes the SVG to the given file path.
func SaveSVG(ctx context.Context, gen Generator, p payload.Payload, path string) error {
	var buf bytes.Buffer
	if err := gen.GenerateToWriter(ctx, p, &buf, FormatSVG); err != nil {
		return err
	}
	return os.WriteFile(path, buf.Bytes(), 0o644) //nolint:gosec // G306: output files are intentionally world-readable
}

// Save renders a QR code from the payload and writes it to the given file path.
// The output format is inferred from the file extension (.png, .svg, or .pdf).
func Save(ctx context.Context, gen Generator, p payload.Payload, path string) error {
	ext := extensionFromPath(path)
	switch ext {
	case ".svg":
		return SaveSVG(ctx, gen, p, path)
	case ".pdf":
		var buf bytes.Buffer
		if err := gen.GenerateToWriter(ctx, p, &buf, FormatPDF); err != nil {
			return err
		}
		return os.WriteFile(path, buf.Bytes(), 0o644) //nolint:gosec // G306: output files are intentionally world-readable
	default:
		return SavePNG(ctx, gen, p, path)
	}
}

// extensionFromPath returns the lowercase file extension of path, including the dot.
func extensionFromPath(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '.' {
			return strings.ToLower(path[i:])
		}
		if path[i] == '/' || path[i] == '\\' {
			break
		}
	}
	return ""
}
