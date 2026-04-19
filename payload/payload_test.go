package payload

import (
	"strings"
	"testing"
	"time"
)

// --- TextPayload ---

func TestTextPayload(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		wantErr bool
	}{
		{name: "valid", text: "Hello World"},
		{name: "empty", text: "", wantErr: true},
		{name: "unicode", text: "日本語テスト"},
		{name: "special chars", text: "test@example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &TextPayload{Text: tt.text}
			err := p.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if p.Type() != "text" {
					t.Errorf("Type() = %q, want %q", p.Type(), "text")
				}
				if p.Size() <= 0 {
					t.Error("Size() should be > 0")
				}
				enc, err := p.Encode()
				if err != nil {
					t.Errorf("Encode() error = %v", err)
				}
				if enc != tt.text {
					t.Errorf("Encode() = %q, want %q", enc, tt.text)
				}
			}
		})
	}
}

// --- URLPayload ---

func TestURLPayload(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{name: "https url", url: "https://example.com"},
		{name: "http url", url: "http://example.com"},
		{name: "plain domain", url: "example.com"},
		{name: "with path", url: "https://example.com/path?q=1"},
		{name: "empty", url: "", wantErr: true},
		{name: "invalid scheme", url: "ftp://example.com", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &URLPayload{URL: tt.url}
			err := p.Validate()
			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if p.Type() != "url" {
					t.Errorf("Type() = %q, want %q", p.Type(), "url")
				}
				if p.Size() <= 0 {
					t.Error("Size() should be > 0")
				}
				enc, err := p.Encode()
				if err != nil {
					t.Errorf("Encode() error = %v", err)
				}
				if !strings.HasPrefix(enc, "https://") && !strings.HasPrefix(enc, "http://") {
					t.Errorf("Encode() = %q, should have scheme", enc)
				}
			}
		})
	}
}

// --- WiFiPayload ---

func TestWiFiPayload(t *testing.T) {
	tests := []struct {
		name        string
		ssid        string
		password    string
		encryption  string
		wantErr     bool
		errContains string
	}{
		{name: "WPA2", ssid: "MyWiFi", password: "pass123", encryption: "WPA2"},
		{name: "WPA", ssid: "Home", password: "secret", encryption: "WPA"},
		{name: "WEP", ssid: "Old", password: "wepkey", encryption: "WEP"},
		{name: "nopass", ssid: "Open", password: "", encryption: "nopass"},
		{name: "WPA3", ssid: "New", password: "wpa3pass", encryption: "WPA3"},
		{name: "SAE", ssid: "Modern", password: "saepass", encryption: "SAE"},
		{name: "empty ssid", ssid: "", password: "x", encryption: "WPA2", wantErr: true, errContains: "SSID"},
		{name: "invalid encryption", ssid: "X", password: "x", encryption: "invalid", wantErr: true, errContains: "invalid encryption"},
		{name: "password required", ssid: "X", password: "", encryption: "WPA2", wantErr: true, errContains: "password"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &WiFiPayload{SSID: tt.ssid, Password: tt.password, Encryption: tt.encryption}
			err := p.Validate()
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error %q should contain %q", err.Error(), tt.errContains)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if p.Type() != "wifi" {
				t.Errorf("Type() = %q, want %q", p.Type(), "wifi")
			}
			enc, err := p.Encode()
			if err != nil {
				t.Fatalf("Encode() error: %v", err)
			}
			if !strings.HasPrefix(enc, "WIFI:T:") {
				t.Errorf("Encode() = %q, should start with WIFI:T:", enc)
			}
			if !strings.HasSuffix(enc, ";;") {
				t.Errorf("Encode() = %q, should end with ;;", enc)
			}
		})
	}
}

// --- EmailPayload ---

func TestEmailPayload(t *testing.T) {
	tests := []struct {
		name    string
		to      string
		wantErr bool
	}{
		{name: "valid", to: "user@example.com"},
		{name: "with subject/body", to: "user@example.com"},
		{name: "empty to", to: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &EmailPayload{To: tt.to, Subject: "Hello", Body: "World"}
			err := p.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if p.Type() != "email" {
					t.Errorf("Type() = %q", p.Type())
				}
				enc, _ := p.Encode()
				if !strings.HasPrefix(enc, "mailto:") {
					t.Errorf("Encode() = %q", enc)
				}
			}
		})
	}
}

// --- PhonePayload ---

func TestPhonePayload(t *testing.T) {
	tests := []struct {
		name    string
		number  string
		wantErr bool
	}{
		{name: "valid", number: "+1234567890"},
		{name: "digits only", number: "5551234"},
		{name: "empty", number: "", wantErr: true},
		{name: "no digits", number: "abc", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PhonePayload{Number: tt.number}
			err := p.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if p.Type() != "phone" {
					t.Errorf("Type() = %q", p.Type())
				}
				enc, _ := p.Encode()
				if !strings.HasPrefix(enc, "tel:") {
					t.Errorf("Encode() = %q", enc)
				}
			}
		})
	}
}

// --- SMSPayload ---

func TestSMSPayload(t *testing.T) {
	tests := []struct {
		name    string
		phone   string
		wantErr bool
	}{
		{name: "valid", phone: "+1234567890"},
		{name: "empty phone", phone: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &SMSPayload{Phone: tt.phone, Message: "Hello"}
			err := p.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if p.Type() != "sms" {
					t.Errorf("Type() = %q", p.Type())
				}
				enc, _ := p.Encode()
				if !strings.HasPrefix(enc, "smsto:") {
					t.Errorf("Encode() = %q", enc)
				}
			}
		})
	}
}

// --- MMSPayload ---

func TestMMSPayload(t *testing.T) {
	p := &MMSPayload{Phone: "+1234567890", Message: "Hi"}
	if err := p.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Type() != "mms" {
		t.Errorf("Type() = %q", p.Type())
	}
	enc, _ := p.Encode()
	if !strings.HasPrefix(enc, "mms:") {
		t.Errorf("Encode() = %q", enc)
	}

	p2 := &MMSPayload{Phone: "abc"}
	if err := p2.Validate(); err == nil {
		t.Error("expected error for no digits")
	}
}

// --- VCardPayload ---

func TestVCardPayload(t *testing.T) {
	tests := []struct {
		name      string
		firstName string
		lastName  string
		wantErr   bool
	}{
		{name: "both names", firstName: "John", lastName: "Doe"},
		{name: "first only", firstName: "John", lastName: ""},
		{name: "last only", firstName: "", lastName: "Doe"},
		{name: "neither", firstName: "", lastName: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &VCardPayload{FirstName: tt.firstName, LastName: tt.lastName, Phone: "555-1234", Email: "j@example.com"}
			err := p.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if p.Type() != "vcard" {
					t.Errorf("Type() = %q", p.Type())
				}
				enc, _ := p.Encode()
				if !strings.Contains(enc, "BEGIN:VCARD") {
					t.Errorf("Encode() missing BEGIN:VCARD")
				}
				if !strings.Contains(enc, "END:VCARD") {
					t.Errorf("Encode() missing END:VCARD")
				}
			}
		})
	}
}

// --- MeCardPayload ---

func TestMeCardPayload(t *testing.T) {
	p := &MeCardPayload{Name: "John Doe", Phone: "555-1234"}
	if err := p.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Type() != "mecard" {
		t.Errorf("Type() = %q", p.Type())
	}
	enc, _ := p.Encode()
	if !strings.HasPrefix(enc, "MECARD:") {
		t.Errorf("Encode() = %q", enc)
	}

	p2 := &MeCardPayload{}
	if err := p2.Validate(); err == nil {
		t.Error("expected error for empty name")
	}
}

// --- GeoPayload ---

func TestGeoPayload(t *testing.T) {
	tests := []struct {
		name     string
		lat, lng float64
		wantErr  bool
	}{
		{name: "valid", lat: 37.7749, lng: -122.4194},
		{name: "zero", lat: 0, lng: 0},
		{name: "south pole", lat: -90, lng: 0},
		{name: "invalid lat", lat: 91, lng: 0, wantErr: true},
		{name: "invalid lng", lat: 0, lng: 181, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &GeoPayload{Latitude: tt.lat, Longitude: tt.lng}
			err := p.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if p.Type() != "geo" {
					t.Errorf("Type() = %q", p.Type())
				}
				enc, _ := p.Encode()
				if !strings.HasPrefix(enc, "geo:") {
					t.Errorf("Encode() = %q", enc)
				}
			}
		})
	}
}

// --- GoogleMapsPayload ---

func TestGoogleMapsPayload(t *testing.T) {
	p := &GoogleMapsPayload{Latitude: 37.77, Longitude: -122.42}
	if err := p.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	enc, _ := p.Encode()
	if !strings.Contains(enc, "maps.google.com") {
		t.Errorf("Encode() = %q", enc)
	}

	p2 := &GoogleMapsPayload{Query: "San Francisco"}
	if err := p2.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- GoogleMapsPlacePayload ---

func TestGoogleMapsPlacePayload(t *testing.T) {
	p := &GoogleMapsPlacePayload{PlaceName: "Statue of Liberty"}
	if err := p.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Type() != "google_maps_place" {
		t.Errorf("Type() = %q", p.Type())
	}

	p2 := &GoogleMapsPlacePayload{}
	if err := p2.Validate(); err == nil {
		t.Error("expected error for empty place name")
	}
}

// --- GoogleMapsDirectionsPayload ---

func TestGoogleMapsDirectionsPayload(t *testing.T) {
	p := &GoogleMapsDirectionsPayload{Origin: "NYC", Destination: "Boston"}
	if err := p.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Type() != "google_maps_directions" {
		t.Errorf("Type() = %q", p.Type())
	}

	p2 := &GoogleMapsDirectionsPayload{}
	if err := p2.Validate(); err == nil {
		t.Error("expected error for empty origin")
	}
}

// --- AppleMapsPayload ---

func TestAppleMapsPayload(t *testing.T) {
	p := &AppleMapsPayload{Latitude: 37.77, Longitude: -122.42}
	if err := p.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	enc, _ := p.Encode()
	if !strings.Contains(enc, "maps.apple.com") {
		t.Errorf("Encode() = %q", enc)
	}
}

// --- CalendarPayload ---

func TestCalendarPayload(t *testing.T) {
	now := time.Now()
	p := &CalendarPayload{
		Title:    "Meeting",
		Location: "Office",
		Start:    now,
		End:      now.Add(2 * time.Hour),
	}
	if err := p.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Type() != "calendar" {
		t.Errorf("Type() = %q", p.Type())
	}
	enc, _ := p.Encode()
	if !strings.Contains(enc, "BEGIN:VEVENT") {
		t.Errorf("Encode() missing BEGIN:VEVENT")
	}

	p2 := &CalendarPayload{Title: "", End: now.Add(time.Hour)}
	if err := p2.Validate(); err == nil {
		t.Error("expected error for empty title")
	}

	p3 := &CalendarPayload{Title: "X", Start: now, End: now.Add(-time.Hour)}
	if err := p3.Validate(); err == nil {
		t.Error("expected error for end before start")
	}
}

// --- EventPayload ---

func TestEventPayload(t *testing.T) {
	p := &EventPayload{EventID: "evt-123", EventName: "Concert"}
	if err := p.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Type() != "event" {
		t.Errorf("Type() = %q", p.Type())
	}

	p2 := &EventPayload{}
	if err := p2.Validate(); err == nil {
		t.Error("expected error for empty event ID")
	}
}

// --- Social payloads ---

func TestTwitterPayload(t *testing.T) {
	p := &TwitterPayload{Username: "testuser"}
	if err := p.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Type() != "twitter" {
		t.Errorf("Type() = %q", p.Type())
	}
	p2 := &TwitterPayload{}
	if err := p2.Validate(); err == nil {
		t.Error("expected error")
	}
}

func TestTwitterFollowPayload(t *testing.T) {
	p := &TwitterFollowPayload{ScreenName: "testuser"}
	if err := p.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Type() != "twitter_follow" {
		t.Errorf("Type() = %q", p.Type())
	}
	p2 := &TwitterFollowPayload{}
	if err := p2.Validate(); err == nil {
		t.Error("expected error")
	}
}

func TestLinkedInPayload(t *testing.T) {
	p := &LinkedInPayload{ProfileURL: "https://linkedin.com/in/test"}
	if err := p.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	p2 := &LinkedInPayload{ProfileURL: "http://linkedin.com/in/test"}
	if err := p2.Validate(); err == nil {
		t.Error("expected error for http")
	}
	p3 := &LinkedInPayload{}
	if err := p3.Validate(); err == nil {
		t.Error("expected error")
	}
}

func TestInstagramPayload(t *testing.T) {
	p := &InstagramPayload{Username: "testuser"}
	if err := p.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Type() != "instagram" {
		t.Errorf("Type() = %q", p.Type())
	}
	p2 := &InstagramPayload{}
	if err := p2.Validate(); err == nil {
		t.Error("expected error")
	}
}

func TestFacebookPayload(t *testing.T) {
	p := &FacebookPayload{PageURL: "https://facebook.com/test"}
	if err := p.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	p2 := &FacebookPayload{PageURL: "http://facebook.com/test"}
	if err := p2.Validate(); err == nil {
		t.Error("expected error for http")
	}
}

func TestYouTubeChannelPayload(t *testing.T) {
	p := &YouTubeChannelPayload{ChannelID: "UC123"}
	if err := p.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Type() != "youtube_channel" {
		t.Errorf("Type() = %q", p.Type())
	}
	p2 := &YouTubeChannelPayload{}
	if err := p2.Validate(); err == nil {
		t.Error("expected error")
	}
}

func TestYouTubeVideoPayload(t *testing.T) {
	p := &YouTubeVideoPayload{VideoID: "dQw4w9WgXcQ"}
	if err := p.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Type() != "youtube_video" {
		t.Errorf("Type() = %q", p.Type())
	}
	p2 := &YouTubeVideoPayload{}
	if err := p2.Validate(); err == nil {
		t.Error("expected error")
	}
}

func TestTelegramPayload(t *testing.T) {
	p := &TelegramPayload{Username: "testuser"}
	if err := p.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Type() != "telegram" {
		t.Errorf("Type() = %q", p.Type())
	}
	p2 := &TelegramPayload{}
	if err := p2.Validate(); err == nil {
		t.Error("expected error")
	}
}

// --- Spotify payloads ---

func TestSpotifyTrackPayload(t *testing.T) {
	p := &SpotifyTrackPayload{TrackID: "abc123"}
	if err := p.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Type() != "spotify_track" {
		t.Errorf("Type() = %q", p.Type())
	}
	p2 := &SpotifyTrackPayload{}
	if err := p2.Validate(); err == nil {
		t.Error("expected error")
	}
}

func TestSpotifyPlaylistPayload(t *testing.T) {
	p := &SpotifyPlaylistPayload{PlaylistID: "pl123"}
	if err := p.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Type() != "spotify_playlist" {
		t.Errorf("Type() = %q", p.Type())
	}
	p2 := &SpotifyPlaylistPayload{}
	if err := p2.Validate(); err == nil {
		t.Error("expected error")
	}
}

func TestAppleMusicTrackPayload(t *testing.T) {
	p := &AppleMusicTrackPayload{AlbumID: "123", SongID: "456"}
	if err := p.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Type() != "apple_music" {
		t.Errorf("Type() = %q", p.Type())
	}
	p2 := &AppleMusicTrackPayload{}
	if err := p2.Validate(); err == nil {
		t.Error("expected error")
	}
}

// --- WhatsAppPayload ---

func TestWhatsAppPayload(t *testing.T) {
	p := &WhatsAppPayload{Phone: "+1234567890", Message: "Hi"}
	if err := p.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Type() != "whatsapp" {
		t.Errorf("Type() = %q", p.Type())
	}
	enc, _ := p.Encode()
	if !strings.Contains(enc, "wa.me") {
		t.Errorf("Encode() = %q", enc)
	}
	p2 := &WhatsAppPayload{}
	if err := p2.Validate(); err == nil {
		t.Error("expected error")
	}
}

// --- ZoomPayload ---

func TestZoomPayload(t *testing.T) {
	p := &ZoomPayload{MeetingID: "123-456-789", Password: "abc"}
	if err := p.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Type() != "zoom" {
		t.Errorf("Type() = %q", p.Type())
	}
	p2 := &ZoomPayload{}
	if err := p2.Validate(); err == nil {
		t.Error("expected error")
	}
}

// --- PayPalPayload ---

func TestPayPalPayload(t *testing.T) {
	p := &PayPalPayload{Username: "user", Amount: "10.00"}
	if err := p.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Type() != "paypal" {
		t.Errorf("Type() = %q", p.Type())
	}
	enc, _ := p.Encode()
	if !strings.Contains(enc, "paypal.me") {
		t.Errorf("Encode() = %q", enc)
	}
	p2 := &PayPalPayload{}
	if err := p2.Validate(); err == nil {
		t.Error("expected error")
	}
}

// --- CryptoPayload ---

func TestCryptoPayload(t *testing.T) {
	tests := []struct {
		name    string
		addr    string
		ct      string
		wantErr bool
	}{
		{name: "BTC", addr: "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa", ct: "BTC"},
		{name: "ETH", addr: "0x1234...", ct: "ETH"},
		{name: "LTC", addr: "L1234...", ct: "LTC"},
		{name: "invalid type", addr: "abc", ct: "DOGE", wantErr: true},
		{name: "empty addr", addr: "", ct: "BTC", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &CryptoPayload{Address: tt.addr, CryptoType: tt.ct}
			err := p.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if p.Type() != "crypto" {
					t.Errorf("Type() = %q", p.Type())
				}
			}
		})
	}
}

// --- PIDPayload ---

func TestPIDPayload(t *testing.T) {
	p := &PIDPayload{IBAN: "CH4431999123000889012", Amount: "100.00"}
	if err := p.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Type() != "pid" {
		t.Errorf("Type() = %q", p.Type())
	}
	p2 := &PIDPayload{}
	if err := p2.Validate(); err == nil {
		t.Error("expected error")
	}
}

// --- MarketPayload ---

func TestMarketPayload(t *testing.T) {
	p := &MarketPayload{Platform: MarketGooglePlay, PackageID: "com.example.app"}
	if err := p.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Type() != "market" {
		t.Errorf("Type() = %q", p.Type())
	}
	enc, _ := p.Encode()
	if !strings.Contains(enc, "play.google.com") {
		t.Errorf("Encode() = %q", enc)
	}

	p2 := &MarketPayload{Platform: MarketAppleApp, PackageID: "123456"}
	enc2, _ := p2.Encode()
	if !strings.Contains(enc2, "apps.apple.com") {
		t.Errorf("Encode() = %q", enc2)
	}

	p3 := &MarketPayload{}
	if err := p3.Validate(); err == nil {
		t.Error("expected error")
	}
}

// --- IBeaconPayload ---

func TestIBeaconPayload(t *testing.T) {
	p := &IBeaconPayload{UUID: "A1B2C3D4-E5F6-7890-ABCD-EF1234567890", Major: 1, Minor: 2}
	if err := p.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Type() != "ibeacon" {
		t.Errorf("Type() = %q", p.Type())
	}
	p2 := &IBeaconPayload{}
	if err := p2.Validate(); err == nil {
		t.Error("expected error")
	}
	p3 := &IBeaconPayload{UUID: "invalid"}
	if err := p3.Validate(); err == nil {
		t.Error("expected error for invalid UUID")
	}
}

// --- NTPLocalePayload ---

func TestNTPLocalePayload(t *testing.T) {
	p := &NTPLocalePayload{Host: "pool.ntp.org"}
	if err := p.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Type() != "ntp" {
		t.Errorf("Type() = %q", p.Type())
	}
	enc, _ := p.Encode()
	if !strings.HasPrefix(enc, "ntp://") {
		t.Errorf("Encode() = %q", enc)
	}
	p2 := &NTPLocalePayload{}
	if err := p2.Validate(); err == nil {
		t.Error("expected error")
	}
	p3 := &NTPLocalePayload{Host: "x", Port: "99999"}
	if err := p3.Validate(); err == nil {
		t.Error("expected error for port out of range")
	}
}

// --- BasePayload ---

func TestBasePayload(t *testing.T) {
	b := &BasePayload{}
	if b.Type() != "unknown" {
		t.Errorf("BasePayload.Type() = %q, want %q", b.Type(), "unknown")
	}
	if b.Size() != 0 {
		t.Errorf("BasePayload.Size() = %d, want 0", b.Size())
	}
}

// --- Comprehensive Encode/Size/Validate补充测试 ---

func TestTwitterPayloadEncode(t *testing.T) {
	p := &TwitterPayload{Username: "testuser"}
	enc, err := p.Encode()
	if err != nil {
		t.Fatalf("Encode() error: %v", err)
	}
	if enc != "https://twitter.com/testuser" {
		t.Errorf("Encode() = %q, want %q", enc, "https://twitter.com/testuser")
	}
	if sz := p.Size(); sz <= 0 {
		t.Errorf("Size() = %d, want > 0", sz)
	}
}

func TestTwitterFollowPayloadEncode(t *testing.T) {
	p := &TwitterFollowPayload{ScreenName: "testuser"}
	enc, err := p.Encode()
	if err != nil {
		t.Fatalf("Encode() error: %v", err)
	}
	if enc != "https://x.com/testuser" {
		t.Errorf("Encode() = %q, want %q", enc, "https://x.com/testuser")
	}
	if sz := p.Size(); sz <= 0 {
		t.Errorf("Size() = %d, want > 0", sz)
	}
}

func TestLinkedInPayloadEncode(t *testing.T) {
	p := &LinkedInPayload{ProfileURL: "https://linkedin.com/in/test"}
	enc, err := p.Encode()
	if err != nil {
		t.Fatalf("Encode() error: %v", err)
	}
	if enc != "https://linkedin.com/in/test" {
		t.Errorf("Encode() = %q", enc)
	}
	if p.Type() != "linkedin" {
		t.Errorf("Type() = %q", p.Type())
	}
	if sz := p.Size(); sz <= 0 {
		t.Errorf("Size() = %d, want > 0", sz)
	}
}

func TestInstagramPayloadEncode(t *testing.T) {
	p := &InstagramPayload{Username: "testuser"}
	enc, err := p.Encode()
	if err != nil {
		t.Fatalf("Encode() error: %v", err)
	}
	if enc != "https://instagram.com/testuser" {
		t.Errorf("Encode() = %q", enc)
	}
	if sz := p.Size(); sz <= 0 {
		t.Errorf("Size() = %d, want > 0", sz)
	}
}

func TestFacebookPayloadEncodeAndType(t *testing.T) {
	p := &FacebookPayload{PageURL: "https://facebook.com/test"}
	enc, err := p.Encode()
	if err != nil {
		t.Fatalf("Encode() error: %v", err)
	}
	if enc != "https://facebook.com/test" {
		t.Errorf("Encode() = %q", enc)
	}
	if p.Type() != "facebook" {
		t.Errorf("Type() = %q", p.Type())
	}
	if sz := p.Size(); sz <= 0 {
		t.Errorf("Size() = %d, want > 0", sz)
	}

	p2 := &FacebookPayload{}
	if err := p2.Validate(); err == nil {
		t.Error("expected error for empty PageURL")
	}
}

func TestYouTubeChannelPayloadEncode(t *testing.T) {
	p := &YouTubeChannelPayload{ChannelID: "UC123"}
	enc, err := p.Encode()
	if err != nil {
		t.Fatalf("Encode() error: %v", err)
	}
	if enc != "https://youtube.com/channel/UC123" {
		t.Errorf("Encode() = %q", enc)
	}
	if sz := p.Size(); sz <= 0 {
		t.Errorf("Size() = %d, want > 0", sz)
	}
}

func TestYouTubeVideoPayloadEncode(t *testing.T) {
	p := &YouTubeVideoPayload{VideoID: "dQw4w9WgXcQ"}
	enc, err := p.Encode()
	if err != nil {
		t.Fatalf("Encode() error: %v", err)
	}
	if enc != "https://youtube.com/watch?v=dQw4w9WgXcQ" {
		t.Errorf("Encode() = %q", enc)
	}
	if sz := p.Size(); sz <= 0 {
		t.Errorf("Size() = %d, want > 0", sz)
	}
}

func TestTelegramPayloadEncode(t *testing.T) {
	p := &TelegramPayload{Username: "testuser"}
	enc, err := p.Encode()
	if err != nil {
		t.Fatalf("Encode() error: %v", err)
	}
	if enc != "https://t.me/testuser" {
		t.Errorf("Encode() = %q", enc)
	}
	if sz := p.Size(); sz <= 0 {
		t.Errorf("Size() = %d, want > 0", sz)
	}
}

func TestSpotifyTrackPayloadEncode(t *testing.T) {
	p := &SpotifyTrackPayload{TrackID: "abc123"}
	enc, err := p.Encode()
	if err != nil {
		t.Fatalf("Encode() error: %v", err)
	}
	if !strings.HasPrefix(enc, "https://open.spotify.com/track/") {
		t.Errorf("Encode() = %q", enc)
	}
	if sz := p.Size(); sz <= 0 {
		t.Errorf("Size() = %d, want > 0", sz)
	}
}

func TestSpotifyPlaylistPayloadEncode(t *testing.T) {
	p := &SpotifyPlaylistPayload{PlaylistID: "pl123"}
	enc, err := p.Encode()
	if err != nil {
		t.Fatalf("Encode() error: %v", err)
	}
	if !strings.HasPrefix(enc, "https://open.spotify.com/playlist/") {
		t.Errorf("Encode() = %q", enc)
	}
	if sz := p.Size(); sz <= 0 {
		t.Errorf("Size() = %d, want > 0", sz)
	}
}

func TestAppleMusicTrackPayloadEncode(t *testing.T) {
	// Without storefront
	p := &AppleMusicTrackPayload{AlbumID: "123", SongID: "456"}
	enc, err := p.Encode()
	if err != nil {
		t.Fatalf("Encode() error: %v", err)
	}
	if !strings.Contains(enc, "music.apple.com/album/123") {
		t.Errorf("Encode() = %q", enc)
	}
	if sz := p.Size(); sz <= 0 {
		t.Errorf("Size() = %d, want > 0", sz)
	}

	// With storefront
	p2 := &AppleMusicTrackPayload{AlbumID: "123", SongID: "456", StoreFront: "us"}
	enc2, err := p2.Encode()
	if err != nil {
		t.Fatalf("Encode() error: %v", err)
	}
	if !strings.Contains(enc2, "music.apple.com/us/album/") {
		t.Errorf("Encode() = %q", enc2)
	}

	// Missing album ID
	p3 := &AppleMusicTrackPayload{SongID: "456"}
	if err := p3.Validate(); err == nil {
		t.Error("expected error for missing AlbumID")
	}
	// Missing song ID
	p4 := &AppleMusicTrackPayload{AlbumID: "123"}
	if err := p4.Validate(); err == nil {
		t.Error("expected error for missing SongID")
	}
}

func TestWhatsAppPayloadEncode(t *testing.T) {
	// With message
	p := &WhatsAppPayload{Phone: "+1234567890", Message: "Hi"}
	enc, err := p.Encode()
	if err != nil {
		t.Fatalf("Encode() error: %v", err)
	}
	if !strings.Contains(enc, "wa.me/1234567890") {
		t.Errorf("Encode() = %q", enc)
	}
	if !strings.Contains(enc, "text=Hi") {
		t.Errorf("Encode() missing text param: %q", enc)
	}
	if sz := p.Size(); sz <= 0 {
		t.Errorf("Size() = %d, want > 0", sz)
	}

	// No digits in phone
	p2 := &WhatsAppPayload{Phone: "abc"}
	if err := p2.Validate(); err == nil {
		t.Error("expected error for no digits")
	}
}

func TestZoomPayloadEncode(t *testing.T) {
	// With password and display name
	p := &ZoomPayload{MeetingID: "123-456-789", Password: "abc", DisplayName: "John"}
	enc, err := p.Encode()
	if err != nil {
		t.Fatalf("Encode() error: %v", err)
	}
	if !strings.HasPrefix(enc, "https://zoom.us/j/123-456-789") {
		t.Errorf("Encode() = %q", enc)
	}
	if !strings.Contains(enc, "pwd=abc") {
		t.Errorf("Encode() missing pwd: %q", enc)
	}
	if !strings.Contains(enc, "uname=John") {
		t.Errorf("Encode() missing uname: %q", enc)
	}
	if sz := p.Size(); sz <= 0 {
		t.Errorf("Size() = %d, want > 0", sz)
	}

	// MeetingID only
	p2 := &ZoomPayload{MeetingID: "999"}
	enc2, _ := p2.Encode()
	if enc2 != "https://zoom.us/j/999" {
		t.Errorf("Encode() = %q", enc2)
	}
}

func TestPayPalPayloadEncode(t *testing.T) {
	// With currency and reference
	p := &PayPalPayload{Username: "user", Amount: "10.00", Currency: "EUR", Reference: "invoice123"}
	enc, err := p.Encode()
	if err != nil {
		t.Fatalf("Encode() error: %v", err)
	}
	if !strings.Contains(enc, "paypal.me/user/10.00/EUR") {
		t.Errorf("Encode() = %q", enc)
	}
	if !strings.Contains(enc, "note=invoice123") {
		t.Errorf("Encode() missing reference: %q", enc)
	}
	if sz := p.Size(); sz <= 0 {
		t.Errorf("Size() = %d, want > 0", sz)
	}

	// Default currency
	p2 := &PayPalPayload{Username: "user", Amount: "5.00"}
	enc2, _ := p2.Encode()
	if !strings.Contains(enc2, "paypal.me/user/5.00/USD") {
		t.Errorf("Encode() default currency = %q", enc2)
	}

	// Empty username
	p3 := &PayPalPayload{Amount: "10"}
	if err := p3.Validate(); err == nil {
		t.Error("expected error for empty username")
	}
	// Empty amount
	p4 := &PayPalPayload{Username: "user"}
	if err := p4.Validate(); err == nil {
		t.Error("expected error for empty amount")
	}
}

func TestCryptoPayloadEncode(t *testing.T) {
	tests := []struct {
		name    string
		p       *CryptoPayload
		wantPre string
	}{
		{name: "BTC", p: &CryptoPayload{Address: "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa", CryptoType: CryptoBTC}, wantPre: "bitcoin:"},
		{name: "ETH", p: &CryptoPayload{Address: "0x1234", CryptoType: CryptoETH}, wantPre: "ethereum:"},
		{name: "LTC", p: &CryptoPayload{Address: "L1234", CryptoType: CryptoLTC}, wantPre: "litecoin:"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enc, err := tt.p.Encode()
			if err != nil {
				t.Fatalf("Encode() error: %v", err)
			}
			if !strings.HasPrefix(enc, tt.wantPre) {
				t.Errorf("Encode() = %q, want prefix %q", enc, tt.wantPre)
			}
			if sz := tt.p.Size(); sz <= 0 {
				t.Errorf("Size() = %d, want > 0", sz)
			}
		})
	}

	// With amount, label, message
	p := &CryptoPayload{Address: "1ABC", CryptoType: CryptoBTC, Amount: "0.5", Label: "test", Message: "hello"}
	enc, _ := p.Encode()
	if !strings.Contains(enc, "amount=0.5") {
		t.Errorf("Encode() missing amount: %q", enc)
	}
	if !strings.Contains(enc, "label=test") {
		t.Errorf("Encode() missing label: %q", enc)
	}
	if !strings.Contains(enc, "message=hello") {
		t.Errorf("Encode() missing message: %q", enc)
	}
}

func TestEventPayloadEncode(t *testing.T) {
	p := &EventPayload{
		EventID:     "evt-123",
		EventName:   "Concert",
		Venue:       "Madison Square Garden",
		Organizer:   "LiveNation",
		Description: "Rock show",
		URL:         "https://example.com/tickets",
	}
	enc, err := p.Encode()
	if err != nil {
		t.Fatalf("Encode() error: %v", err)
	}
	if !strings.HasPrefix(enc, "EVENT-TICKET:") {
		t.Errorf("Encode() = %q", enc)
	}
	if !strings.Contains(enc, "Concert") {
		t.Errorf("Encode() missing EventName: %q", enc)
	}
	if !strings.Contains(enc, "Madison") {
		t.Errorf("Encode() missing Venue: %q", enc)
	}
	if sz := p.Size(); sz <= 0 {
		t.Errorf("Size() = %d, want > 0", sz)
	}

	// With StartTime
	p2 := &EventPayload{EventID: "evt-456", StartTime: time.Now()}
	enc2, _ := p2.Encode()
	if !strings.Contains(enc2, "evt-456") {
		t.Errorf("Encode() = %q", enc2)
	}

	// With Seat, Category
	p3 := &EventPayload{EventID: "evt-789", Seat: "A12", Category: "VIP"}
	enc3, _ := p3.Encode()
	if !strings.Contains(enc3, "A12") || !strings.Contains(enc3, "VIP") {
		t.Errorf("Encode() = %q", enc3)
	}
}

func TestPIDPayloadEncode(t *testing.T) {
	// Full PID
	p := &PIDPayload{
		PIDType:        "QRR",
		CreditorName:   "ACME Corp",
		IBAN:           "CH4431999123000889012",
		Reference:      "REF-123",
		Amount:         "100.00",
		Currency:       "CHF",
		DebtorName:     "John Doe",
		RemittanceInfo: "Invoice 2024",
	}
	enc, err := p.Encode()
	if err != nil {
		t.Fatalf("Encode() error: %v", err)
	}
	if !strings.HasPrefix(enc, "PID:QRR|") {
		t.Errorf("Encode() = %q", enc)
	}
	if !strings.Contains(enc, "ACME Corp") {
		t.Errorf("Encode() missing CreditorName: %q", enc)
	}
	if sz := p.Size(); sz <= 0 {
		t.Errorf("Size() = %d, want > 0", sz)
	}

	// Default PIDType
	p2 := &PIDPayload{IBAN: "CH9300762011623852957"}
	enc2, _ := p2.Encode()
	if !strings.HasPrefix(enc2, "PID:QRR|") {
		t.Errorf("Encode() default PIDType = %q", enc2)
	}

	// SCOR type
	p3 := &PIDPayload{PIDType: "SCOR", IBAN: "CH9300762011623852957"}
	enc3, _ := p3.Encode()
	if !strings.HasPrefix(enc3, "PID:SCOR|") {
		t.Errorf("Encode() SCOR = %q", enc3)
	}

	// NON type
	p4 := &PIDPayload{PIDType: "NON", IBAN: "CH9300762011623852957"}
	enc4, _ := p4.Encode()
	if !strings.HasPrefix(enc4, "PID:NON|") {
		t.Errorf("Encode() NON = %q", enc4)
	}

	// Invalid PID type
	p5 := &PIDPayload{PIDType: "INVALID", IBAN: "CH9300762011623852957"}
	if err := p5.Validate(); err == nil {
		t.Error("expected error for invalid PID type")
	}
}

func TestMarketPayloadEncode(t *testing.T) {
	// Google Play with campaign
	p := &MarketPayload{Platform: MarketGooglePlay, PackageID: "com.example.app", Campaign: "qr_campaign"}
	enc, err := p.Encode()
	if err != nil {
		t.Fatalf("Encode() error: %v", err)
	}
	if !strings.Contains(enc, "play.google.com/store/apps/details?id=com.example.app") {
		t.Errorf("Encode() = %q", enc)
	}
	if !strings.Contains(enc, "utm_campaign=qr_campaign") {
		t.Errorf("Encode() missing campaign: %q", enc)
	}

	// Google Play with AppName only
	p2 := &MarketPayload{Platform: MarketGooglePlay, AppName: "My App"}
	enc2, _ := p2.Encode()
	if !strings.Contains(enc2, "play.google.com/store/search?q=My+App") {
		t.Errorf("Encode() = %q", enc2)
	}

	// Apple App Store
	p3 := &MarketPayload{Platform: MarketAppleApp, PackageID: "123456", Campaign: "promo"}
	enc3, _ := p3.Encode()
	if !strings.Contains(enc3, "apps.apple.com/app/123456") {
		t.Errorf("Encode() = %q", enc3)
	}
	if !strings.Contains(enc3, "utm_campaign=promo") {
		t.Errorf("Encode() missing campaign: %q", enc3)
	}

	// Apple with AppName
	p4 := &MarketPayload{Platform: MarketAppleApp, AppName: "MyApp"}
	enc4, _ := p4.Encode()
	if !strings.Contains(enc4, "apps.apple.com/search?term=MyApp") {
		t.Errorf("Encode() = %q", enc4)
	}

	// Invalid platform
	p5 := &MarketPayload{Platform: "windows", PackageID: "com.test"}
	if err := p5.Validate(); err == nil {
		t.Error("expected error for invalid platform")
	}

	// Size > 0
	if sz := p.Size(); sz <= 0 {
		t.Errorf("Size() = %d, want > 0", sz)
	}
}

func TestIBeaconPayloadEncode(t *testing.T) {
	p := &IBeaconPayload{UUID: "A1B2C3D4-E5F6-7890-ABCD-EF1234567890", Major: 1, Minor: 2}
	enc, err := p.Encode()
	if err != nil {
		t.Fatalf("Encode() error: %v", err)
	}
	if !strings.Contains(enc, "beacon.github.io/beacon/") {
		t.Errorf("Encode() = %q", enc)
	}
	if !strings.Contains(enc, "A1B2C3D4-E5F6-7890-ABCD-EF1234567890") {
		t.Errorf("Encode() missing UUID: %q", enc)
	}
	if sz := p.Size(); sz <= 0 {
		t.Errorf("Size() = %d, want > 0", sz)
	}

	// With manufacturer
	p2 := &IBeaconPayload{UUID: "A1B2C3D4-E5F6-7890-ABCD-EF1234567890", Major: 10, Minor: 20, Manufacturer: "Apple"}
	enc2, _ := p2.Encode()
	if !strings.Contains(enc2, "apple.github.io/beacon/") {
		t.Errorf("Encode() = %q", enc2)
	}
	if !strings.Contains(enc2, "major=10") || !strings.Contains(enc2, "minor=20") {
		t.Errorf("Encode() missing major/minor: %q", enc2)
	}

	// UUID too short
	p3 := &IBeaconPayload{UUID: "too-short"}
	if err := p3.Validate(); err == nil {
		t.Error("expected error for short UUID")
	}

	// UUID with invalid chars
	p4 := &IBeaconPayload{UUID: "GGGGGGGG-GGGG-GGGG-GGGG-GGGGGGGGGGGG"}
	if err := p4.Validate(); err == nil {
		t.Error("expected error for invalid UUID chars")
	}
}

func TestNTPLocalePayloadEncode(t *testing.T) {
	// Default port
	p := &NTPLocalePayload{Host: "pool.ntp.org"}
	enc, err := p.Encode()
	if err != nil {
		t.Fatalf("Encode() error: %v", err)
	}
	if enc != "ntp://pool.ntp.org" {
		t.Errorf("Encode() = %q", enc)
	}
	if sz := p.Size(); sz <= 0 {
		t.Errorf("Size() = %d, want > 0", sz)
	}

	// Custom port (not 123)
	p2 := &NTPLocalePayload{Host: "time.example.com", Port: "4567"}
	enc2, _ := p2.Encode()
	if !strings.Contains(enc2, "ntp://time.example.com:4567") {
		t.Errorf("Encode() = %q", enc2)
	}

	// Port 123 should be omitted
	p3 := &NTPLocalePayload{Host: "x.ntp.org", Port: "123"}
	enc3, _ := p3.Encode()
	if enc3 != "ntp://x.ntp.org" {
		t.Errorf("Encode() port 123 = %q", enc3)
	}

	// With description
	p4 := &NTPLocalePayload{Host: "pool.ntp.org", Description: "US pool"}
	enc4, _ := p4.Encode()
	if !strings.Contains(enc4, "#US+pool") {
		t.Errorf("Encode() = %q", enc4)
	}

	// With version 3
	p5 := &NTPLocalePayload{Host: "ntp.local", Version: 3}
	if err := p5.Validate(); err != nil {
		t.Fatalf("Validate() error: %v", err)
	}

	// Invalid version
	p6 := &NTPLocalePayload{Host: "ntp.local", Version: 2}
	if err := p6.Validate(); err == nil {
		t.Error("expected error for invalid version 2")
	}

	// Non-numeric port
	p7 := &NTPLocalePayload{Host: "ntp.local", Port: "abc"}
	if err := p7.Validate(); err == nil {
		t.Error("expected error for non-numeric port")
	}

	// Port 0 (out of range)
	p8 := &NTPLocalePayload{Host: "ntp.local", Port: "0"}
	if err := p8.Validate(); err == nil {
		t.Error("expected error for port 0")
	}

	// String() method
	s := p.String()
	if !strings.Contains(s, "NTP://pool.ntp.org") {
		t.Errorf("String() = %q", s)
	}

	p9 := &NTPLocalePayload{Host: "t.local", Version: 4}
	s2 := p9.String()
	if !strings.Contains(s2, "(v4)") {
		t.Errorf("String() = %q", s2)
	}
}

func TestMeCardPayloadEncode(t *testing.T) {
	// Full MeCard
	p := &MeCardPayload{
		Name:     "John Doe",
		Phone:    "555-1234",
		Email:    "john@example.com",
		URL:      "https://example.com",
		Birthday: "19900101",
		Note:     "Test note",
		Address:  "123 Main St",
		Nickname: "Johnny",
	}
	enc, err := p.Encode()
	if err != nil {
		t.Fatalf("Encode() error: %v", err)
	}
	if !strings.HasPrefix(enc, "MECARD:N:John Doe;") {
		t.Errorf("Encode() = %q", enc)
	}
	if !strings.Contains(enc, "TEL:555-1234") {
		t.Errorf("Encode() missing TEL: %q", enc)
	}
	if !strings.Contains(enc, "EMAIL:john@example.com") {
		t.Errorf("Encode() missing EMAIL: %q", enc)
	}
	if !strings.Contains(enc, "BDAY:19900101") {
		t.Errorf("Encode() missing BDAY: %q", enc)
	}
	if !strings.Contains(enc, "NOTE:Test note") {
		t.Errorf("Encode() missing NOTE: %q", enc)
	}
	if !strings.Contains(enc, "ADR:123 Main St") {
		t.Errorf("Encode() missing ADR: %q", enc)
	}
	if !strings.Contains(enc, "NICKNAME:Johnny") {
		t.Errorf("Encode() missing NICKNAME: %q", enc)
	}
	if sz := p.Size(); sz <= 0 {
		t.Errorf("Size() = %d, want > 0", sz)
	}

	// Escape special chars
	p2 := &MeCardPayload{Name: "Doe;John"}
	enc2, _ := p2.Encode()
	if !strings.Contains(enc2, "Doe\\;John") {
		t.Errorf("Encode() should escape semicolons: %q", enc2)
	}

	p3 := &MeCardPayload{Name: "Test:Colon"}
	enc3, _ := p3.Encode()
	if !strings.Contains(enc3, "Test\\:Colon") {
		t.Errorf("Encode() should escape colons: %q", enc3)
	}

	p4 := &MeCardPayload{Name: "Back\\Slash"}
	enc4, _ := p4.Encode()
	if !strings.Contains(enc4, "Back\\\\Slash") {
		t.Errorf("Encode() should escape backslashes: %q", enc4)
	}
}

func TestGoogleMapsPayloadEncode(t *testing.T) {
	// Coordinates
	p := &GoogleMapsPayload{Latitude: 37.77, Longitude: -122.42, Zoom: 15}
	enc, err := p.Encode()
	if err != nil {
		t.Fatalf("Encode() error: %v", err)
	}
	if !strings.Contains(enc, "maps.google.com/maps?q=loc:37.77,-122.42") {
		t.Errorf("Encode() = %q", enc)
	}
	if !strings.Contains(enc, "zoom=15") {
		t.Errorf("Encode() missing zoom: %q", enc)
	}
	if p.Type() != "google_maps" {
		t.Errorf("Type() = %q", p.Type())
	}
	if sz := p.Size(); sz <= 0 {
		t.Errorf("Size() = %d, want > 0", sz)
	}

	// Query
	p2 := &GoogleMapsPayload{Query: "Golden Gate Bridge"}
	enc2, _ := p2.Encode()
	if !strings.Contains(enc2, "q=Golden+Gate+Bridge") {
		t.Errorf("Encode() = %q", enc2)
	}

	// Invalid coords
	p3 := &GoogleMapsPayload{Latitude: 91}
	if err := p3.Validate(); err == nil {
		t.Error("expected error for lat 91")
	}
	p4 := &GoogleMapsPayload{Longitude: 181}
	if err := p4.Validate(); err == nil {
		t.Error("expected error for lng 181")
	}
}

func TestGoogleMapsPlacePayloadEncode(t *testing.T) {
	p := &GoogleMapsPlacePayload{PlaceName: "Statue of Liberty"}
	enc, err := p.Encode()
	if err != nil {
		t.Fatalf("Encode() error: %v", err)
	}
	if !strings.Contains(enc, "maps.google.com/maps?q=Statue+of+Liberty") {
		t.Errorf("Encode() = %q", enc)
	}
	if sz := p.Size(); sz <= 0 {
		t.Errorf("Size() = %d, want > 0", sz)
	}
}

func TestGoogleMapsDirectionsPayloadEncode(t *testing.T) {
	p := &GoogleMapsDirectionsPayload{Origin: "NYC", Destination: "Boston", TravelMode: "walking"}
	enc, err := p.Encode()
	if err != nil {
		t.Fatalf("Encode() error: %v", err)
	}
	if !strings.Contains(enc, "origin=NYC") {
		t.Errorf("Encode() = %q", enc)
	}
	if !strings.Contains(enc, "destination=Boston") {
		t.Errorf("Encode() = %q", enc)
	}
	if !strings.Contains(enc, "travelmode=walking") {
		t.Errorf("Encode() = %q", enc)
	}
	if sz := p.Size(); sz <= 0 {
		t.Errorf("Size() = %d, want > 0", sz)
	}

	// Default travel mode
	p2 := &GoogleMapsDirectionsPayload{Origin: "A", Destination: "B"}
	enc2, _ := p2.Encode()
	if !strings.Contains(enc2, "travelmode=driving") {
		t.Errorf("Encode() default mode = %q", enc2)
	}

	// Empty destination
	p3 := &GoogleMapsDirectionsPayload{Origin: "A"}
	if err := p3.Validate(); err == nil {
		t.Error("expected error for empty destination")
	}
}

func TestAppleMapsPayloadEncode(t *testing.T) {
	p := &AppleMapsPayload{Latitude: 37.77, Longitude: -122.42, Query: "Coffee", Zoom: 10}
	enc, err := p.Encode()
	if err != nil {
		t.Fatalf("Encode() error: %v", err)
	}
	if !strings.Contains(enc, "maps.apple.com/maps?ll=37.77,-122.42") {
		t.Errorf("Encode() = %q", enc)
	}
	if !strings.Contains(enc, "q=Coffee") {
		t.Errorf("Encode() missing query: %q", enc)
	}
	if !strings.Contains(enc, "t=10") {
		t.Errorf("Encode() missing zoom: %q", enc)
	}
	if p.Type() != "apple_maps" {
		t.Errorf("Type() = %q", p.Type())
	}
	if sz := p.Size(); sz <= 0 {
		t.Errorf("Size() = %d, want > 0", sz)
	}

	// Invalid coords
	p2 := &AppleMapsPayload{Latitude: -91}
	if err := p2.Validate(); err == nil {
		t.Error("expected error for invalid lat")
	}
}

func TestEmailPayloadEncode(t *testing.T) {
	// With CC
	p := &EmailPayload{To: "user@example.com", Subject: "Hello", Body: "World", CC: []string{"cc1@example.com", "cc2@example.com"}}
	enc, err := p.Encode()
	if err != nil {
		t.Fatalf("Encode() error: %v", err)
	}
	if !strings.HasPrefix(enc, "mailto:user@example.com") {
		t.Errorf("Encode() = %q", enc)
	}
	if !strings.Contains(enc, "subject=Hello") {
		t.Errorf("Encode() missing subject: %q", enc)
	}
	if !strings.Contains(enc, "body=World") {
		t.Errorf("Encode() missing body: %q", enc)
	}
	if !strings.Contains(enc, "cc=cc1%40example.com") && !strings.Contains(enc, "cc=cc1@example.com") {
		t.Errorf("Encode() missing cc1: %q", enc)
	}
	if !strings.Contains(enc, "cc=cc2%40example.com") && !strings.Contains(enc, "cc=cc2@example.com") {
		t.Errorf("Encode() missing cc2: %q", enc)
	}
	if sz := p.Size(); sz <= 0 {
		t.Errorf("Size() = %d, want > 0", sz)
	}

	// Empty CC entries are skipped
	p2 := &EmailPayload{To: "u@e.com", CC: []string{"", "valid@example.com"}}
	enc2, _ := p2.Encode()
	if strings.Count(enc2, "cc=") != 1 {
		t.Errorf("Encode() should skip empty CC: %q", enc2)
	}
}

func TestSMSPayloadEncode(t *testing.T) {
	// With message
	p := &SMSPayload{Phone: "+1234567890", Message: "Hello"}
	enc, err := p.Encode()
	if err != nil {
		t.Fatalf("Encode() error: %v", err)
	}
	if enc != "smsto:+1234567890:Hello" {
		t.Errorf("Encode() = %q", enc)
	}
	if sz := p.Size(); sz <= 0 {
		t.Errorf("Size() = %d, want > 0", sz)
	}

	// Without message
	p2 := &SMSPayload{Phone: "5551234"}
	enc2, _ := p2.Encode()
	if enc2 != "smsto:5551234" {
		t.Errorf("Encode() = %q", enc2)
	}
}

func TestMMSPayloadEncode(t *testing.T) {
	// With subject and body
	p := &MMSPayload{Phone: "+1234567890", Subject: "Pic", Message: "See this"}
	enc, err := p.Encode()
	if err != nil {
		t.Fatalf("Encode() error: %v", err)
	}
	if !strings.HasPrefix(enc, "mms:+1234567890") {
		t.Errorf("Encode() = %q", enc)
	}
	if !strings.Contains(enc, "subject=Pic") {
		t.Errorf("Encode() missing subject: %q", enc)
	}
	if !strings.Contains(enc, "body=See this") {
		t.Errorf("Encode() missing body: %q", enc)
	}
	if sz := p.Size(); sz <= 0 {
		t.Errorf("Size() = %d, want > 0", sz)
	}

	// Empty phone
	p2 := &MMSPayload{}
	if err := p2.Validate(); err == nil {
		t.Error("expected error for empty phone")
	}
}

func TestURLPayloadTitle(t *testing.T) {
	p := &URLPayload{URL: "https://example.com", Title: "My Page"}
	enc, err := p.Encode()
	if err != nil {
		t.Fatalf("Encode() error: %v", err)
	}
	if !strings.Contains(enc, "#My+Page") {
		t.Errorf("Encode() = %q", enc)
	}
}

func TestURLPayloadProtocolRelative(t *testing.T) {
	p := &URLPayload{URL: "//example.com"}
	enc, err := p.Encode()
	if err != nil {
		t.Fatalf("Encode() error: %v", err)
	}
	if !strings.HasPrefix(enc, "https://") {
		t.Errorf("Encode() = %q", enc)
	}
}

func TestWiFiPayloadHidden(t *testing.T) {
	p := &WiFiPayload{SSID: "HiddenNet", Password: "pass", Encryption: "WPA2", Hidden: true}
	enc, err := p.Encode()
	if err != nil {
		t.Fatalf("Encode() error: %v", err)
	}
	if !strings.Contains(enc, "H:true") {
		t.Errorf("Encode() missing H:true: %q", enc)
	}
	if sz := p.Size(); sz <= 0 {
		t.Errorf("Size() = %d, want > 0", sz)
	}

	// Escape special WiFi chars
	p2 := &WiFiPayload{SSID: "My;SSID", Password: "p:ss", Encryption: "WPA2"}
	enc2, _ := p2.Encode()
	if !strings.Contains(enc2, "3B") {
		t.Errorf("Encode() should escape ; in SSID: %q", enc2)
	}
}

func TestVCardPayloadEncode(t *testing.T) {
	// Full vCard
	p := &VCardPayload{
		FirstName:    "John",
		LastName:     "Doe",
		Phone:        "555-1234",
		Email:        "john@example.com",
		Organization: "Acme",
		Title:        "Engineer",
		URL:          "https://example.com",
		Address:      "123 Main St",
		Note:         "Note",
	}
	enc, err := p.Encode()
	if err != nil {
		t.Fatalf("Encode() error: %v", err)
	}
	if !strings.Contains(enc, "VERSION:3.0") {
		t.Errorf("Encode() missing version: %q", enc)
	}
	if !strings.Contains(enc, "FN:John Doe") {
		t.Errorf("Encode() missing FN: %q", enc)
	}
	if !strings.Contains(enc, "ORG:Acme") {
		t.Errorf("Encode() missing ORG: %q", enc)
	}
	if !strings.Contains(enc, "TITLE:Engineer") {
		t.Errorf("Encode() missing TITLE: %q", enc)
	}
	if !strings.Contains(enc, "URL:https://example.com") {
		t.Errorf("Encode() missing URL: %q", enc)
	}
	if !strings.Contains(enc, "ADR:;;123 Main St;;;;") {
		t.Errorf("Encode() missing ADR: %q", enc)
	}
	if !strings.Contains(enc, "NOTE:Note") {
		t.Errorf("Encode() missing NOTE: %q", enc)
	}
	if sz := p.Size(); sz <= 0 {
		t.Errorf("Size() = %d, want > 0", sz)
	}

	// Custom version
	p2 := &VCardPayload{FirstName: "Jane", Version: "4.0"}
	enc2, _ := p2.Encode()
	if !strings.Contains(enc2, "VERSION:4.0") {
		t.Errorf("Encode() = %q", enc2)
	}

	// Invalid version
	p3 := &VCardPayload{FirstName: "Test", Version: "1.0"}
	if err := p3.Validate(); err == nil {
		t.Error("expected error for invalid vCard version")
	}
}

func TestCalendarPayloadEncode(t *testing.T) {
	// All-day event
	p := &CalendarPayload{
		Title:    "Birthday",
		Location: "Home",
		Start:    time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		End:      time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC),
		AllDay:   true,
	}
	enc, err := p.Encode()
	if err != nil {
		t.Fatalf("Encode() error: %v", err)
	}
	if !strings.Contains(enc, "BEGIN:VEVENT") {
		t.Errorf("Encode() missing BEGIN:VEVENT")
	}
	if !strings.Contains(enc, "SUMMARY:Birthday") {
		t.Errorf("Encode() missing SUMMARY: %q", enc)
	}
	if !strings.Contains(enc, "LOCATION:Home") {
		t.Errorf("Encode() missing LOCATION: %q", enc)
	}
	if !strings.Contains(enc, "20240115") {
		t.Errorf("Encode() missing date for all-day: %q", enc)
	}
	if sz := p.Size(); sz <= 0 {
		t.Errorf("Size() = %d, want > 0", sz)
	}

	// With description
	p2 := &CalendarPayload{
		Title:       "Meeting",
		Description: "Team sync",
		Start:       time.Now(),
		End:         time.Now().Add(time.Hour),
	}
	enc2, _ := p2.Encode()
	if !strings.Contains(enc2, "DESCRIPTION:Team sync") {
		t.Errorf("Encode() missing DESCRIPTION: %q", enc2)
	}
}

func TestGeoPayloadEncode(t *testing.T) {
	p := &GeoPayload{Latitude: 37.7749, Longitude: -122.4194}
	enc, err := p.Encode()
	if err != nil {
		t.Fatalf("Encode() error: %v", err)
	}
	if !strings.HasPrefix(enc, "geo:37.7749,-122.4194") {
		t.Errorf("Encode() = %q", enc)
	}
	if sz := p.Size(); sz <= 0 {
		t.Errorf("Size() = %d, want > 0", sz)
	}
}

func TestPhonePayloadEncode(t *testing.T) {
	p := &PhonePayload{Number: "+1234567890"}
	enc, err := p.Encode()
	if err != nil {
		t.Fatalf("Encode() error: %v", err)
	}
	if enc != "tel:+1234567890" {
		t.Errorf("Encode() = %q", enc)
	}
	if sz := p.Size(); sz <= 0 {
		t.Errorf("Size() = %d, want > 0", sz)
	}
}

func TestTextPayloadSize(t *testing.T) {
	p := &TextPayload{Text: "Hello"}
	if sz := p.Size(); sz != 5 {
		t.Errorf("Size() = %d, want 5", sz)
	}
}

func TestCleanPhoneNumber(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"+1-234-567-8900", "12345678900"},
		{"(+1) 234 567 8900", "12345678900"},
		{"abc123def", "123"},
		{"", ""},
	}
	for _, tt := range tests {
		got := cleanPhoneNumber(tt.input)
		if got != tt.want {
			t.Errorf("cleanPhoneNumber(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestContainsDigit(t *testing.T) {
	if !containsDigit("abc1def") {
		t.Error("containsDigit should return true for 'abc1def'")
	}
	if containsDigit("abcdef") {
		t.Error("containsDigit should return false for 'abcdef'")
	}
	if containsDigit("") {
		t.Error("containsDigit should return false for empty string")
	}
}

func TestFormatCoord(t *testing.T) {
	tests := []struct {
		input float64
		want  string
	}{
		{37.5, "37.5"},
		{-122.0, "-122"},
		{0.0, "0"},
		{1.234567, "1.234567"},
	}
	for _, tt := range tests {
		got := formatCoord(tt.input)
		if got != tt.want {
			t.Errorf("formatCoord(%v) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"https://example.com", "https://example.com"},
		{"http://example.com", "http://example.com"},
		{"//example.com", "https:////example.com"},
		{"example.com", "https://example.com"},
	}
	for _, tt := range tests {
		got := normalizeURL(tt.input)
		if got != tt.want {
			t.Errorf("normalizeURL(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestIsValidUUID(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"A1B2C3D4-E5F6-7890-ABCD-EF1234567890", true},
		{"a1b2c3d4-e5f6-7890-abcd-ef1234567890", true},
		{"A1B2C3D4E5F67890ABCDEF1234567890", true}, // no dashes, 32 hex chars
		{"short", false},
		{"GGGGGGGG-GGGG-GGGG-GGGG-GGGGGGGGGGGG", false},
	}
	for _, tt := range tests {
		got := isValidUUID(tt.input)
		if got != tt.want {
			t.Errorf("isValidUUID(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestEscapeWiFi(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello", "hello"},
		{"a;b", "a\\3Bb"},
		{"a:b", "a\\3Ab"},
		{`a\b`, "a\\5Cb"},
		{"a,b", "a\\2Cb"},
		{`a"b`, "a\\22b"},
	}
	for _, tt := range tests {
		got := escapeWiFi(tt.input)
		if got != tt.want {
			t.Errorf("escapeWiFi(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestEscapeMeCard(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello", "hello"},
		{"a;b", "a\\;b"},
		{"a:b", "a\\:b"},
		{`a\b`, "a\\\\b"},
	}
	for _, tt := range tests {
		got := escapeMeCard(tt.input)
		if got != tt.want {
			t.Errorf("escapeMeCard(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
