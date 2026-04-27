package payload

import (
	"strings"
	"testing"
)

func TestBasePayload(t *testing.T) {
	bp := &BasePayload{}
	if bp.Type() != "unknown" {
		t.Errorf("expected 'unknown', got %q", bp.Type())
	}
	if bp.Size() != 0 {
		t.Errorf("expected 0, got %d", bp.Size())
	}
}

func TestTextPayload(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		p := &TextPayload{Text: "Hello World"}
		enc, err := p.Encode()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if enc != "Hello World" {
			t.Errorf("expected 'Hello World', got %q", enc)
		}
		if p.Type() != "text" {
			t.Errorf("expected 'text', got %q", p.Type())
		}
		if p.Size() != 11 {
			t.Errorf("expected size 11, got %d", p.Size())
		}
	})

	t.Run("empty text", func(t *testing.T) {
		p := &TextPayload{Text: ""}
		if err := p.Validate(); err == nil {
			t.Error("expected error for empty text")
		}
	})

	t.Run("too long", func(t *testing.T) {
		p := &TextPayload{Text: string(make([]byte, maxTextLength+1))}
		if err := p.Validate(); err == nil {
			t.Error("expected error for too long text")
		}
	})

	t.Run("max length", func(t *testing.T) {
		p := &TextPayload{Text: string(make([]byte, maxTextLength))}
		if err := p.Validate(); err != nil {
			t.Errorf("max length should be valid: %v", err)
		}
	})
}

func TestURLPayload(t *testing.T) {
	t.Run("https", func(t *testing.T) {
		p := &URLPayload{URL: "https://example.com"}
		enc, err := p.Encode()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if enc != "https://example.com" {
			t.Errorf("expected 'https://example.com', got %q", enc)
		}
		if p.Type() != "url" {
			t.Errorf("expected 'url', got %q", p.Type())
		}
	})

	t.Run("normalize bare domain", func(t *testing.T) {
		p := &URLPayload{URL: "example.com"}
		enc, _ := p.Encode()
		if enc != "https://example.com" {
			t.Errorf("expected 'https://example.com', got %q", enc)
		}
	})

	t.Run("with title", func(t *testing.T) {
		p := &URLPayload{URL: "https://example.com", Title: "My Site"}
		enc, _ := p.Encode()
		if enc != "https://example.com#My+Site" {
			t.Errorf("got %q", enc)
		}
	})

	t.Run("empty URL", func(t *testing.T) {
		p := &URLPayload{URL: ""}
		if err := p.Validate(); err == nil {
			t.Error("expected error for empty URL")
		}
	})

	t.Run("unsupported scheme", func(t *testing.T) {
		p := &URLPayload{URL: "ftp://example.com"}
		if err := p.Validate(); err == nil {
			t.Error("expected error for unsupported scheme")
		}
	})
}

func TestWiFiPayload(t *testing.T) {
	t.Run("WPA2", func(t *testing.T) {
		p := &WiFiPayload{SSID: "MyWiFi", Password: "pass123", Encryption: EncryptionWPA2}
		enc, err := p.Encode()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p.Type() != "wifi" {
			t.Errorf("expected 'wifi', got %q", p.Type())
		}
		if enc != "WIFI:T:WPA2;S:MyWiFi;P:pass123;;" {
			t.Errorf("got %q", enc)
		}
	})

	t.Run("nopass", func(t *testing.T) {
		p := &WiFiPayload{SSID: "OpenNet", Encryption: EncryptionNoPass}
		enc, _ := p.Encode()
		if enc != "WIFI:T:nopass;S:OpenNet;;" {
			t.Errorf("got %q", enc)
		}
	})

	t.Run("hidden", func(t *testing.T) {
		p := &WiFiPayload{SSID: "Hidden", Password: "pw", Encryption: EncryptionWPA, Hidden: true}
		enc, _ := p.Encode()
		if enc != "WIFI:T:WPA;S:Hidden;P:pw;H:true;;" {
			t.Errorf("got %q", enc)
		}
	})

	t.Run("escape special chars", func(t *testing.T) {
		p := &WiFiPayload{SSID: "My;WiFi", Password: "p" + string(byte(92)) + "w", Encryption: EncryptionWPA2}
		enc, _ := p.Encode()
		// Semicolons and backslashes should be escaped — output should differ from raw input.
		if strings.Contains(enc, "My;WiFi") {
			t.Error("semicolon in SSID should be escaped")
		}
	})

	t.Run("empty SSID", func(t *testing.T) {
		p := &WiFiPayload{SSID: "", Encryption: EncryptionWPA2}
		if err := p.Validate(); err == nil {
			t.Error("expected error for empty SSID")
		}
	})

	t.Run("invalid encryption", func(t *testing.T) {
		p := &WiFiPayload{SSID: "Test", Encryption: "INVALID"}
		if err := p.Validate(); err == nil {
			t.Error("expected error for invalid encryption")
		}
	})

	t.Run("password required", func(t *testing.T) {
		p := &WiFiPayload{SSID: "Test", Password: "", Encryption: EncryptionWPA2}
		if err := p.Validate(); err == nil {
			t.Error("expected error for missing password")
		}
	})
}

func TestSMSPayload(t *testing.T) {
	t.Run("with message", func(t *testing.T) {
		p := &SMSPayload{Phone: "+1234567890", Message: "Hi there"}
		enc, err := p.Encode()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if enc != "smsto:+1234567890:Hi there" {
			t.Errorf("got %q", enc)
		}
		if p.Type() != "sms" {
			t.Errorf("expected 'sms', got %q", p.Type())
		}
	})

	t.Run("no message", func(t *testing.T) {
		p := &SMSPayload{Phone: "+1234567890"}
		enc, _ := p.Encode()
		if enc != "smsto:+1234567890" {
			t.Errorf("got %q", enc)
		}
	})

	t.Run("empty phone", func(t *testing.T) {
		p := &SMSPayload{Phone: ""}
		if err := p.Validate(); err == nil {
			t.Error("expected error for empty phone")
		}
	})
}

func TestEmailPayload(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		p := &EmailPayload{To: "test@example.com", Subject: "Hello", Body: "World"}
		enc, err := p.Encode()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p.Type() != "email" {
			t.Errorf("expected 'email', got %q", p.Type())
		}
		// Subject and body should be query-escaped.
		if enc != "mailto:test@example.com?subject=Hello&body=World" {
			t.Errorf("got %q", enc)
		}
	})

	t.Run("with CC", func(t *testing.T) {
		p := &EmailPayload{To: "a@b.com", CC: []string{"c@d.com", "e@f.com"}}
		enc, _ := p.Encode()
		if enc != "mailto:a@b.com?cc=c%40d.com&cc=e%40f.com" {
			t.Errorf("got %q", enc)
		}
	})

	t.Run("empty to", func(t *testing.T) {
		p := &EmailPayload{To: ""}
		if err := p.Validate(); err == nil {
			t.Error("expected error for empty To")
		}
	})
}

func TestVCardPayload(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		p := &VCardPayload{FirstName: "John", LastName: "Doe", Phone: "+1234", Email: "j@doe.com"}
		enc, err := p.Encode()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p.Type() != "vcard" {
			t.Errorf("expected 'vcard', got %q", p.Type())
		}
		if !containsStr(enc, "BEGIN:VCARD") || !containsStr(enc, "END:VCARD") {
			t.Errorf("vCard should have BEGIN/END markers, got: %s", enc)
		}
	})

	t.Run("default version", func(t *testing.T) {
		p := &VCardPayload{FirstName: "Jane"}
		enc, _ := p.Encode()
		if !containsStr(enc, "VERSION:3.0") {
			t.Errorf("default version should be 3.0, got: %s", enc)
		}
	})

	t.Run("custom version", func(t *testing.T) {
		p := &VCardPayload{FirstName: "Jane", Version: "4.0"}
		enc, _ := p.Encode()
		if !containsStr(enc, "VERSION:4.0") {
			t.Errorf("expected version 4.0, got: %s", enc)
		}
	})

	t.Run("empty name", func(t *testing.T) {
		p := &VCardPayload{}
		if err := p.Validate(); err == nil {
			t.Error("expected error for empty name")
		}
	})

	t.Run("invalid version", func(t *testing.T) {
		p := &VCardPayload{FirstName: "Jane", Version: "5.0"}
		if err := p.Validate(); err == nil {
			t.Error("expected error for invalid version")
		}
	})
}

func TestGeoPayload(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		p := &GeoPayload{Latitude: 37.7749, Longitude: -122.4194}
		enc, err := p.Encode()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p.Type() != "geo" {
			t.Errorf("expected 'geo', got %q", p.Type())
		}
		if !containsStr(enc, "geo:37.7749,-122.4194") {
			t.Errorf("got %q", enc)
		}
	})

	t.Run("latitude out of range", func(t *testing.T) {
		p := &GeoPayload{Latitude: 91, Longitude: 0}
		if err := p.Validate(); err == nil {
			t.Error("expected error for latitude > 90")
		}
	})

	t.Run("longitude out of range", func(t *testing.T) {
		p := &GeoPayload{Latitude: 0, Longitude: -181}
		if err := p.Validate(); err == nil {
			t.Error("expected error for longitude < -180")
		}
	})
}

func TestIsValidEncryption(t *testing.T) {
	valid := []string{EncryptionWEP, EncryptionWPA, EncryptionWPA2, EncryptionWPA3, EncryptionSAE, EncryptionNoPass}
	for _, enc := range valid {
		if !isValidEncryption(enc) {
			t.Errorf("expected %q to be valid", enc)
		}
	}
	if isValidEncryption("FAKE") {
		t.Error("FAKE should not be valid")
	}
}

func TestEscapeWiFi(t *testing.T) {
	if escapeWiFi("hello") != "hello" {
		t.Error("no special chars should pass through")
	}
	escaped := escapeWiFi("a;b")
	if escaped == "a;b" {
		t.Error("semicolon should be escaped")
	}
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
