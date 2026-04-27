package qrcode

import (
	"context"
	"time"

	"github.com/os-gomod/qrcode/v2/payload"
)

// quickSize resolves the optional size argument, defaulting to 256.
// This is the single shared implementation used by all Quick helpers.
func quickSize(size ...int) int {
	if len(size) > 0 && size[0] > 0 {
		return size[0]
	}
	return 256
}

// ---------------------------------------------------------------------------
// Quick helpers — convenience functions that create a temporary Client,
// perform a single operation, and close it. For high-throughput or
// custom-configuration use, create a Client via New() or NewClient() and reuse it.
// ---------------------------------------------------------------------------

// Quick generates a QR code as PNG bytes for the given text data.
func Quick(data string, size ...int) ([]byte, error) {
	return quickRender(&payload.TextPayload{Text: data}, FormatPNG, size...)
}

// QuickSVG generates a QR code as an SVG string for the given text data.
func QuickSVG(data string, size ...int) (string, error) {
	raw, err := quickRender(&payload.TextPayload{Text: data}, FormatSVG, size...)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

// QuickFile generates a QR code and saves it to a file.
// The output format is inferred from the file extension (.png, .svg, .pdf, .txt).
func QuickFile(data, path string, size ...int) error {
	ctx := context.Background()
	gen, err := New(WithDefaultSize(quickSize(size...)))
	if err != nil {
		return err
	}
	defer func() { _ = gen.Close() }()
	return gen.Save(ctx, &payload.TextPayload{Text: data}, path)
}

// QuickURL generates a URL QR code as PNG bytes.
func QuickURL(url string, size ...int) ([]byte, error) {
	return quickRender(&payload.URLPayload{URL: url}, FormatPNG, size...)
}

// QuickWiFi generates a WiFi QR code as PNG bytes.
func QuickWiFi(ssid, password, encryption string, size ...int) ([]byte, error) {
	return quickRender(&payload.WiFiPayload{
		SSID:       ssid,
		Password:   password,
		Encryption: encryption,
	}, FormatPNG, size...)
}

// QuickContact generates a vCard QR code as PNG bytes.
func QuickContact(firstName, lastName, phone, email string, size ...int) ([]byte, error) {
	return quickRender(&payload.VCardPayload{
		FirstName: firstName,
		LastName:  lastName,
		Phone:     phone,
		Email:     email,
	}, FormatPNG, size...)
}

// QuickSMS generates an SMS QR code as PNG bytes.
func QuickSMS(phone, message string, size ...int) ([]byte, error) {
	return quickRender(&payload.SMSPayload{
		Phone:   phone,
		Message: message,
	}, FormatPNG, size...)
}

// QuickEmail generates an email QR code as PNG bytes.
func QuickEmail(to, subject, body string, size ...int) ([]byte, error) {
	return quickRender(&payload.EmailPayload{
		To:      to,
		Subject: subject,
		Body:    body,
	}, FormatPNG, size...)
}

// QuickGeo generates a geo-location QR code as PNG bytes.
func QuickGeo(lat, lng float64, size ...int) ([]byte, error) {
	return quickRender(&payload.GeoPayload{
		Latitude:  lat,
		Longitude: lng,
	}, FormatPNG, size...)
}

// QuickEvent generates a calendar event QR code as PNG bytes.
func QuickEvent(title, location string, start, end time.Time, size ...int) ([]byte, error) {
	return quickRender(&payload.CalendarPayload{
		Title:    title,
		Location: location,
		Start:    start,
		End:      end,
	}, FormatPNG, size...)
}

// QuickPayment generates a PayPal payment QR code as PNG bytes.
func QuickPayment(username, amount string, size ...int) ([]byte, error) {
	return quickRender(&payload.PayPalPayload{
		Username: username,
		Amount:   amount,
	}, FormatPNG, size...)
}

// quickRender is the single shared implementation for all Quick PNG/String helpers.
// It creates a temporary client, renders to the specified format, and closes it.
func quickRender(p payload.Payload, format Format, size ...int) ([]byte, error) {
	ctx := context.Background()
	gen, err := New(WithDefaultSize(quickSize(size...)))
	if err != nil {
		return nil, err
	}
	defer func() { _ = gen.Close() }()
	return gen.Render(ctx, p, format)
}
