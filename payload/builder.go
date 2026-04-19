// Package payload provides convenience builder functions for constructing validated
// payload instances. Each builder validates its inputs before returning, making
// them the recommended way to create payload objects.
//
// Builder functions that accept variable options (Contact, Event) use the
// functional options pattern with ContactOption and EventOption respectively.
package payload

import (
	"fmt"
	"time"
)

// Text creates a validated TextPayload for encoding arbitrary plain text into
// a QR code. The text must be non-empty and at most 4296 characters (the
// maximum data capacity of a QR code at version 40 with low error correction).
//
// Example:
//
//	p, err := payload.Text("Hello, world!")
//	if err != nil { /* handle error */ }
//	data, _ := p.Encode() // "Hello, world!"
func Text(text string) (*TextPayload, error) {
	p := &TextPayload{Text: text}
	if err := p.Validate(); err != nil {
		return nil, err
	}
	return p, nil
}

// URL creates a validated URLPayload for encoding a web link into a QR code.
// The URL is normalized to HTTPS if no scheme is provided. Only http and https
// schemes are supported.
//
// Example:
//
//	p, err := payload.URL("https://example.com")
//	data, _ := p.Encode() // "https://example.com"
//
//	p, err = payload.URL("example.com")
//	data, _ = p.Encode() // "https://example.com" (auto-prefixed)
func URL(rawURL string) (*URLPayload, error) {
	p := &URLPayload{URL: rawURL}
	if err := p.Validate(); err != nil {
		return nil, err
	}
	return p, nil
}

// Email creates a validated EmailPayload encoded as a RFC 6068 mailto: URI.
// The subject, body, and CC addresses are URL-encoded in the query string.
//
// Example:
//
//	p, err := payload.Email("alice@example.com", "Hello", "World", "bob@example.com")
//	data, _ := p.Encode() // "mailto:alice@example.com?subject=Hello&body=World&cc=bob@example.com"
func Email(to, subject, body string, cc ...string) (*EmailPayload, error) {
	p := &EmailPayload{
		To:      to,
		Subject: subject,
		Body:    body,
		CC:      cc,
	}
	if err := p.Validate(); err != nil {
		return nil, err
	}
	return p, nil
}

// SMS creates a validated SMSPayload encoded as an smsto: URI.
// If a message is provided, it is appended after a colon separator.
//
// Example:
//
//	p, err := payload.SMS("+14155552671", "Hi there!")
//	data, _ := p.Encode() // "smsto:+14155552671:Hi there!"
func SMS(phone, message string) (*SMSPayload, error) {
	p := &SMSPayload{
		Phone:   phone,
		Message: message,
	}
	if err := p.Validate(); err != nil {
		return nil, err
	}
	return p, nil
}

// WhatsApp creates a validated WhatsAppPayload that encodes a wa.me chat link.
// The phone number is cleaned to digits only (leading + is stripped).
// An optional pre-filled message is appended as a ?text= query parameter.
//
// Example:
//
//	p, err := payload.Whatsapp("+1-415-555-2671", "Ready?")
//	data, _ := p.Encode() // "https://wa.me/14155552671?text=Ready%3F"
func WhatsApp(phone, message string) (*WhatsAppPayload, error) {
	p := &WhatsAppPayload{
		Phone:   phone,
		Message: message,
	}
	if err := p.Validate(); err != nil {
		return nil, err
	}
	return p, nil
}

// Phone creates a validated PhonePayload encoded as a tel: URI (RFC 3966).
// The phone number must contain at least one digit.
//
// Example:
//
//	p, err := payload.Phone("+1-555-0123")
//	data, _ := p.Encode() // "tel:+1-555-0123"
func Phone(number string) (*PhonePayload, error) {
	p := &PhonePayload{Number: number}
	if err := p.Validate(); err != nil {
		return nil, err
	}
	return p, nil
}

// WiFi creates a validated WiFiPayload using the standard WIFI:T:...;S:...;P:...;;
// QR code format recognized by most smartphones. Special characters in the SSID
// and password are escaped using hex encoding (\XX).
//
// The encryption parameter must be one of: WEP, WPA, WPA2, WPA3, SAE, or nopass.
// For open networks, use "nopass" and an empty password.
//
// Example:
//
//	p, err := payload.WiFi("MyNetwork", "s3cret", "WPA2")
//	data, _ := p.Encode() // "WIFI:T:WPA2;S:MyNetwork;P:s3cret;;"
func WiFi(ssid, password, encryption string) (*WiFiPayload, error) {
	p := &WiFiPayload{
		SSID:       ssid,
		Password:   password,
		Encryption: encryption,
	}
	if err := p.Validate(); err != nil {
		return nil, err
	}
	return p, nil
}

// WiFiWithHidden creates a validated WiFiPayload with the hidden SSID flag set.
// This adds ;H:true to the encoded string, indicating that the network does not
// broadcast its SSID.
//
// Example:
//
//	p, err := payload.WiFiWithHidden("HiddenNet", "p@ss", "WPA3")
//	data, _ := p.Encode() // "WIFI:T:WPA3;S:HiddenNet;P:p@ss;H:true;;"
func WiFiWithHidden(ssid, password, encryption string) (*WiFiPayload, error) {
	p := &WiFiPayload{
		SSID:       ssid,
		Password:   password,
		Encryption: encryption,
		Hidden:     true,
	}
	if err := p.Validate(); err != nil {
		return nil, err
	}
	return p, nil
}

// Contact creates a validated VCardPayload (vCard 3.0 by default) with optional
// functional options for adding phone, email, organization, title, address, URL,
// and notes. The vCard format follows RFC 6350 (with backward compatibility for
// versions 2.1 and 3.0).
//
// Example:
//
//	p, err := payload.Contact("Jane", "Doe",
//	    payload.WithPhone("+1-555-0123"),
//	    payload.WithEmail("jane@example.com"),
//	    payload.WithOrganization("Acme Inc"),
//	)
//	data, _ := p.Encode() // "BEGIN:VCARD\r\nVERSION:3.0\r\nN:Doe;Jane\r\n..."
func Contact(firstName, lastName string, opts ...ContactOption) (*VCardPayload, error) {
	p := &VCardPayload{
		FirstName: firstName,
		LastName:  lastName,
	}
	for _, opt := range opts {
		opt(p)
	}
	if err := p.Validate(); err != nil {
		return nil, err
	}
	return p, nil
}

// ContactOption is a functional option that configures a VCardPayload during
// construction via the Contact builder function.
type ContactOption func(*VCardPayload)

// WithPhone sets the phone number on a VCardPayload.
// The value is encoded as a TEL property in the vCard output.
func WithPhone(phone string) ContactOption {
	return func(v *VCardPayload) {
		v.Phone = phone
	}
}

// WithEmail sets the email address on a VCardPayload.
// The value is encoded as an EMAIL property in the vCard output.
func WithEmail(email string) ContactOption {
	return func(v *VCardPayload) {
		v.Email = email
	}
}

// WithOrganization sets the organization name on a VCardPayload.
// The value is encoded as an ORG property in the vCard output.
func WithOrganization(org string) ContactOption {
	return func(v *VCardPayload) {
		v.Organization = org
	}
}

// WithTitle sets the job title on a VCardPayload.
// The value is encoded as a TITLE property in the vCard output.
func WithTitle(title string) ContactOption {
	return func(v *VCardPayload) {
		v.Title = title
	}
}

// WithAddress sets the postal address on a VCardPayload.
// The value is encoded as an ADR property in the vCard output.
func WithAddress(addr string) ContactOption {
	return func(v *VCardPayload) {
		v.Address = addr
	}
}

// WithURL sets the website URL on a VCardPayload.
// The value is encoded as a URL property in the vCard output.
func WithURL(url string) ContactOption {
	return func(v *VCardPayload) {
		v.URL = url
	}
}

// WithNote sets the free-text note on a VCardPayload.
// The value is encoded as a NOTE property in the vCard output.
func WithNote(note string) ContactOption {
	return func(v *VCardPayload) {
		v.Note = note
	}
}

// Event creates a validated CalendarPayload using the iCalendar VEVENT format
// (RFC 5545). Dates are encoded in UTC. By default, the event uses date-time
// format (YYYYMMDDTHHMMSSZ); use WithAllDay() for date-only events.
//
// Example:
//
//	start := time.Date(2025, 7, 15, 9, 0, 0, 0, time.UTC)
//	end := time.Date(2025, 7, 15, 10, 0, 0, 0, time.UTC)
//	p, err := payload.Event("Standup", "Room 42", start, end,
//	    payload.WithDescription("Daily sync"),
//	)
//	data, _ := p.Encode() // "BEGIN:VEVENT\r\nSUMMARY:Standup\r\n..."
func Event(title, location string, start, end time.Time, opts ...EventOption) (*CalendarPayload, error) {
	p := &CalendarPayload{
		Title:    title,
		Location: location,
		Start:    start,
		End:      end,
	}
	for _, opt := range opts {
		opt(p)
	}
	if err := p.Validate(); err != nil {
		return nil, err
	}
	return p, nil
}

// EventOption is a functional option that configures a CalendarPayload during
// construction via the Event builder function.
type EventOption func(*CalendarPayload)

// WithAllDay marks a CalendarPayload as an all-day event, causing dates to be
// encoded in date-only format (YYYYMMDD) instead of date-time format.
func WithAllDay() EventOption {
	return func(c *CalendarPayload) {
		c.AllDay = true
	}
}

// WithDescription sets the description on a CalendarPayload.
// The value is encoded as a DESCRIPTION property in the VEVENT output.
func WithDescription(desc string) EventOption {
	return func(c *CalendarPayload) {
		c.Description = desc
	}
}

// Geo creates a validated GeoPayload encoded as a geo: URI (RFC 5870).
// Coordinates are in decimal degrees. Latitude must be in [-90, 90] and
// longitude in [-180, 180].
//
// Example:
//
//	p, err := payload.Geo(37.7749, -122.4194)
//	data, _ := p.Encode() // "geo:37.7749,-122.4194"
func Geo(lat, lng float64) (*GeoPayload, error) {
	p := &GeoPayload{Latitude: lat, Longitude: lng}
	if err := p.Validate(); err != nil {
		return nil, err
	}
	return p, nil
}

// GoogleMaps creates a validated GoogleMapsPayload from coordinates, producing
// a maps.google.com URL with the loc: query parameter.
//
// Example:
//
//	p, err := payload.GoogleMaps(37.7749, -122.4194)
//	data, _ := p.Encode() // "https://maps.google.com/maps?q=loc:37.7749,-122.4194"
func GoogleMaps(lat, lng float64) (*GoogleMapsPayload, error) {
	p := &GoogleMapsPayload{Latitude: lat, Longitude: lng}
	if err := p.Validate(); err != nil {
		return nil, err
	}
	return p, nil
}

// GoogleMapsQuery creates a validated GoogleMapsPayload from a search query,
// producing a maps.google.com URL that opens the search results.
//
// Example:
//
//	p, err := payload.GoogleMapsQuery("coffee shop near central park")
//	data, _ := p.Encode() // "https://maps.google.com/maps?q=coffee+shop+near+central+park"
func GoogleMapsQuery(query string) (*GoogleMapsPayload, error) {
	if query == "" {
		return nil, fmt.Errorf("google_maps builder: query must not be empty")
	}
	p := &GoogleMapsPayload{Query: query}
	if err := p.Validate(); err != nil {
		return nil, err
	}
	return p, nil
}

// AppleMaps creates a validated AppleMapsPayload from coordinates, producing
// a maps.apple.com URL with the ll (lat/lng) query parameter.
//
// Example:
//
//	p, err := payload.AppleMaps(37.7749, -122.4194)
//	data, _ := p.Encode() // "https://maps.apple.com/maps?ll=37.7749,-122.4194"
func AppleMaps(lat, lng float64) (*AppleMapsPayload, error) {
	p := &AppleMapsPayload{Latitude: lat, Longitude: lng}
	if err := p.Validate(); err != nil {
		return nil, err
	}
	return p, nil
}

// Twitter creates a validated TwitterPayload that encodes a twitter.com profile
// link for the given username (without the @ prefix).
//
// Example:
//
//	p, err := payload.Twitter("golang")
//	data, _ := p.Encode() // "https://twitter.com/golang"
func Twitter(username string) (*TwitterPayload, error) {
	p := &TwitterPayload{Username: username}
	if err := p.Validate(); err != nil {
		return nil, err
	}
	return p, nil
}

// LinkedIn creates a validated LinkedInPayload from a full LinkedIn profile URL.
// The URL must start with https://.
//
// Example:
//
//	p, err := payload.LinkedIn("https://www.linkedin.com/in/johndoe")
//	data, _ := p.Encode() // "https://www.linkedin.com/in/johndoe"
func LinkedIn(profileURL string) (*LinkedInPayload, error) {
	p := &LinkedInPayload{ProfileURL: profileURL}
	if err := p.Validate(); err != nil {
		return nil, err
	}
	return p, nil
}

// Telegram creates a validated TelegramPayload that encodes a t.me profile link
// for the given username (without the @ prefix).
//
// Example:
//
//	p, err := payload.Telegram("golang")
//	data, _ := p.Encode() // "https://t.me/golang"
func Telegram(username string) (*TelegramPayload, error) {
	p := &TelegramPayload{Username: username}
	if err := p.Validate(); err != nil {
		return nil, err
	}
	return p, nil
}

// Instagram creates a validated InstagramPayload that encodes an
// instagram.com profile link for the given username.
//
// Example:
//
//	p, err := payload.Instagram("natgeo")
//	data, _ := p.Encode() // "https://instagram.com/natgeo"
func Instagram(username string) (*InstagramPayload, error) {
	p := &InstagramPayload{Username: username}
	if err := p.Validate(); err != nil {
		return nil, err
	}
	return p, nil
}

// Facebook creates a validated FacebookPayload from a full Facebook page URL.
// The URL must start with https://.
//
// Example:
//
//	p, err := payload.Facebook("https://www.facebook.com/golang")
//	data, _ := p.Encode() // "https://www.facebook.com/golang"
func Facebook(pageURL string) (*FacebookPayload, error) {
	p := &FacebookPayload{PageURL: pageURL}
	if err := p.Validate(); err != nil {
		return nil, err
	}
	return p, nil
}

// SpotifyTrack creates a validated SpotifyTrackPayload that encodes an
// open.spotify.com/track/ deep link for the given track ID.
//
// Example:
//
//	p, err := payload.SpotifyTrack("4cOdK2wGLETKBW3PvgPWqT")
//	data, _ := p.Encode() // "https://open.spotify.com/track/4cOdK2wGLETKBW3PvgPWqT"
func SpotifyTrack(trackID string) (*SpotifyTrackPayload, error) {
	p := &SpotifyTrackPayload{TrackID: trackID}
	if err := p.Validate(); err != nil {
		return nil, err
	}
	return p, nil
}

// YouTubeVideo creates a validated YouTubeVideoPayload that encodes a
// youtube.com/watch?v= URL for the given video ID.
//
// Example:
//
//	p, err := payload.YouTubeVideo("dQw4w9WgXcQ")
//	data, _ := p.Encode() // "https://youtube.com/watch?v=dQw4w9WgXcQ"
func YouTubeVideo(videoID string) (*YouTubeVideoPayload, error) {
	p := &YouTubeVideoPayload{VideoID: videoID}
	if err := p.Validate(); err != nil {
		return nil, err
	}
	return p, nil
}

// AppStore creates a validated MarketPayload for the Apple App Store.
// The appID is the App Store numeric ID used in the apps.apple.com/app/ URL.
//
// Example:
//
//	p, err := payload.AppStore("1234567890")
//	data, _ := p.Encode() // "https://apps.apple.com/app/1234567890"
func AppStore(appID string) (*MarketPayload, error) {
	if appID == "" {
		return nil, fmt.Errorf("market builder: app ID must not be empty")
	}
	p := &MarketPayload{
		Platform:  MarketAppleApp,
		PackageID: appID,
	}
	if err := p.Validate(); err != nil {
		return nil, err
	}
	return p, nil
}

// PlayStore creates a validated MarketPayload for the Google Play Store.
// The packageID is the application package name (e.g. "com.example.app").
//
// Example:
//
//	p, err := payload.PlayStore("com.example.app")
//	data, _ := p.Encode() // "https://play.google.com/store/apps/details?id=com.example.app"
func PlayStore(packageID string) (*MarketPayload, error) {
	if packageID == "" {
		return nil, fmt.Errorf("market builder: package ID must not be empty")
	}
	p := &MarketPayload{
		Platform:  MarketGooglePlay,
		PackageID: packageID,
	}
	if err := p.Validate(); err != nil {
		return nil, err
	}
	return p, nil
}

// Payment creates a validated PayPalPayload that encodes a paypal.me payment
// link with the given username, amount, and three-letter currency code.
//
// Example:
//
//	p, err := payload.Payment("johndoe", "25.00", "USD")
//	data, _ := p.Encode() // "https://paypal.me/johndoe/25.00/USD"
func Payment(username, amount, currency string) (*PayPalPayload, error) {
	p := &PayPalPayload{
		Username: username,
		Amount:   amount,
		Currency: currency,
	}
	if err := p.Validate(); err != nil {
		return nil, err
	}
	return p, nil
}

// IBeacon creates a validated IBeaconPayload that encodes an iBeacon
// configuration as a beacon registry URL (https://beacon.github.io/beacon/).
// The UUID must be in the format XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX.
// Major and minor values are 16-bit unsigned integers.
//
// Example:
//
//	p, err := payload.IBeacon("A1B2C3D4-E5F6-7890-ABCD-EF1234567890", 1, 42)
//	data, _ := p.Encode() // "https://beacon.github.io/beacon/?uuid=...&major=1&minor=42"
func IBeacon(uuid string, major, minor uint16) (*IBeaconPayload, error) {
	p := &IBeaconPayload{
		UUID:  uuid,
		Major: major,
		Minor: minor,
	}
	if err := p.Validate(); err != nil {
		return nil, err
	}
	return p, nil
}

// MMS creates a validated MMSPayload encoded as an mms: URI with optional
// subject and body query parameters. The phone number must contain at least
// one digit.
//
// Example:
//
//	p, err := payload.MMS("+14155552671", "Check this out!")
//	data, _ := p.Encode() // "mms:+14155552671?body=Check this out!"
func MMS(phone, message string) (*MMSPayload, error) {
	p := &MMSPayload{
		Phone:   phone,
		Message: message,
	}
	if err := p.Validate(); err != nil {
		return nil, err
	}
	return p, nil
}

// Zoom creates a validated ZoomPayload that encodes a zoom.us/j/ meeting
// join link with an optional password (pwd) and display name (uname).
//
// Example:
//
//	p, err := payload.Zoom("1234567890", "secret")
//	data, _ := p.Encode() // "https://zoom.us/j/1234567890?pwd=secret"
func Zoom(meetingID, password string) (*ZoomPayload, error) {
	p := &ZoomPayload{
		MeetingID: meetingID,
		Password:  password,
	}
	if err := p.Validate(); err != nil {
		return nil, err
	}
	return p, nil
}
