package payload

import (
	"testing"
	"time"
)

func TestBuilderText(t *testing.T) {
	p, err := Text("Hello World")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Type() != "text" {
		t.Errorf("Type() = %q", p.Type())
	}
	_, err = Text("")
	if err == nil {
		t.Error("expected error for empty text")
	}
}

func TestBuilderURL(t *testing.T) {
	p, err := URL("https://example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Type() != "url" {
		t.Errorf("Type() = %q", p.Type())
	}
}

func TestBuilderEmail(t *testing.T) {
	p, err := Email("user@example.com", "Hello", "World")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Type() != "email" {
		t.Errorf("Type() = %q", p.Type())
	}
	_, err = Email("", "Hi", "Body")
	if err == nil {
		t.Error("expected error")
	}
}

func TestBuilderSMS(t *testing.T) {
	p, err := SMS("+1234567890", "Hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Type() != "sms" {
		t.Errorf("Type() = %q", p.Type())
	}
}

func TestBuilderPhone(t *testing.T) {
	p, err := Phone("+1234567890")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Type() != "phone" {
		t.Errorf("Type() = %q", p.Type())
	}
}

func TestBuilderWiFi(t *testing.T) {
	p, err := WiFi("MyNetwork", "password123", "WPA2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Type() != "wifi" {
		t.Errorf("Type() = %q", p.Type())
	}
}

func TestBuilderWiFiWithHidden(t *testing.T) {
	p, err := WiFiWithHidden("HiddenNet", "pass", "WPA2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !p.Hidden {
		t.Error("Hidden should be true")
	}
}

func TestBuilderContact(t *testing.T) {
	p, err := Contact("John", "Doe",
		WithPhone("555-1234"),
		WithEmail("john@example.com"),
		WithOrganization("Acme"),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Type() != "vcard" {
		t.Errorf("Type() = %q", p.Type())
	}
	if p.Phone != "555-1234" {
		t.Errorf("Phone = %q", p.Phone)
	}
	if p.Email != "john@example.com" {
		t.Errorf("Email = %q", p.Email)
	}
}

func TestBuilderEvent(t *testing.T) {
	start := time.Now()
	end := start.Add(2 * time.Hour)
	p, err := Event("Meeting", "Office", start, end,
		WithAllDay(),
		WithDescription("Team sync"),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Type() != "calendar" {
		t.Errorf("Type() = %q", p.Type())
	}
	if !p.AllDay {
		t.Error("AllDay should be true")
	}
}

func TestBuilderGeo(t *testing.T) {
	p, err := Geo(37.7749, -122.4194)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Type() != "geo" {
		t.Errorf("Type() = %q", p.Type())
	}
}

func TestBuilderGoogleMaps(t *testing.T) {
	p, err := GoogleMaps(37.77, -122.42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Type() != "google_maps" {
		t.Errorf("Type() = %q", p.Type())
	}
}

func TestBuilderGoogleMapsQuery(t *testing.T) {
	p, err := GoogleMapsQuery("Golden Gate Bridge")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Type() != "google_maps" {
		t.Errorf("Type() = %q", p.Type())
	}
	_, err = GoogleMapsQuery("")
	if err == nil {
		t.Error("expected error for empty query")
	}
}

func TestBuilderAppleMaps(t *testing.T) {
	p, err := AppleMaps(37.77, -122.42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Type() != "apple_maps" {
		t.Errorf("Type() = %q", p.Type())
	}
}

func TestBuilderSocial(t *testing.T) { //nolint:errcheck // test helpers use _ for known-valid inputs
	tests := []struct {
		name string
		fn   func() Payload
	}{
		{name: "twitter", fn: func() Payload { p, _ := Twitter("user"); return p }},
		{name: "linkedin", fn: func() Payload { p, _ := LinkedIn("https://linkedin.com/in/user"); return p }},
		{name: "instagram", fn: func() Payload { p, _ := Instagram("user"); return p }},
		{name: "facebook", fn: func() Payload { p, _ := Facebook("https://facebook.com/page"); return p }},
		{name: "telegram", fn: func() Payload { p, _ := Telegram("user"); return p }},
		{name: "spotify_track", fn: func() Payload { p, _ := SpotifyTrack("abc"); return p }},
		{name: "youtube_video", fn: func() Payload { p, _ := YouTubeVideo("dQw4w9WgXcQ"); return p }},
		{name: "whatsapp", fn: func() Payload { p, _ := WhatsApp("+1234567890", "hi"); return p }},
		{name: "zoom", fn: func() Payload { p, _ := Zoom("123-456-789", "pass"); return p }},
		{name: "payment", fn: func() Payload { p, _ := Payment("user", "10.00", "USD"); return p }},
		{name: "mms", fn: func() Payload { p, _ := MMS("+1234567890", "hi"); return p }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := tt.fn()
			if err := p.Validate(); err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestBuilderAppStore(t *testing.T) {
	p, err := AppStore("1234567890")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Type() != "market" {
		t.Errorf("Type() = %q", p.Type())
	}
	_, err = AppStore("")
	if err == nil {
		t.Error("expected error for empty app ID")
	}
}

func TestBuilderPlayStore(t *testing.T) {
	p, err := PlayStore("com.example.app")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Type() != "market" {
		t.Errorf("Type() = %q", p.Type())
	}
	_, err = PlayStore("")
	if err == nil {
		t.Error("expected error for empty package ID")
	}
}

func TestBuilderIBeacon(t *testing.T) {
	p, err := IBeacon("A1B2C3D4-E5F6-7890-ABCD-EF1234567890", 1, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Type() != "ibeacon" {
		t.Errorf("Type() = %q", p.Type())
	}
}
