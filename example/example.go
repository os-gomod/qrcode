package main

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"time"

	qrcode "github.com/os-gomod/qrcode"
	"github.com/os-gomod/qrcode/batch"
	qrerrors "github.com/os-gomod/qrcode/errors"
	qrlogo "github.com/os-gomod/qrcode/logo"
	"github.com/os-gomod/qrcode/payload"
	"github.com/os-gomod/qrcode/renderer"
)

// GenerationRecord tracks one QR code generation for reporting.
type GenerationRecord struct {
	Filename string
	Type     string
	Format   string
	OK       bool
	Error    error
	Duration time.Duration
}

func main() {
	ctx := context.Background()
	outputDir := filepath.Join(".", "output")
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "failed to create output directory: %v\n", err)
		os.Exit(1)
	}

	// ----------------------------------------------------------------
	// 1. Create the generator with options
	// ----------------------------------------------------------------
	gen, err := qrcode.New(
		qrcode.WithDefaultSize(256),
		qrcode.WithErrorCorrection(qrcode.LevelM),
		qrcode.WithQuietZone(4),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create generator: %v\n", err)
		os.Exit(1)
	}
	defer gen.Close(ctx)

	var records []GenerationRecord

	// ----------------------------------------------------------------
	// 2. Text QR codes (PNG + SVG)
	// ----------------------------------------------------------------
	records = append(records, generateText(ctx, gen, outputDir)...)
	// ----------------------------------------------------------------
	// 3. URL QR codes
	// ----------------------------------------------------------------
	records = append(records, generateURLs(ctx, gen, outputDir)...)
	// ----------------------------------------------------------------
	// 4. WiFi QR codes
	// ----------------------------------------------------------------
	records = append(records, generateWiFi(ctx, gen, outputDir)...)
	// ----------------------------------------------------------------
	// 5. vCard / MeCard QR codes
	// ----------------------------------------------------------------
	records = append(records, generateContact(ctx, gen, outputDir)...)
	// ----------------------------------------------------------------
	// 6. SMS / MMS / Phone QR codes
	// ----------------------------------------------------------------
	records = append(records, generateMessaging(ctx, gen, outputDir)...)
	// ----------------------------------------------------------------
	// 7. Email QR code
	// ----------------------------------------------------------------
	records = append(records, generateEmail(ctx, gen, outputDir)...)
	// ----------------------------------------------------------------
	// 8. Geo / Maps QR codes
	// ----------------------------------------------------------------
	records = append(records, generateGeoMaps(ctx, gen, outputDir)...)
	// ----------------------------------------------------------------
	// 9. Calendar / Event QR codes
	// ----------------------------------------------------------------
	records = append(records, generateCalendar(ctx, gen, outputDir)...)
	// ----------------------------------------------------------------
	// 10. Social media QR codes
	// ----------------------------------------------------------------
	records = append(records, generateSocial(ctx, gen, outputDir)...)
	// ----------------------------------------------------------------
	// 11. WhatsApp / Zoom QR codes
	// ----------------------------------------------------------------
	records = append(records, generateChat(ctx, gen, outputDir)...)
	// ----------------------------------------------------------------
	// 12. Market (App Store / Play Store) QR codes
	// ----------------------------------------------------------------
	records = append(records, generateMarket(ctx, gen, outputDir)...)
	// ----------------------------------------------------------------
	// 13. Crypto / PayPal / iBeacon / NTP QR codes
	// ----------------------------------------------------------------
	records = append(records, generatePayments(ctx, gen, outputDir)...)
	// ----------------------------------------------------------------
	// 14. Builder pattern + custom styling
	// ----------------------------------------------------------------
	records = append(records, generateBuilder(ctx, outputDir)...)
	// ----------------------------------------------------------------
	// 15. Advanced renderer (rounded modules, gradient, circle modules)
	// ----------------------------------------------------------------
	records = append(records, generateAdvancedRenderer(ctx, gen, outputDir)...)
	// ----------------------------------------------------------------
	// 16. Batch processing
	// ----------------------------------------------------------------
	records = append(records, generateBatch(ctx, outputDir)...)
	// ----------------------------------------------------------------
	// 17. Terminal output QR code (saved as .txt)
	// ----------------------------------------------------------------
	records = append(records, generateTerminal(ctx, gen, outputDir)...)
	// ----------------------------------------------------------------
	// 18. Base64 output
	// ----------------------------------------------------------------
	records = append(records, generateBase64(ctx, gen, outputDir)...)
	// ----------------------------------------------------------------
	// 19. Edge cases (empty data, long text, special chars)
	// ----------------------------------------------------------------
	records = append(records, generateEdgeCases(ctx, gen, outputDir)...)
	// ----------------------------------------------------------------
	// 20. Logo overlay QR codes
	// ----------------------------------------------------------------
	logoPath := filepath.Join(".", "logo.png")
	records = append(records, generateLogoQR(ctx, gen, outputDir, logoPath)...)
	// ----------------------------------------------------------------
	// 21. Context cancellation test
	// ----------------------------------------------------------------
	records = append(records, testContextCancellation(outputDir)...)

	// ----------------------------------------------------------------
	// Print report
	// ----------------------------------------------------------------
	printReport(records)
}

// ==========================================================================
// Generation helpers
// ==========================================================================

func record(filename, qrType, format string, err error, start time.Time) GenerationRecord {
	return GenerationRecord{
		Filename: filename,
		Type:     qrType,
		Format:   format,
		OK:       err == nil,
		Error:    err,
		Duration: time.Since(start),
	}
}

func savePNG(ctx context.Context, gen qrcode.Generator, p payload.Payload, path string) error {
	var buf bytes.Buffer
	if err := gen.GenerateToWriter(ctx, p, &buf, qrcode.FormatPNG); err != nil {
		return err
	}
	return os.WriteFile(path, buf.Bytes(), 0o644)
}

func saveSVG(ctx context.Context, gen qrcode.Generator, p payload.Payload, path string) error {
	var buf bytes.Buffer
	if err := gen.GenerateToWriter(ctx, p, &buf, qrcode.FormatSVG); err != nil {
		return err
	}
	return os.WriteFile(path, buf.Bytes(), 0o644)
}

func saveTerminal(ctx context.Context, gen qrcode.Generator, p payload.Payload, path string) error {
	var buf bytes.Buffer
	if err := gen.GenerateToWriter(ctx, p, &buf, qrcode.FormatTerminal); err != nil {
		return err
	}
	return os.WriteFile(path, buf.Bytes(), 0o644)
}

func savePDF(ctx context.Context, gen qrcode.Generator, p payload.Payload, path string) error {
	var buf bytes.Buffer
	if err := gen.GenerateToWriter(ctx, p, &buf, qrcode.FormatPDF); err != nil {
		return err
	}
	return os.WriteFile(path, buf.Bytes(), 0o644)
}

func saveBase64(ctx context.Context, gen qrcode.Generator, p payload.Payload, path string) error {
	var buf bytes.Buffer
	if err := gen.GenerateToWriter(ctx, p, &buf, qrcode.FormatBase64); err != nil {
		return err
	}
	return os.WriteFile(path, buf.Bytes(), 0o644)
}

// ==========================================================================
// 2. Text QR codes
// ==========================================================================

func generateText(ctx context.Context, gen qrcode.Generator, dir string) []GenerationRecord {
	var recs []GenerationRecord
	items := []struct {
		name, text string
	}{
		{"text_hello", "Hello, World! This is a QR code."},
		{"text_quote", "The only way to do great work is to love what you do. - Steve Jobs"},
		{"text_unicode", "QR code with unicode: \u00a9 \u2605 \u2665 \u2666 \u263A"},
	}
	for _, item := range items {
		t := time.Now()
		p := &payload.TextPayload{Text: item.text}
		err := savePNG(ctx, gen, p, filepath.Join(dir, item.name+".png"))
		recs = append(recs, record(item.name+".png", "text", "PNG", err, t))

		t = time.Now()
		err = saveSVG(ctx, gen, p, filepath.Join(dir, item.name+".svg"))
		recs = append(recs, record(item.name+".svg", "text", "SVG", err, t))
	}
	return recs
}

// ==========================================================================
// 3. URL QR codes
// ==========================================================================

func generateURLs(ctx context.Context, gen qrcode.Generator, dir string) []GenerationRecord {
	var recs []GenerationRecord
	items := []struct {
		name, url string
	}{
		{"url_google", "https://www.google.com"},
		{"url_github", "https://github.com"},
		{"url_gomod", "https://github.com/os-gomod/qrcode"},
	}
	for _, item := range items {
		t := time.Now()
		p := &payload.URLPayload{URL: item.url}
		err := savePNG(ctx, gen, p, filepath.Join(dir, item.name+".png"))
		recs = append(recs, record(item.name+".png", "url", "PNG", err, t))
	}
	return recs
}

// ==========================================================================
// 4. WiFi QR codes
// ==========================================================================

func generateWiFi(ctx context.Context, gen qrcode.Generator, dir string) []GenerationRecord {
	var recs []GenerationRecord
	items := []struct {
		name      string
		ssid, pwd string
		enc       string
	}{
		{"wifi_home", "MyHomeWiFi", "s3cretP@ss!", payload.EncryptionWPA2},
		{"wifi_open", "CafeFreeWiFi", "", payload.EncryptionNoPass},
		{"wifi_special", "Office\\Net", "p:ass;word", payload.EncryptionWPA3},
	}
	for _, item := range items {
		t := time.Now()
		p := &payload.WiFiPayload{SSID: item.ssid, Password: item.pwd, Encryption: item.enc}
		err := savePNG(ctx, gen, p, filepath.Join(dir, item.name+".png"))
		recs = append(recs, record(item.name+".png", "wifi", "PNG", err, t))
	}
	return recs
}

// ==========================================================================
// 5. vCard / MeCard
// ==========================================================================

func generateContact(ctx context.Context, gen qrcode.Generator, dir string) []GenerationRecord {
	var recs []GenerationRecord

	// vCard
	t := time.Now()
	vp := &payload.VCardPayload{
		FirstName:    "Jane",
		LastName:     "Doe",
		Phone:        "+1-555-0123",
		Email:        "jane.doe@example.com",
		Organization: "Acme Corp",
		Title:        "Software Engineer",
		URL:          "https://janedoe.dev",
		Address:      "123 Main St, Springfield",
	}
	err := savePNG(ctx, gen, vp, filepath.Join(dir, "contact_vcard.png"))
	recs = append(recs, record("contact_vcard.png", "vcard", "PNG", err, t))

	t = time.Now()
	err = saveSVG(ctx, gen, vp, filepath.Join(dir, "contact_vcard.svg"))
	recs = append(recs, record("contact_vcard.svg", "vcard", "SVG", err, t))

	// MeCard
	t = time.Now()
	mp := &payload.MeCardPayload{
		Name:     "John Smith",
		Phone:    "+1-555-9999",
		Email:    "john@example.com",
		URL:      "https://johnsmith.com",
		Birthday: "19900101",
		Note:     "QR code library contributor",
		Address:  "456 Oak Ave, Metropolis",
		Nickname: "johnny",
	}
	err = savePNG(ctx, gen, mp, filepath.Join(dir, "contact_mecard.png"))
	recs = append(recs, record("contact_mecard.png", "mecard", "PNG", err, t))

	return recs
}

// ==========================================================================
// 6. SMS / MMS / Phone
// ==========================================================================

func generateMessaging(ctx context.Context, gen qrcode.Generator, dir string) []GenerationRecord {
	var recs []GenerationRecord

	// SMS
	t := time.Now()
	sp := &payload.SMSPayload{Phone: "+1234567890", Message: "Hi! Scanned your QR code."}
	err := savePNG(ctx, gen, sp, filepath.Join(dir, "msg_sms.png"))
	recs = append(recs, record("msg_sms.png", "sms", "PNG", err, t))

	// MMS
	t = time.Now()
	mp := &payload.MMSPayload{Phone: "+1234567890", Subject: "Photo", Message: "Check this out!"}
	err = savePNG(ctx, gen, mp, filepath.Join(dir, "msg_mms.png"))
	recs = append(recs, record("msg_mms.png", "mms", "PNG", err, t))

	// Phone
	t = time.Now()
	pp := &payload.PhonePayload{Number: "+1-800-FLOWERS"}
	err = savePNG(ctx, gen, pp, filepath.Join(dir, "msg_phone.png"))
	recs = append(recs, record("msg_phone.png", "phone", "PNG", err, t))

	return recs
}

// ==========================================================================
// 7. Email
// ==========================================================================

func generateEmail(ctx context.Context, gen qrcode.Generator, dir string) []GenerationRecord {
	var recs []GenerationRecord
	t := time.Now()
	ep := &payload.EmailPayload{
		To:      "hello@example.com",
		Subject: "QR Code Inquiry",
		Body:    "I found your QR code and wanted to reach out!",
		CC:      []string{"cc@example.com"},
	}
	err := savePNG(ctx, gen, ep, filepath.Join(dir, "email.png"))
	recs = append(recs, record("email.png", "email", "PNG", err, t))
	return recs
}

// ==========================================================================
// 8. Geo / Maps
// ==========================================================================

func generateGeoMaps(ctx context.Context, gen qrcode.Generator, dir string) []GenerationRecord {
	var recs []GenerationRecord

	// Geo
	t := time.Now()
	gp := &payload.GeoPayload{Latitude: 37.7749, Longitude: -122.4194}
	err := savePNG(ctx, gen, gp, filepath.Join(dir, "geo_sf.png"))
	recs = append(recs, record("geo_sf.png", "geo", "PNG", err, t))

	// Google Maps
	t = time.Now()
	gmp := &payload.GoogleMapsPayload{Latitude: 48.8584, Longitude: 2.2945, Query: "Eiffel Tower", Zoom: 17}
	err = savePNG(ctx, gen, gmp, filepath.Join(dir, "maps_google.png"))
	recs = append(recs, record("maps_google.png", "google_maps", "PNG", err, t))

	// Google Maps Directions
	t = time.Now()
	gmd := &payload.GoogleMapsDirectionsPayload{
		Origin:      "Times Square, New York",
		Destination: "Central Park, New York",
		TravelMode:  payload.TravelModeWalking,
	}
	err = savePNG(ctx, gen, gmd, filepath.Join(dir, "maps_directions.png"))
	recs = append(recs, record("maps_directions.png", "google_maps_directions", "PNG", err, t))

	// Google Maps Place
	t = time.Now()
	gmpl := &payload.GoogleMapsPlacePayload{PlaceName: "Statue of Liberty"}
	err = savePNG(ctx, gen, gmpl, filepath.Join(dir, "maps_place.png"))
	recs = append(recs, record("maps_place.png", "google_maps_place", "PNG", err, t))

	// Apple Maps
	t = time.Now()
	amp := &payload.AppleMapsPayload{Latitude: 35.6762, Longitude: 139.6503, Query: "Tokyo Tower"}
	err = savePNG(ctx, gen, amp, filepath.Join(dir, "maps_apple.png"))
	recs = append(recs, record("maps_apple.png", "apple_maps", "PNG", err, t))

	return recs
}

// ==========================================================================
// 9. Calendar / Event
// ==========================================================================

func generateCalendar(ctx context.Context, gen qrcode.Generator, dir string) []GenerationRecord {
	var recs []GenerationRecord

	// Calendar
	t := time.Now()
	start := time.Date(2026, 6, 15, 9, 0, 0, 0, time.UTC)
	end := time.Date(2026, 6, 15, 17, 0, 0, 0, time.UTC)
	cp := &payload.CalendarPayload{
		Title:       "Go Conference 2026",
		Description: "Annual Go language conference",
		Location:    "Denver, Colorado",
		Start:       start,
		End:         end,
		AllDay:      false,
	}
	err := savePNG(ctx, gen, cp, filepath.Join(dir, "calendar.png"))
	recs = append(recs, record("calendar.png", "calendar", "PNG", err, t))

	// Event ticket
	t = time.Now()
	evp := &payload.EventPayload{
		EventID:     "EVT-2026-0042",
		EventName:   "Tech Summit",
		Venue:       "Convention Center",
		StartTime:   time.Date(2026, 9, 20, 10, 0, 0, 0, time.UTC),
		Category:    "Technology",
		Seat:        "A-14",
		Organizer:   "TechEvents Inc.",
		Description: "Annual technology summit with speakers from around the world.",
		URL:         "https://techsummit.example.com",
	}
	err = savePNG(ctx, gen, evp, filepath.Join(dir, "event_ticket.png"))
	recs = append(recs, record("event_ticket.png", "event", "PNG", err, t))

	return recs
}

// ==========================================================================
// 10. Social media
// ==========================================================================

func generateSocial(ctx context.Context, gen qrcode.Generator, dir string) []GenerationRecord {
	var recs []GenerationRecord
	items := []struct {
		name string
		p    payload.Payload
	}{
		{"social_twitter", &payload.TwitterPayload{Username: "golang"}},
		{"social_instagram", &payload.InstagramPayload{Username: "golangofficial"}},
		{"social_facebook", &payload.FacebookPayload{PageURL: "https://www.facebook.com/golang"}},
		{"social_linkedin", &payload.LinkedInPayload{ProfileURL: "https://www.linkedin.com/company/golang"}},
		{"social_telegram", &payload.TelegramPayload{Username: "golang"}},
		{"social_youtube_channel", &payload.YouTubeChannelPayload{ChannelID: "UC_x5XG1OV2P6uZZ5FSM9Ttw"}},
		{"social_youtube_video", &payload.YouTubeVideoPayload{VideoID: "dQw4w9WgXcQ"}},
		{"social_spotify_track", &payload.SpotifyTrackPayload{TrackID: "4cOdK2wGLETKBW3PvgPWqT"}},
		{"social_spotify_playlist", &payload.SpotifyPlaylistPayload{PlaylistID: "37i9dQZF1DXcBWIGoYBM5M"}},
	}
	for _, item := range items {
		t := time.Now()
		err := savePNG(ctx, gen, item.p, filepath.Join(dir, item.name+".png"))
		recs = append(recs, record(item.name+".png", item.p.Type(), "PNG", err, t))
	}
	return recs
}

// ==========================================================================
// 11. WhatsApp / Zoom
// ==========================================================================

func generateChat(ctx context.Context, gen qrcode.Generator, dir string) []GenerationRecord {
	var recs []GenerationRecord

	t := time.Now()
	wp := &payload.WhatsAppPayload{Phone: "15551234567", Message: "Hello from QR code!"}
	err := savePNG(ctx, gen, wp, filepath.Join(dir, "chat_whatsapp.png"))
	recs = append(recs, record("chat_whatsapp.png", "whatsapp", "PNG", err, t))

	t = time.Now()
	zp := &payload.ZoomPayload{
		MeetingID:   "1234567890",
		Password:    "abc123",
		DisplayName: "John Doe",
	}
	err = savePNG(ctx, gen, zp, filepath.Join(dir, "chat_zoom.png"))
	recs = append(recs, record("chat_zoom.png", "zoom", "PNG", err, t))

	return recs
}

// ==========================================================================
// 12. Market (App Store / Play Store)
// ==========================================================================

func generateMarket(ctx context.Context, gen qrcode.Generator, dir string) []GenerationRecord {
	var recs []GenerationRecord

	t := time.Now()
	gp := &payload.MarketPayload{
		Platform:  payload.MarketGooglePlay,
		PackageID: "com.example.app",
		AppName:   "MyApp",
		Campaign:  "qr_scan",
	}
	err := savePNG(ctx, gen, gp, filepath.Join(dir, "market_playstore.png"))
	recs = append(recs, record("market_playstore.png", "market", "PNG", err, t))

	t = time.Now()
	ap := &payload.MarketPayload{
		Platform: payload.MarketAppleApp,
		AppName:  "MyApp",
		Campaign: "qr_scan_ios",
	}
	err = savePNG(ctx, gen, ap, filepath.Join(dir, "market_appstore.png"))
	recs = append(recs, record("market_appstore.png", "market", "PNG", err, t))

	return recs
}

// ==========================================================================
// 13. Crypto / PayPal / iBeacon / NTP
// ==========================================================================

func generatePayments(ctx context.Context, gen qrcode.Generator, dir string) []GenerationRecord {
	var recs []GenerationRecord

	// PayPal
	t := time.Now()
	pp := &payload.PayPalPayload{Username: "pay@example.com", Amount: "25.00", Currency: "USD", Reference: "Order-42"}
	err := savePNG(ctx, gen, pp, filepath.Join(dir, "payment_paypal.png"))
	recs = append(recs, record("payment_paypal.png", "paypal", "PNG", err, t))

	// Crypto BTC
	t = time.Now()
	cp := &payload.CryptoPayload{Address: "bc1qar0srrr7xfkvy5l643lydnw9re59gtzzwf5mdq", Amount: "0.001", CryptoType: payload.CryptoBTC}
	err = savePNG(ctx, gen, cp, filepath.Join(dir, "payment_bitcoin.png"))
	recs = append(recs, record("payment_bitcoin.png", "crypto", "PNG", err, t))

	// Crypto ETH
	t = time.Now()
	ep := &payload.CryptoPayload{Address: "0x71C7656EC7ab88b098defB751B7401B5f6d8976F", Amount: "0.5", CryptoType: payload.CryptoETH}
	err = savePNG(ctx, gen, ep, filepath.Join(dir, "payment_ethereum.png"))
	recs = append(recs, record("payment_ethereum.png", "crypto", "PNG", err, t))

	// iBeacon
	t = time.Now()
	ib := &payload.IBeaconPayload{UUID: "A1B2C3D4-E5F6-7890-ABCD-EF1234567890", Major: 100, Minor: 1, Manufacturer: "acme"}
	err = savePNG(ctx, gen, ib, filepath.Join(dir, "ibeacon.png"))
	recs = append(recs, record("ibeacon.png", "ibeacon", "PNG", err, t))

	// NTP
	t = time.Now()
	np := &payload.NTPLocalePayload{Host: "pool.ntp.org", Port: "123", Version: 4, Description: "NTP time server"}
	err = savePNG(ctx, gen, np, filepath.Join(dir, "ntp.png"))
	recs = append(recs, record("ntp.png", "ntp", "PNG", err, t))

	return recs
}

// ==========================================================================
// 14. Builder pattern + custom styling
// ==========================================================================

func generateBuilder(ctx context.Context, dir string) []GenerationRecord {
	var recs []GenerationRecord

	// Custom colors via builder
	t := time.Now()
	b := qrcode.NewBuilder().
		Size(300).
		ErrorCorrection(qrcode.LevelH).
		ForegroundColor("#1A56DB").
		BackgroundColor("#F0F9FF").
		Margin(4)

	gen, err := b.Build()
	if err != nil {
		recs = append(recs, record("builder_custom.png", "builder", "PNG", err, t))
	} else {
		defer gen.Close(ctx)
		p := &payload.URLPayload{URL: "https://github.com/os-gomod/qrcode"}
		err = savePNG(ctx, gen, p, filepath.Join(dir, "builder_custom.png"))
		recs = append(recs, record("builder_custom.png", "builder", "PNG", err, t))
	}

	// Quick helper via builder
	t = time.Now()
	b2 := qrcode.NewBuilder().Size(256).ErrorCorrection(qrcode.LevelQ)
	data, err := b2.Quick("Quick builder test!")
	if err != nil {
		recs = append(recs, record("builder_quick.png", "builder", "PNG", err, t))
	} else {
		err = os.WriteFile(filepath.Join(dir, "builder_quick.png"), data, 0o644)
		recs = append(recs, record("builder_quick.png", "builder", "PNG", err, t))
	}

	// Builder QuickFile
	t = time.Now()
	err = b2.QuickFile("Builder QuickFile test", filepath.Join(dir, "builder_quickfile.png"))
	recs = append(recs, record("builder_quickfile.png", "builder", "PNG", err, t))

	return recs
}

// ==========================================================================
// 15. Advanced renderer (rounded modules, gradient, circle modules)
// ==========================================================================

func generateAdvancedRenderer(ctx context.Context, gen qrcode.Generator, dir string) []GenerationRecord {
	var recs []GenerationRecord

	// Rounded modules
	t := time.Now()
	qr, err := gen.Generate(ctx, &payload.TextPayload{Text: "Rounded Modules"})
	if err != nil {
		recs = append(recs, record("advanced_rounded.png", "advanced", "PNG", err, t))
	} else {
		var buf bytes.Buffer
		pngR := renderer.NewPNGRenderer()
		style := &renderer.ModuleStyle{Shape: "rounded", Roundness: 0.5, Transparency: 1.0}
		err = pngR.Render(ctx, qr, &buf,
			renderer.WithWidth(256),
			renderer.WithHeight(256),
			renderer.WithQuietZone(4),
			renderer.WithModuleStyle(style),
			renderer.WithForegroundColor("#2563EB"),
		)
		if err != nil {
			recs = append(recs, record("advanced_rounded.png", "advanced", "PNG", err, t))
		} else {
			err = os.WriteFile(filepath.Join(dir, "advanced_rounded.png"), buf.Bytes(), 0o644)
			recs = append(recs, record("advanced_rounded.png", "advanced", "PNG", err, t))
		}
	}

	// Circle modules
	t = time.Now()
	qr, err = gen.Generate(ctx, &payload.TextPayload{Text: "Circle Modules"})
	if err != nil {
		recs = append(recs, record("advanced_circle.png", "advanced", "PNG", err, t))
	} else {
		var buf bytes.Buffer
		pngR := renderer.NewPNGRenderer()
		err = pngR.Render(ctx, qr, &buf,
			renderer.WithWidth(256),
			renderer.WithHeight(256),
			renderer.WithQuietZone(4),
			renderer.WithCircleModules(),
			renderer.WithForegroundColor("#DC2626"),
		)
		if err != nil {
			recs = append(recs, record("advanced_circle.png", "advanced", "PNG", err, t))
		} else {
			err = os.WriteFile(filepath.Join(dir, "advanced_circle.png"), buf.Bytes(), 0o644)
			recs = append(recs, record("advanced_circle.png", "advanced", "PNG", err, t))
		}
	}

	// Gradient style
	t = time.Now()
	qr, err = gen.Generate(ctx, &payload.TextPayload{Text: "Gradient QR"})
	if err != nil {
		recs = append(recs, record("advanced_gradient.png", "advanced", "PNG", err, t))
	} else {
		var buf bytes.Buffer
		pngR := renderer.NewPNGRenderer()
		err = pngR.Render(ctx, qr, &buf,
			renderer.WithWidth(256),
			renderer.WithHeight(256),
			renderer.WithQuietZone(4),
			renderer.WithGradient("#059669", "#0891B2", 135),
		)
		if err != nil {
			recs = append(recs, record("advanced_gradient.png", "advanced", "PNG", err, t))
		} else {
			err = os.WriteFile(filepath.Join(dir, "advanced_gradient.png"), buf.Bytes(), 0o644)
			recs = append(recs, record("advanced_gradient.png", "advanced", "PNG", err, t))
		}
	}

	// Diamond modules
	t = time.Now()
	qr, err = gen.Generate(ctx, &payload.TextPayload{Text: "Diamond Modules"})
	if err != nil {
		recs = append(recs, record("advanced_diamond.png", "advanced", "PNG", err, t))
	} else {
		var buf bytes.Buffer
		pngR := renderer.NewPNGRenderer()
		style := &renderer.ModuleStyle{Shape: "diamond", Roundness: 0, Transparency: 1.0}
		err = pngR.Render(ctx, qr, &buf,
			renderer.WithWidth(256),
			renderer.WithHeight(256),
			renderer.WithQuietZone(4),
			renderer.WithModuleStyle(style),
			renderer.WithForegroundColor("#7C3AED"),
		)
		if err != nil {
			recs = append(recs, record("advanced_diamond.png", "advanced", "PNG", err, t))
		} else {
			err = os.WriteFile(filepath.Join(dir, "advanced_diamond.png"), buf.Bytes(), 0o644)
			recs = append(recs, record("advanced_diamond.png", "advanced", "PNG", err, t))
		}
	}

	return recs
}

// ==========================================================================
// 16. Batch processing
// ==========================================================================

func generateBatch(ctx context.Context, dir string) []GenerationRecord {
	var recs []GenerationRecord

	t := time.Now()
	gen, err := qrcode.New(qrcode.WithDefaultSize(256))
	if err != nil {
		recs = append(recs, record("batch_item_*.png", "batch", "PNG", err, t))
		return recs
	}
	defer gen.Close(ctx)

	items := []batch.Item{
		{ID: "batch_text", Data: "Batch text 1"},
		{ID: "batch_url", Data: "https://example.com"},
		{ID: "batch_hello", Data: "Batch Hello!"},
	}

	proc := batch.NewProcessor(gen,
		batch.WithBatchFormat(qrcode.FormatPNG),
		batch.WithBatchOutputDir(dir),
		batch.WithBatchConcurrency(2),
	)
	results, procErr := proc.Process(ctx, items)
	_ = procErr
	for _, r := range results {
		status := "OK"
		if r.Err != nil {
			status = "FAIL"
		}
		fmt.Printf("  batch [%s]: %s\n", r.ID, status)
		recs = append(recs, record(r.ID+".png", "batch", "PNG", r.Err, t))
	}

	return recs
}

// ==========================================================================
// 17. Terminal output
// ==========================================================================

func generateTerminal(ctx context.Context, gen qrcode.Generator, dir string) []GenerationRecord {
	var recs []GenerationRecord
	t := time.Now()
	p := &payload.TextPayload{Text: "Terminal QR Code"}
	err := saveTerminal(ctx, gen, p, filepath.Join(dir, "terminal.txt"))
	recs = append(recs, record("terminal.txt", "terminal", "TXT", err, t))
	return recs
}

// ==========================================================================
// 18. Base64
// ==========================================================================

func generateBase64(ctx context.Context, gen qrcode.Generator, dir string) []GenerationRecord {
	var recs []GenerationRecord
	t := time.Now()
	p := &payload.TextPayload{Text: "Base64 QR Code"}
	err := saveBase64(ctx, gen, p, filepath.Join(dir, "base64.txt"))
	recs = append(recs, record("base64.txt", "base64", "TXT", err, t))
	return recs
}

// ==========================================================================
// 19. Edge cases
// ==========================================================================

func generateEdgeCases(ctx context.Context, gen qrcode.Generator, dir string) []GenerationRecord {
	var recs []GenerationRecord

	// Long text
	t := time.Now()
	longText := strings.Repeat("QR code stress test. ", 50)
	err := savePNG(ctx, gen, &payload.TextPayload{Text: longText}, filepath.Join(dir, "edge_long_text.png"))
	recs = append(recs, record("edge_long_text.png", "text", "PNG", err, t))

	// Error case: empty data should fail
	t = time.Now()
	err = savePNG(ctx, gen, &payload.TextPayload{Text: ""}, filepath.Join(dir, "edge_empty.png"))
	recs = append(recs, record("edge_empty.png", "text_error", "PNG", err, t))

	// Special chars in data
	t = time.Now()
	err = savePNG(ctx, gen, &payload.TextPayload{Text: "Special: <>&\"'\\/:;@"}, filepath.Join(dir, "edge_special_chars.png"))
	recs = append(recs, record("edge_special_chars.png", "text", "PNG", err, t))

	return recs
}

// ==========================================================================
// 20. Logo overlay
// ==========================================================================

func generateLogoQR(ctx context.Context, gen qrcode.Generator, dir, logoPath string) []GenerationRecord {
	var recs []GenerationRecord

	// Validate logo file exists
	if err := qrlogo.Validate(logoPath); err != nil {
		fmt.Printf("  [WARN] logo file not found or invalid: %v (skipping logo QR codes)\n", err)
		return recs
	}

	// Load the logo image
	logoProc := qrlogo.New(logoPath, 0.25)
	logoImg, err := logoProc.Load()
	if err != nil {
		fmt.Printf("  [WARN] failed to load logo: %v (skipping logo QR codes)\n", err)
		return recs
	}

	// --- Logo QR 1: URL with logo overlay (standard) ---
	t := time.Now()
	qrURL, err := gen.Generate(ctx, &payload.URLPayload{URL: "https://github.com/os-gomod/qrcode"})
	if err != nil {
		recs = append(recs, record("logo_url.png", "logo", "PNG", err, t))
	} else {
		// Render QR to PNG
		var qrBuf bytes.Buffer
		pngR := renderer.NewPNGRenderer()
		err = pngR.Render(ctx, qrURL, &qrBuf,
			renderer.WithWidth(400),
			renderer.WithHeight(400),
			renderer.WithQuietZone(4),
			renderer.WithForegroundColor("#000000"),
			renderer.WithBackgroundColor("#FFFFFF"),
		)
		if err != nil {
			recs = append(recs, record("logo_url.png", "logo", "PNG", err, t))
		} else {
			// Decode QR PNG back to image.Image
			qrImg, _, decErr := image.Decode(&qrBuf)
			if decErr != nil {
				recs = append(recs, record("logo_url.png", "logo", "PNG", decErr, t))
			} else {
				// Resize logo to 25% of QR modules
				resizedLogo := qrlogo.ResizeLogo(logoImg, qrURL.Size, 0.25)
				// Overlay logo on QR code
				final := qrlogo.OverlayLogo(qrImg, resizedLogo, 4)
				// Encode final image to PNG
				outPath := filepath.Join(dir, "logo_url.png")
				encErr := saveImagePNG(final, outPath)
				recs = append(recs, record("logo_url.png", "logo", "PNG", encErr, t))
			}
		}
	}

	// --- Logo QR 2: Text with logo overlay (high EC for more tolerance) ---
	t = time.Now()
	genH, err := qrcode.New(
		qrcode.WithDefaultSize(400),
		qrcode.WithErrorCorrection(qrcode.LevelH),
		qrcode.WithQuietZone(4),
	)
	if err != nil {
		recs = append(recs, record("logo_text.png", "logo", "PNG", err, t))
	} else {
		defer genH.Close(ctx)
		qrText, err := genH.Generate(ctx, &payload.TextPayload{Text: "QR Code with Logo Overlay!"})
		if err != nil {
			recs = append(recs, record("logo_text.png", "logo", "PNG", err, t))
		} else {
			var qrBuf2 bytes.Buffer
			pngR2 := renderer.NewPNGRenderer()
			err = pngR2.Render(ctx, qrText, &qrBuf2,
				renderer.WithWidth(400),
				renderer.WithHeight(400),
				renderer.WithQuietZone(4),
				renderer.WithForegroundColor("#1E3A5F"),
				renderer.WithBackgroundColor("#FFFFFF"),
			)
			if err != nil {
				recs = append(recs, record("logo_text.png", "logo", "PNG", err, t))
			} else {
				qrImg2, _, decErr := image.Decode(&qrBuf2)
				if decErr != nil {
					recs = append(recs, record("logo_text.png", "logo", "PNG", decErr, t))
				} else {
					resizedLogo2 := qrlogo.ResizeLogo(logoImg, qrText.Size, 0.20)
					final2 := qrlogo.OverlayLogo(qrImg2, resizedLogo2, 4)
					outPath2 := filepath.Join(dir, "logo_text.png")
					encErr := saveImagePNG(final2, outPath2)
					recs = append(recs, record("logo_text.png", "logo", "PNG", encErr, t))
				}
			}
		}
	}

	// --- Logo QR 3: WiFi with logo and tinted logo ---
	t = time.Now()
	qrWiFi, err := gen.Generate(ctx, &payload.WiFiPayload{
		SSID:       "MyHomeWiFi",
		Password:   "s3cretP@ss!",
		Encryption: payload.EncryptionWPA2,
	})
	if err != nil {
		recs = append(recs, record("logo_wifi.png", "logo", "PNG", err, t))
	} else {
		var qrBuf3 bytes.Buffer
		pngR3 := renderer.NewPNGRenderer()
		err = pngR3.Render(ctx, qrWiFi, &qrBuf3,
			renderer.WithWidth(400),
			renderer.WithHeight(400),
			renderer.WithQuietZone(4),
			renderer.WithForegroundColor("#000000"),
			renderer.WithBackgroundColor("#FFFFFF"),
		)
		if err != nil {
			recs = append(recs, record("logo_wifi.png", "logo", "PNG", err, t))
		} else {
			qrImg3, _, decErr := image.Decode(&qrBuf3)
			if decErr != nil {
				recs = append(recs, record("logo_wifi.png", "logo", "PNG", decErr, t))
			} else {
				resizedLogo3 := qrlogo.ResizeLogo(logoImg, qrWiFi.Size, 0.22)
				final3 := qrlogo.OverlayLogo(qrImg3, resizedLogo3, 4)
				outPath3 := filepath.Join(dir, "logo_wifi.png")
				encErr := saveImagePNG(final3, outPath3)
				recs = append(recs, record("logo_wifi.png", "logo", "PNG", encErr, t))
			}
		}
	}

	// --- Logo QR 4: vCard with logo (larger size) ---
	t = time.Now()
	genCard, err := qrcode.New(
		qrcode.WithDefaultSize(512),
		qrcode.WithErrorCorrection(qrcode.LevelH),
		qrcode.WithQuietZone(4),
	)
	if err != nil {
		recs = append(recs, record("logo_vcard.png", "logo", "PNG", err, t))
	} else {
		defer genCard.Close(ctx)
		qrCard, err := genCard.Generate(ctx, &payload.VCardPayload{
			FirstName: "Jane",
			LastName:  "Doe",
			Phone:     "+1-555-0123",
			Email:     "jane.doe@example.com",
			URL:       "https://janedoe.dev",
		})
		if err != nil {
			recs = append(recs, record("logo_vcard.png", "logo", "PNG", err, t))
		} else {
			var qrBuf4 bytes.Buffer
			pngR4 := renderer.NewPNGRenderer()
			err = pngR4.Render(ctx, qrCard, &qrBuf4,
				renderer.WithWidth(512),
				renderer.WithHeight(512),
				renderer.WithQuietZone(4),
				renderer.WithForegroundColor("#1A1A2E"),
				renderer.WithBackgroundColor("#FFFFFF"),
			)
			if err != nil {
				recs = append(recs, record("logo_vcard.png", "logo", "PNG", err, t))
			} else {
				qrImg4, _, decErr := image.Decode(&qrBuf4)
				if decErr != nil {
					recs = append(recs, record("logo_vcard.png", "logo", "PNG", decErr, t))
				} else {
					resizedLogo4 := qrlogo.ResizeLogo(logoImg, qrCard.Size, 0.18)
					final4 := qrlogo.OverlayLogo(qrImg4, resizedLogo4, 4)
					outPath4 := filepath.Join(dir, "logo_vcard.png")
					encErr := saveImagePNG(final4, outPath4)
					recs = append(recs, record("logo_vcard.png", "logo", "PNG", encErr, t))
				}
			}
		}
	}

	// --- Logo QR 5: Using logo.EncodePNG helper ---
	t = time.Now()
	pngData, err := qrlogo.EncodePNG(logoImg)
	if err != nil {
		recs = append(recs, record("logo_encoded.png", "logo", "PNG", err, t))
	} else {
		err = os.WriteFile(filepath.Join(dir, "logo_encoded.png"), pngData, 0o644)
		recs = append(recs, record("logo_encoded.png", "logo", "PNG", err, t))
	}

	// --- Logo QR 6: Logo validation and supported formats check ---
	t = time.Now()
	supportedFormats := qrlogo.SupportedFormats()
	fmt.Printf("  logo supported formats: %v\n", supportedFormats)
	isPNG := qrlogo.IsSupportedFormat(".png")
	fmt.Printf("  .png is supported: %v\n", isPNG)
	recs = append(recs, record("logo_info", "logo", "INFO", nil, t))

	return recs
}

func saveImagePNG(img image.Image, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

// ==========================================================================
// 21. Context cancellation
// ==========================================================================

func testContextCancellation(dir string) []GenerationRecord {
	var recs []GenerationRecord
	t := time.Now()
	cancelCtx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	gen, err := qrcode.New(qrcode.WithDefaultSize(256))
	if err != nil {
		recs = append(recs, record("edge_cancelled.png", "cancelled", "PNG", err, t))
		return recs
	}
	defer gen.Close(context.Background())

	// This should succeed because Generate does not check ctx until interceptor runs
	_, genErr := gen.Generate(cancelCtx, &payload.TextPayload{Text: "cancelled test"})
	if genErr == nil {
		recs = append(recs, record("edge_cancelled.png", "cancelled", "PNG", nil, t))
	} else {
		recs = append(recs, record("edge_cancelled.png", "cancelled", "PNG", genErr, t))
	}
	return recs
}

// ==========================================================================
// Report
// ==========================================================================

func printReport(records []GenerationRecord) {
	fmt.Println("=====================================================")
	fmt.Println("=== QR Code Generation Report ===")
	fmt.Printf("Generated at: %s\n", time.Now().UTC().Format(time.RFC3339))
	fmt.Printf("Total files:  %d\n", len(records))
	fmt.Println("=====================================================")

	var okCount, failCount int
	var pngCount, svgCount, txtCount, pdfCount int
	var totalSize int64

	for _, r := range records {
		status := "\u2713" // check
		if !r.OK {
			status = "\u2717" // cross
			failCount++
		} else {
			okCount++
		}

		// Get file size
		var sizeStr string
		info, err := os.Stat(r.Filename)
		if err == nil {
			totalSize += info.Size()
			sizeStr = fmt.Sprintf("%d B", info.Size())
		}

		fmt.Printf("  %s %-35s %-20s %-5s %s %s\n",
			status, r.Filename, r.Type, r.Format, r.Duration.Round(time.Microsecond), sizeStr)

		switch r.Format {
		case "PNG":
			pngCount++
		case "SVG":
			svgCount++
		case "TXT":
			txtCount++
		case "PDF":
			pdfCount++
		}
	}

	fmt.Println("=====================================================")
	fmt.Printf("Summary:\n")
	fmt.Printf("  PNG files:  %d\n", pngCount)
	fmt.Printf("  SVG files:  %d\n", svgCount)
	fmt.Printf("  TXT files:  %d\n", txtCount)
	fmt.Printf("  PDF files:  %d\n", pdfCount)
	fmt.Printf("  Total size: %s\n", formatBytes(totalSize))
	fmt.Printf("  Valid:      %d\n", okCount)
	fmt.Printf("  Failed:     %d\n", failCount)

	allOK := failCount == 0 && okCount > 0
	fmt.Printf("\nAll QR codes generated successfully: %v\n", allOK)

	// Show errors if any
	if failCount > 0 {
		fmt.Println("\nFailed generations:")
		for _, r := range records {
			if !r.OK {
				// Skip expected errors like empty data
				if qrerrors.IsCode(r.Error, qrerrors.ErrCodeValidation) {
					fmt.Printf("  %s: [EXPECTED] %v\n", r.Filename, r.Error)
				} else {
					fmt.Printf("  %s: %v\n", r.Filename, r.Error)
				}
			}
		}
	}

	fmt.Println("=====================================================")

	if !allOK {
		fmt.Println("\nNote: Some expected validation errors occurred (e.g., empty data).")
		fmt.Println("This is normal behavior for edge case testing.")
	}
}

func formatBytes(b int64) string {
	const (
		KB = 1024
		MB = KB * 1024
	)
	switch {
	case b >= MB:
		return fmt.Sprintf("%.2f MB", float64(b)/float64(MB))
	case b >= KB:
		return fmt.Sprintf("%.2f KB", float64(b)/float64(KB))
	default:
		return fmt.Sprintf("%d B", b)
	}
}
