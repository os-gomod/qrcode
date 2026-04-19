// Package payload provides 30+ typed payload structures for encoding various
// kinds of data into QR codes. Each payload type implements the Payload interface
// and produces a standards-compliant string representation suitable for embedding
// in a QR code symbol.
//
// The package covers a wide range of QR code use cases:
//   - Plain text and URLs
//   - WiFi network credentials (WIFI:T:WPA;S:...;P:...;; format)
//   - Electronic business cards (vCard 2.1/3.0/4.0 and MeCard formats)
//   - Messaging: SMS, MMS, WhatsApp, email (mailto: / smsto: URIs)
//   - Phone numbers (tel: URI)
//   - Geographic locations (geo: URI, Google Maps, Apple Maps)
//   - Calendar events (iCalendar VEVENT format)
//   - Social media profile links (Twitter/X, LinkedIn, Instagram, Facebook, Telegram, YouTube)
//   - Media streaming links (Spotify tracks/playlists, Apple Music, YouTube videos)
//   - App store download links (Google Play, Apple App Store)
//   - Payments: PayPal.me links and cryptocurrency payment URIs (BTC, ETH, LTC)
//   - Swiss QR-bill payment instructions (PID format)
//   - NTP time server configurations (ntp:// URI)
//   - Video conferencing: Zoom meeting join links
//   - Bluetooth beacon configurations: iBeacon (beacon registry URL format)
//
// # Usage
//
// Every payload type offers both a convenience builder function and direct struct
// construction. Builder functions (e.g. Text, URL, WiFi, Contact, Event) validate
// their inputs before returning, making them the recommended approach:
//
//	// Plain text
//	p, err := payload.Text("Hello, world!")
//	data, err := p.Encode() // "Hello, world!"
//
//	// URL with optional title fragment
//	p, err := payload.URL("https://example.com")
//	data, err := p.Encode() // "https://example.com"
//
//	// WiFi credentials
//	p, err := payload.WiFi("MyNetwork", "s3cret", "WPA2")
//	data, err := p.Encode() // "WIFI:T:WPA2;S:MyNetwork;P:s3cret;;"
//
//	// vCard contact with functional options
//	p, err := payload.Contact("Jane", "Doe",
//	    payload.WithPhone("+1-555-0123"),
//	    payload.WithEmail("jane@example.com"),
//	    payload.WithOrganization("Acme Inc"),
//	)
//	data, err := p.Encode() // "BEGIN:VCARD\r\nVERSION:3.0\r\n..."
//
//	// Calendar event with options
//	p, err := payload.Event("Team Standup", "Room 42", start, end,
//	    payload.WithAllDay(),
//	    payload.WithDescription("Weekly sync"),
//	)
//	data, err := p.Encode() // "BEGIN:VEVENT\r\nSUMMARY:Team Standup\r\n..."
//
// # The Payload Interface
//
// All payload types satisfy the Payload interface, which requires four methods:
//   - Encode() (string, error) — returns the encoded QR code data string
//   - Validate() error — checks that all fields are well-formed
//   - Type() string — returns a short identifier (e.g. "text", "url", "wifi")
//   - Size() int — returns the byte length of the encoded data
//
// BasePayload provides default no-op implementations for Type and Size, which
// individual payload types may embed and override as needed.
package payload

// Payload is the interface that all QR code data types must implement.
//
// Each method serves a specific purpose in the payload lifecycle:
//   - Encode produces the final string that gets embedded in the QR code symbol.
//   - Validate checks structural integrity (e.g. non-empty fields, valid ranges).
//   - Type identifies the payload category for logging, serialization, or routing.
//   - Size reports the encoded byte length, useful for QR version selection.
type Payload interface {
	// Encode returns the encoded string representation of the payload data.
	// The returned string is ready to be passed directly to a QR code encoder.
	Encode() (string, error)
	// Type returns a short identifier for the payload kind (e.g. "text", "url").
	Type() string
	// Validate checks that the payload fields are well-formed.
	Validate() error
	// Size returns the length of the encoded data in bytes.
	Size() int
}

// BasePayload provides default no-op implementations for the Payload interface.
// Individual payload types may embed BasePayload and override only the methods
// they need, reducing boilerplate for Type() and Size().
type BasePayload struct{}

// Type returns "unknown".
func (b *BasePayload) Type() string {
	return "unknown"
}

// Size returns 0.
func (b *BasePayload) Size() int {
	return 0
}
