package payload

import (
	"net/url"
	"strings"
	"testing"
	"time"
)

// =============================================================================
// Social Payloads
// =============================================================================

func TestTwitterPayload(t *testing.T) {
	t.Run("valid encode", func(t *testing.T) {
		p := &TwitterPayload{Username: "jack"}
		enc, err := p.Encode()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := "https://twitter.com/jack"
		if enc != want {
			t.Errorf("expected %q, got %q", want, enc)
		}
	})

	t.Run("type", func(t *testing.T) {
		p := &TwitterPayload{Username: "jack"}
		if p.Type() != "twitter" {
			t.Errorf("expected %q, got %q", "twitter", p.Type())
		}
	})

	t.Run("size positive", func(t *testing.T) {
		p := &TwitterPayload{Username: "jack"}
		if p.Size() <= 0 {
			t.Errorf("expected positive size, got %d", p.Size())
		}
	})

	t.Run("validate empty username", func(t *testing.T) {
		p := &TwitterPayload{}
		if err := p.Validate(); err == nil {
			t.Error("expected error for empty username")
		}
	})

	t.Run("encode fails on empty", func(t *testing.T) {
		p := &TwitterPayload{}
		if _, err := p.Encode(); err == nil {
			t.Error("expected encode error for empty username")
		}
	})
}

func TestTwitterFollowPayload(t *testing.T) {
	t.Run("valid encode", func(t *testing.T) {
		p := &TwitterFollowPayload{ScreenName: "elonmusk"}
		enc, err := p.Encode()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := "https://twitter.com/intent/follow?screen_name=elonmusk"
		if enc != want {
			t.Errorf("expected %q, got %q", want, enc)
		}
	})

	t.Run("type", func(t *testing.T) {
		p := &TwitterFollowPayload{ScreenName: "user"}
		if p.Type() != "twitter_follow" {
			t.Errorf("expected %q, got %q", "twitter_follow", p.Type())
		}
	})

	t.Run("size positive", func(t *testing.T) {
		p := &TwitterFollowPayload{ScreenName: "user"}
		if p.Size() <= 0 {
			t.Errorf("expected positive size, got %d", p.Size())
		}
	})

	t.Run("validate empty screen name", func(t *testing.T) {
		p := &TwitterFollowPayload{}
		if err := p.Validate(); err == nil {
			t.Error("expected error for empty screen name")
		}
	})
}

func TestLinkedInPayload(t *testing.T) {
	t.Run("valid encode", func(t *testing.T) {
		p := &LinkedInPayload{ProfileURL: "https://linkedin.com/in/johndoe"}
		enc, err := p.Encode()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := "https://linkedin.com/in/johndoe"
		if enc != want {
			t.Errorf("expected %q, got %q", want, enc)
		}
	})

	t.Run("type", func(t *testing.T) {
		p := &LinkedInPayload{ProfileURL: "https://linkedin.com/in/johndoe"}
		if p.Type() != "linkedin" {
			t.Errorf("expected %q, got %q", "linkedin", p.Type())
		}
	})

	t.Run("size positive", func(t *testing.T) {
		p := &LinkedInPayload{ProfileURL: "https://linkedin.com/in/johndoe"}
		if p.Size() <= 0 {
			t.Errorf("expected positive size, got %d", p.Size())
		}
	})

	t.Run("validate empty URL", func(t *testing.T) {
		p := &LinkedInPayload{}
		if err := p.Validate(); err == nil {
			t.Error("expected error for empty profile URL")
		}
	})

	t.Run("validate not https", func(t *testing.T) {
		p := &LinkedInPayload{ProfileURL: "http://linkedin.com/in/johndoe"}
		if err := p.Validate(); err == nil {
			t.Error("expected error for non-https URL")
		}
	})
}

func TestInstagramPayload(t *testing.T) {
	t.Run("valid encode", func(t *testing.T) {
		p := &InstagramPayload{Username: "natgeo"}
		enc, err := p.Encode()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := "https://instagram.com/natgeo"
		if enc != want {
			t.Errorf("expected %q, got %q", want, enc)
		}
	})

	t.Run("type", func(t *testing.T) {
		p := &InstagramPayload{Username: "user"}
		if p.Type() != "instagram" {
			t.Errorf("expected %q, got %q", "instagram", p.Type())
		}
	})

	t.Run("size positive", func(t *testing.T) {
		p := &InstagramPayload{Username: "user"}
		if p.Size() <= 0 {
			t.Errorf("expected positive size, got %d", p.Size())
		}
	})

	t.Run("validate empty username", func(t *testing.T) {
		p := &InstagramPayload{}
		if err := p.Validate(); err == nil {
			t.Error("expected error for empty username")
		}
	})
}

func TestFacebookPayload(t *testing.T) {
	t.Run("valid encode", func(t *testing.T) {
		p := &FacebookPayload{PageURL: "https://facebook.com/golang"}
		enc, err := p.Encode()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := "https://facebook.com/golang"
		if enc != want {
			t.Errorf("expected %q, got %q", want, enc)
		}
	})

	t.Run("type", func(t *testing.T) {
		p := &FacebookPayload{PageURL: "https://facebook.com/x"}
		if p.Type() != "facebook" {
			t.Errorf("expected %q, got %q", "facebook", p.Type())
		}
	})

	t.Run("size positive", func(t *testing.T) {
		p := &FacebookPayload{PageURL: "https://facebook.com/x"}
		if p.Size() <= 0 {
			t.Errorf("expected positive size, got %d", p.Size())
		}
	})

	t.Run("validate empty URL", func(t *testing.T) {
		p := &FacebookPayload{}
		if err := p.Validate(); err == nil {
			t.Error("expected error for empty page URL")
		}
	})

	t.Run("validate not https", func(t *testing.T) {
		p := &FacebookPayload{PageURL: "http://facebook.com/x"}
		if err := p.Validate(); err == nil {
			t.Error("expected error for non-https URL")
		}
	})
}

func TestYouTubeChannelPayload(t *testing.T) {
	t.Run("valid encode", func(t *testing.T) {
		p := &YouTubeChannelPayload{ChannelID: "UC_x5XG1OV2P6uZZ5FSM9Ttw"}
		enc, err := p.Encode()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := "https://www.youtube.com/channel/UC_x5XG1OV2P6uZZ5FSM9Ttw"
		if enc != want {
			t.Errorf("expected %q, got %q", want, enc)
		}
	})

	t.Run("type", func(t *testing.T) {
		p := &YouTubeChannelPayload{ChannelID: "abc123"}
		if p.Type() != "youtube_channel" {
			t.Errorf("expected %q, got %q", "youtube_channel", p.Type())
		}
	})

	t.Run("size positive", func(t *testing.T) {
		p := &YouTubeChannelPayload{ChannelID: "abc123"}
		if p.Size() <= 0 {
			t.Errorf("expected positive size, got %d", p.Size())
		}
	})

	t.Run("validate empty channel ID", func(t *testing.T) {
		p := &YouTubeChannelPayload{}
		if err := p.Validate(); err == nil {
			t.Error("expected error for empty channel ID")
		}
	})
}

func TestTelegramPayload(t *testing.T) {
	t.Run("valid encode", func(t *testing.T) {
		p := &TelegramPayload{Username: "durov"}
		enc, err := p.Encode()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := "https://t.me/durov"
		if enc != want {
			t.Errorf("expected %q, got %q", want, enc)
		}
	})

	t.Run("type", func(t *testing.T) {
		p := &TelegramPayload{Username: "user"}
		if p.Type() != "telegram" {
			t.Errorf("expected %q, got %q", "telegram", p.Type())
		}
	})

	t.Run("size positive", func(t *testing.T) {
		p := &TelegramPayload{Username: "user"}
		if p.Size() <= 0 {
			t.Errorf("expected positive size, got %d", p.Size())
		}
	})

	t.Run("validate empty username", func(t *testing.T) {
		p := &TelegramPayload{}
		if err := p.Validate(); err == nil {
			t.Error("expected error for empty username")
		}
	})
}

// =============================================================================
// Market Payload
// =============================================================================

func TestMarketPayload(t *testing.T) {
	t.Run("google with PackageID", func(t *testing.T) {
		p := &MarketPayload{Platform: MarketGooglePlay, PackageID: "com.example.app"}
		enc, err := p.Encode()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.HasPrefix(enc, "https://play.google.com/store/apps/details?id=") {
			t.Errorf("unexpected google play URL: %s", enc)
		}
		if !strings.HasSuffix(enc, url.QueryEscape("com.example.app")) {
			t.Errorf("URL should contain escaped PackageID: %s", enc)
		}
	})

	t.Run("google with AppName", func(t *testing.T) {
		p := &MarketPayload{Platform: MarketGooglePlay, AppName: "My App"}
		enc, err := p.Encode()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.HasPrefix(enc, "https://play.google.com/store/search?q=") {
			t.Errorf("unexpected google play search URL: %s", enc)
		}
	})

	t.Run("apple with PackageID", func(t *testing.T) {
		p := &MarketPayload{Platform: MarketAppleApp, PackageID: "1234567890"}
		enc, err := p.Encode()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.HasPrefix(enc, "https://apps.apple.com/app/id") {
			t.Errorf("unexpected apple URL: %s", enc)
		}
		if !strings.HasSuffix(enc, "1234567890") {
			t.Errorf("URL should contain PackageID: %s", enc)
		}
	})

	t.Run("apple with AppName", func(t *testing.T) {
		p := &MarketPayload{Platform: MarketAppleApp, AppName: "Super App"}
		enc, err := p.Encode()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.HasPrefix(enc, "https://apps.apple.com/search?term=") {
			t.Errorf("unexpected apple search URL: %s", enc)
		}
	})

	t.Run("empty platform defaults to google", func(t *testing.T) {
		p := &MarketPayload{PackageID: "com.default.app"}
		enc, err := p.Encode()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.HasPrefix(enc, "https://play.google.com/store/apps/details?id=") {
			t.Errorf("empty platform should default to google play: %s", enc)
		}
	})

	t.Run("with campaign UTM params", func(t *testing.T) {
		p := &MarketPayload{Platform: MarketGooglePlay, PackageID: "com.test", Campaign: "summer2024"}
		enc, err := p.Encode()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(enc, "utm_source=qr") {
			t.Errorf("should contain utm_source: %s", enc)
		}
		if !strings.Contains(enc, "utm_medium=scan") {
			t.Errorf("should contain utm_medium: %s", enc)
		}
		if !strings.Contains(enc, "utm_campaign=summer2024") {
			t.Errorf("should contain utm_campaign: %s", enc)
		}
	})

	t.Run("campaign with search URL uses ampersand separator", func(t *testing.T) {
		p := &MarketPayload{Platform: MarketGooglePlay, AppName: "TestApp", Campaign: "promo"}
		enc, err := p.Encode()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// The base URL already has "?" from search query, so campaign should use "&"
		if !strings.Contains(enc, "q=TestApp&utm_source=qr") {
			t.Errorf("campaign should join with & after existing query: %s", enc)
		}
	})

	t.Run("type", func(t *testing.T) {
		p := &MarketPayload{PackageID: "com.test"}
		if p.Type() != "market" {
			t.Errorf("expected %q, got %q", "market", p.Type())
		}
	})

	t.Run("size positive", func(t *testing.T) {
		p := &MarketPayload{PackageID: "com.test"}
		if p.Size() <= 0 {
			t.Errorf("expected positive size, got %d", p.Size())
		}
	})

	t.Run("validate invalid platform", func(t *testing.T) {
		p := &MarketPayload{Platform: "windows", PackageID: "com.test"}
		if err := p.Validate(); err == nil {
			t.Error("expected error for invalid platform")
		}
	})

	t.Run("validate empty both PackageID and AppName", func(t *testing.T) {
		p := &MarketPayload{Platform: MarketGooglePlay}
		if err := p.Validate(); err == nil {
			t.Error("expected error when both PackageID and AppName are empty")
		}
	})
}

// =============================================================================
// Maps Payloads
// =============================================================================

func TestGoogleMapsPayload(t *testing.T) {
	t.Run("valid coordinates", func(t *testing.T) {
		p := &GoogleMapsPayload{Latitude: 37.7749, Longitude: -122.4194}
		enc, err := p.Encode()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.HasPrefix(enc, "https://maps.google.com/?q=") {
			t.Errorf("unexpected maps URL: %s", enc)
		}
		if !containsStr(enc, "37.7749") || !containsStr(enc, "-122.4194") {
			t.Errorf("should contain lat/lng: %s", enc)
		}
	})

	t.Run("with query", func(t *testing.T) {
		p := &GoogleMapsPayload{Query: "Golden Gate Bridge"}
		enc, err := p.Encode()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(enc, "q=Golden+Gate+Bridge") {
			t.Errorf("should contain query: %s", enc)
		}
	})

	t.Run("with zoom", func(t *testing.T) {
		p := &GoogleMapsPayload{Latitude: 40.7128, Longitude: -74.006, Zoom: 15}
		enc, err := p.Encode()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(enc, "&zoom=15") {
			t.Errorf("should contain zoom: %s", enc)
		}
	})

	t.Run("zero zoom is omitted", func(t *testing.T) {
		p := &GoogleMapsPayload{Latitude: 40.7128, Longitude: -74.006, Zoom: 0}
		enc, err := p.Encode()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if strings.Contains(enc, "zoom") {
			t.Errorf("zoom=0 should be omitted: %s", enc)
		}
	})

	t.Run("type", func(t *testing.T) {
		p := &GoogleMapsPayload{Query: "test"}
		if p.Type() != "google_maps" {
			t.Errorf("expected %q, got %q", "google_maps", p.Type())
		}
	})

	t.Run("size positive", func(t *testing.T) {
		p := &GoogleMapsPayload{Query: "test"}
		if p.Size() <= 0 {
			t.Errorf("expected positive size, got %d", p.Size())
		}
	})

	t.Run("latitude out of range", func(t *testing.T) {
		p := &GoogleMapsPayload{Latitude: 91, Longitude: 0}
		if err := p.Validate(); err == nil {
			t.Error("expected error for latitude > 90")
		}
	})

	t.Run("longitude out of range", func(t *testing.T) {
		p := &GoogleMapsPayload{Latitude: 0, Longitude: -181}
		if err := p.Validate(); err == nil {
			t.Error("expected error for longitude < -180")
		}
	})

	t.Run("boundary latitude 90", func(t *testing.T) {
		p := &GoogleMapsPayload{Latitude: 90, Longitude: 0}
		if err := p.Validate(); err != nil {
			t.Errorf("latitude 90 should be valid: %v", err)
		}
	})

	t.Run("boundary latitude -90", func(t *testing.T) {
		p := &GoogleMapsPayload{Latitude: -90, Longitude: 0}
		if err := p.Validate(); err != nil {
			t.Errorf("latitude -90 should be valid: %v", err)
		}
	})

	t.Run("boundary longitude 180", func(t *testing.T) {
		p := &GoogleMapsPayload{Latitude: 0, Longitude: 180}
		if err := p.Validate(); err != nil {
			t.Errorf("longitude 180 should be valid: %v", err)
		}
	})

	t.Run("boundary longitude -180", func(t *testing.T) {
		p := &GoogleMapsPayload{Latitude: 0, Longitude: -180}
		if err := p.Validate(); err != nil {
			t.Errorf("longitude -180 should be valid: %v", err)
		}
	})

	t.Run("query bypasses coordinate validation", func(t *testing.T) {
		p := &GoogleMapsPayload{Query: "Paris", Latitude: 999, Longitude: 999}
		if err := p.Validate(); err != nil {
			t.Errorf("query should bypass coordinate check: %v", err)
		}
	})
}

func TestGoogleMapsPlacePayload(t *testing.T) {
	t.Run("valid encode", func(t *testing.T) {
		p := &GoogleMapsPlacePayload{PlaceName: "Eiffel Tower"}
		enc, err := p.Encode()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.HasPrefix(enc, "https://www.google.com/maps/place/") {
			t.Errorf("unexpected place URL: %s", enc)
		}
		if !strings.Contains(enc, url.QueryEscape("Eiffel Tower")) {
			t.Errorf("should contain escaped place name: %s", enc)
		}
	})

	t.Run("type", func(t *testing.T) {
		p := &GoogleMapsPlacePayload{PlaceName: "test"}
		if p.Type() != "google_maps_place" {
			t.Errorf("expected %q, got %q", "google_maps_place", p.Type())
		}
	})

	t.Run("size positive", func(t *testing.T) {
		p := &GoogleMapsPlacePayload{PlaceName: "test"}
		if p.Size() <= 0 {
			t.Errorf("expected positive size, got %d", p.Size())
		}
	})

	t.Run("validate empty place name", func(t *testing.T) {
		p := &GoogleMapsPlacePayload{}
		if err := p.Validate(); err == nil {
			t.Error("expected error for empty place name")
		}
	})
}

func TestGoogleMapsDirectionsPayload(t *testing.T) {
	t.Run("valid with default travel mode", func(t *testing.T) {
		p := &GoogleMapsDirectionsPayload{Origin: "New York", Destination: "Boston"}
		enc, err := p.Encode()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.HasPrefix(enc, "https://maps.google.com/maps/dir/") {
			t.Errorf("unexpected directions URL: %s", enc)
		}
		if !strings.Contains(enc, "travelmode=driving") {
			t.Errorf("default travel mode should be driving: %s", enc)
		}
		if !strings.Contains(enc, "origin=New+York") {
			t.Errorf("should contain origin: %s", enc)
		}
		if !strings.Contains(enc, "destination=Boston") {
			t.Errorf("should contain destination: %s", enc)
		}
	})

	t.Run("custom travel mode", func(t *testing.T) {
		p := &GoogleMapsDirectionsPayload{
			Origin: "A", Destination: "B", TravelMode: TravelModeWalking,
		}
		enc, err := p.Encode()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(enc, "travelmode=walking") {
			t.Errorf("should contain walking mode: %s", enc)
		}
	})

	t.Run("bicycling travel mode", func(t *testing.T) {
		p := &GoogleMapsDirectionsPayload{
			Origin: "A", Destination: "B", TravelMode: TravelModeBicycling,
		}
		enc, _ := p.Encode()
		if !strings.Contains(enc, "travelmode=bicycling") {
			t.Errorf("should contain bicycling mode: %s", enc)
		}
	})

	t.Run("transit travel mode", func(t *testing.T) {
		p := &GoogleMapsDirectionsPayload{
			Origin: "A", Destination: "B", TravelMode: TravelModeTransit,
		}
		enc, _ := p.Encode()
		if !strings.Contains(enc, "travelmode=transit") {
			t.Errorf("should contain transit mode: %s", enc)
		}
	})

	t.Run("type", func(t *testing.T) {
		p := &GoogleMapsDirectionsPayload{Origin: "A", Destination: "B"}
		if p.Type() != "google_maps_directions" {
			t.Errorf("expected %q, got %q", "google_maps_directions", p.Type())
		}
	})

	t.Run("size positive", func(t *testing.T) {
		p := &GoogleMapsDirectionsPayload{Origin: "A", Destination: "B"}
		if p.Size() <= 0 {
			t.Errorf("expected positive size, got %d", p.Size())
		}
	})

	t.Run("validate empty origin", func(t *testing.T) {
		p := &GoogleMapsDirectionsPayload{Destination: "B"}
		if err := p.Validate(); err == nil {
			t.Error("expected error for empty origin")
		}
	})

	t.Run("validate empty destination", func(t *testing.T) {
		p := &GoogleMapsDirectionsPayload{Origin: "A"}
		if err := p.Validate(); err == nil {
			t.Error("expected error for empty destination")
		}
	})
}

func TestAppleMapsPayload(t *testing.T) {
	t.Run("valid coordinates", func(t *testing.T) {
		p := &AppleMapsPayload{Latitude: 37.7749, Longitude: -122.4194}
		enc, err := p.Encode()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.HasPrefix(enc, "https://maps.apple.com/?ll=") {
			t.Errorf("unexpected apple maps URL: %s", enc)
		}
	})

	t.Run("with query", func(t *testing.T) {
		p := &AppleMapsPayload{Latitude: 37.7749, Longitude: -122.4194, Query: "Coffee"}
		enc, err := p.Encode()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(enc, "&q=Coffee") {
			t.Errorf("should contain query param: %s", enc)
		}
	})

	t.Run("with zoom", func(t *testing.T) {
		p := &AppleMapsPayload{Latitude: 37.7749, Longitude: -122.4194, Zoom: 10}
		enc, err := p.Encode()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(enc, "&t=10") {
			t.Errorf("should contain zoom: %s", enc)
		}
	})

	t.Run("type", func(t *testing.T) {
		p := &AppleMapsPayload{Latitude: 1, Longitude: 1}
		if p.Type() != "apple_maps" {
			t.Errorf("expected %q, got %q", "apple_maps", p.Type())
		}
	})

	t.Run("size positive", func(t *testing.T) {
		p := &AppleMapsPayload{Latitude: 1, Longitude: 1}
		if p.Size() <= 0 {
			t.Errorf("expected positive size, got %d", p.Size())
		}
	})

	t.Run("latitude out of range", func(t *testing.T) {
		p := &AppleMapsPayload{Latitude: 91, Longitude: 0}
		if err := p.Validate(); err == nil {
			t.Error("expected error for latitude > 90")
		}
	})

	t.Run("longitude out of range", func(t *testing.T) {
		p := &AppleMapsPayload{Latitude: 0, Longitude: 200}
		if err := p.Validate(); err == nil {
			t.Error("expected error for longitude > 180")
		}
	})

	t.Run("zero zoom omitted", func(t *testing.T) {
		p := &AppleMapsPayload{Latitude: 1, Longitude: 1, Zoom: 0}
		enc, err := p.Encode()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if strings.Contains(enc, "&t=") {
			t.Errorf("zoom=0 should be omitted: %s", enc)
		}
	})
}

// =============================================================================
// PayPal Payload
// =============================================================================

func TestPayPalPayload(t *testing.T) {
	t.Run("valid with all fields", func(t *testing.T) {
		p := &PayPalPayload{Username: "johndoe", Amount: "25.00", Currency: "EUR", Reference: "Coffee"}
		enc, err := p.Encode()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.HasPrefix(enc, "https://www.paypal.me/") {
			t.Errorf("unexpected paypal URL: %s", enc)
		}
		if !strings.Contains(enc, "johndoe") {
			t.Errorf("should contain username: %s", enc)
		}
		if !strings.Contains(enc, "25.00") {
			t.Errorf("should contain amount: %s", enc)
		}
		if !strings.Contains(enc, "EUR") {
			t.Errorf("should contain currency: %s", enc)
		}
		if !strings.Contains(enc, "&note=Coffee") {
			t.Errorf("should contain reference note: %s", enc)
		}
	})

	t.Run("currency defaults to USD", func(t *testing.T) {
		p := &PayPalPayload{Username: "user", Amount: "10"}
		enc, err := p.Encode()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(enc, "/USD") {
			t.Errorf("default currency should be USD: %s", enc)
		}
	})

	t.Run("without reference", func(t *testing.T) {
		p := &PayPalPayload{Username: "user", Amount: "10"}
		enc, err := p.Encode()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if strings.Contains(enc, "&note=") {
			t.Errorf("should not contain note when reference is empty: %s", enc)
		}
	})

	t.Run("type", func(t *testing.T) {
		p := &PayPalPayload{Username: "u", Amount: "1"}
		if p.Type() != "paypal" {
			t.Errorf("expected %q, got %q", "paypal", p.Type())
		}
	})

	t.Run("size positive", func(t *testing.T) {
		p := &PayPalPayload{Username: "u", Amount: "1"}
		if p.Size() <= 0 {
			t.Errorf("expected positive size, got %d", p.Size())
		}
	})

	t.Run("validate empty username", func(t *testing.T) {
		p := &PayPalPayload{Amount: "10"}
		if err := p.Validate(); err == nil {
			t.Error("expected error for empty username")
		}
	})

	t.Run("validate empty amount", func(t *testing.T) {
		p := &PayPalPayload{Username: "user"}
		if err := p.Validate(); err == nil {
			t.Error("expected error for empty amount")
		}
	})

	t.Run("validate both empty", func(t *testing.T) {
		p := &PayPalPayload{}
		if err := p.Validate(); err == nil {
			t.Error("expected error for empty username and amount")
		}
	})
}

// =============================================================================
// Crypto Payload
// =============================================================================

func TestCryptoPayload(t *testing.T) {
	t.Run("BTC with amount and label", func(t *testing.T) {
		p := &CryptoPayload{
			Address: "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa",
			Amount:  "0.5", Label: "Donation", CryptoType: CryptoBTC,
		}
		enc, err := p.Encode()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.HasPrefix(enc, "bitcoin:1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa") {
			t.Errorf("unexpected BTC URL: %s", enc)
		}
		if !strings.Contains(enc, "amount=0.5") {
			t.Errorf("should contain amount: %s", enc)
		}
		if !strings.Contains(enc, "label=Donation") {
			t.Errorf("should contain label: %s", enc)
		}
	})

	t.Run("ETH with message", func(t *testing.T) {
		p := &CryptoPayload{
			Address: "0x742d35Cc6634C0532925a3b844Bc9e7595f2bD08",
			Message: "For services", CryptoType: CryptoETH,
		}
		enc, err := p.Encode()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.HasPrefix(enc, "ethereum:0x742d35Cc6634C0532925a3b844Bc9e7595f2bD08") {
			t.Errorf("unexpected ETH URL: %s", enc)
		}
		if !strings.Contains(enc, "message=For services") {
			t.Errorf("should contain message: %s", enc)
		}
	})

	t.Run("LTC minimal", func(t *testing.T) {
		p := &CryptoPayload{
			Address:    "LSN5D4PGRXWS3YFZNKV5XDYZJV5V3RHLTL",
			CryptoType: CryptoLTC,
		}
		enc, err := p.Encode()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.HasPrefix(enc, "litecoin:") {
			t.Errorf("unexpected LTC URL: %s", enc)
		}
		// No params should mean no "?" separator
		if strings.Contains(enc, "?") {
			t.Errorf("minimal payload should not have query: %s", enc)
		}
	})

	t.Run("all optional fields", func(t *testing.T) {
		p := &CryptoPayload{
			Address: "1A1zP1", Amount: "1.0", Label: "L", Message: "M", CryptoType: CryptoBTC,
		}
		enc, _ := p.Encode()
		if !strings.Contains(enc, "amount=1.0") {
			t.Errorf("should contain amount: %s", enc)
		}
		if !strings.Contains(enc, "label=L") {
			t.Errorf("should contain label: %s", enc)
		}
		if !strings.Contains(enc, "message=M") {
			t.Errorf("should contain message: %s", enc)
		}
	})

	t.Run("type", func(t *testing.T) {
		p := &CryptoPayload{Address: "a", CryptoType: CryptoBTC}
		if p.Type() != "crypto" {
			t.Errorf("expected %q, got %q", "crypto", p.Type())
		}
	})

	t.Run("size positive", func(t *testing.T) {
		p := &CryptoPayload{Address: "a", CryptoType: CryptoBTC}
		if p.Size() <= 0 {
			t.Errorf("expected positive size, got %d", p.Size())
		}
	})

	t.Run("validate empty address", func(t *testing.T) {
		p := &CryptoPayload{CryptoType: CryptoBTC}
		if err := p.Validate(); err == nil {
			t.Error("expected error for empty address")
		}
	})

	t.Run("validate invalid crypto type", func(t *testing.T) {
		p := &CryptoPayload{Address: "abc123", CryptoType: "XRP"}
		if err := p.Validate(); err == nil {
			t.Error("expected error for invalid crypto type")
		}
	})
}

func TestIsValidCryptoType(t *testing.T) {
	tests := []struct {
		ct   string
		want bool
	}{
		{CryptoBTC, true},
		{CryptoETH, true},
		{CryptoLTC, true},
		{"XRP", false},
		{"", false},
		{"btc", false},
		{"ETH ", false},
	}
	for _, tt := range tests {
		t.Run(tt.ct, func(t *testing.T) {
			got := isValidCryptoType(tt.ct)
			if got != tt.want {
				t.Errorf("isValidCryptoType(%q) = %v, want %v", tt.ct, got, tt.want)
			}
		})
	}
}

func TestCryptoScheme(t *testing.T) {
	tests := []struct {
		ct   string
		want string
	}{
		{CryptoBTC, "bitcoin"},
		{CryptoETH, "ethereum"},
		{CryptoLTC, "litecoin"},
		{"DOGE", "doge"},
	}
	for _, tt := range tests {
		t.Run(tt.ct, func(t *testing.T) {
			got := cryptoScheme(tt.ct)
			if got != tt.want {
				t.Errorf("cryptoScheme(%q) = %q, want %q", tt.ct, got, tt.want)
			}
		})
	}
}

// =============================================================================
// IBeacon Payload
// =============================================================================

func TestIBeaconPayload(t *testing.T) {
	t.Run("valid encode", func(t *testing.T) {
		uuid := "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
		p := &IBeaconPayload{UUID: uuid, Major: 1, Minor: 2}
		enc, err := p.Encode()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.HasPrefix(enc, "https://beacon.github.io/beacon/") {
			t.Errorf("unexpected URL prefix: %s", enc)
		}
		// UUID should be uppercased in output
		if !strings.Contains(enc, "uuid=A1B2C3D4") {
			t.Errorf("UUID should be uppercased: %s", enc)
		}
		if !strings.Contains(enc, "&major=1") {
			t.Errorf("should contain major: %s", enc)
		}
		if !strings.Contains(enc, "&minor=2") {
			t.Errorf("should contain minor: %s", enc)
		}
	})

	t.Run("custom manufacturer", func(t *testing.T) {
		p := &IBeaconPayload{
			UUID:  "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
			Major: 10, Minor: 20, Manufacturer: "MyBrand",
		}
		enc, err := p.Encode()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.HasPrefix(enc, "https://mybrand.github.io/beacon/") {
			t.Errorf("manufacturer should be lowercased: %s", enc)
		}
	})

	t.Run("type", func(t *testing.T) {
		p := &IBeaconPayload{UUID: "a1b2c3d4-e5f6-7890-abcd-ef1234567890"}
		if p.Type() != "ibeacon" {
			t.Errorf("expected %q, got %q", "ibeacon", p.Type())
		}
	})

	t.Run("size positive", func(t *testing.T) {
		p := &IBeaconPayload{UUID: "a1b2c3d4-e5f6-7890-abcd-ef1234567890"}
		if p.Size() <= 0 {
			t.Errorf("expected positive size, got %d", p.Size())
		}
	})

	t.Run("validate empty UUID", func(t *testing.T) {
		p := &IBeaconPayload{}
		if err := p.Validate(); err == nil {
			t.Error("expected error for empty UUID")
		}
	})

	t.Run("validate invalid UUID format", func(t *testing.T) {
		p := &IBeaconPayload{UUID: "not-a-uuid"}
		if err := p.Validate(); err == nil {
			t.Error("expected error for invalid UUID")
		}
	})

	t.Run("validate invalid UUID chars", func(t *testing.T) {
		p := &IBeaconPayload{UUID: "gggggggg-gggg-gggg-gggg-gggggggggggg"}
		if err := p.Validate(); err == nil {
			t.Error("expected error for UUID with non-hex chars")
		}
	})
}

func TestIsHexDigit(t *testing.T) {
	tests := []struct {
		c    byte
		want bool
	}{
		{'0', true},
		{'9', true},
		{'a', true},
		{'f', true},
		{'A', true},
		{'F', true},
		{'g', false},
		{'z', false},
		{'G', false},
		{'Z', false},
		{' ', false},
		{'-', false},
		{0, false},
		{127, false},
	}
	for _, tt := range tests {
		name := string([]byte{tt.c})
		if tt.c < 32 || tt.c > 126 {
			name = "non-printable"
		}
		t.Run(name, func(t *testing.T) {
			got := isHexDigit(tt.c)
			if got != tt.want {
				t.Errorf("isHexDigit(%q) = %v, want %v", tt.c, got, tt.want)
			}
		})
	}
}

func TestIsValidUUID(t *testing.T) {
	tests := []struct {
		name string
		uuid string
		want bool
	}{
		{"valid lower", "a1b2c3d4-e5f6-7890-abcd-ef1234567890", true},
		{"valid upper", "A1B2C3D4-E5F6-7890-ABCD-EF1234567890", true},
		{"valid mixed", "A1b2C3d4-e5F6-7890-AbCd-eF1234567890", true},
		{"too short", "a1b2c3d4-e5f6-7890-abcd", false},
		{"too long", "a1b2c3d4-e5f6-7890-abcd-ef1234567890-extra", false},
		{"invalid chars", "gggggggg-gggg-gggg-gggg-gggggggggggg", false},
		{"empty", "", false},
		{"no dashes", "a1b2c3d4e5f67890abcdef1234567890", true},
		{"wrong dash positions", "a1b2c3d4-e5f67890-abcd-ef1234567890", true},
		{"31 hex chars", "a1b2c3d4-e5f6-7890-abcd-ef123456789", false},
		{"33 hex chars", "a1b2c3d4-e5f6-7890-abcd-ef12345678900", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidUUID(tt.uuid)
			if got != tt.want {
				t.Errorf("isValidUUID(%q) = %v, want %v", tt.uuid, got, tt.want)
			}
		})
	}
}

// =============================================================================
// Media Payloads
// =============================================================================

func TestSpotifyTrackPayload(t *testing.T) {
	t.Run("valid encode", func(t *testing.T) {
		p := &SpotifyTrackPayload{TrackID: "4cOdK2wGLETKBW3PvgPWqT"}
		enc, err := p.Encode()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := "https://open.spotify.com/track/4cOdK2wGLETKBW3PvgPWqT"
		if enc != want {
			t.Errorf("expected %q, got %q", want, enc)
		}
	})

	t.Run("type", func(t *testing.T) {
		p := &SpotifyTrackPayload{TrackID: "abc"}
		if p.Type() != "spotify_track" {
			t.Errorf("expected %q, got %q", "spotify_track", p.Type())
		}
	})

	t.Run("size positive", func(t *testing.T) {
		p := &SpotifyTrackPayload{TrackID: "abc"}
		if p.Size() <= 0 {
			t.Errorf("expected positive size, got %d", p.Size())
		}
	})

	t.Run("validate empty track ID", func(t *testing.T) {
		p := &SpotifyTrackPayload{}
		if err := p.Validate(); err == nil {
			t.Error("expected error for empty track ID")
		}
	})
}

func TestSpotifyPlaylistPayload(t *testing.T) {
	t.Run("valid encode", func(t *testing.T) {
		p := &SpotifyPlaylistPayload{PlaylistID: "37i9dQZF1DXcBWIGoYBM5M"}
		enc, err := p.Encode()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := "https://open.spotify.com/playlist/37i9dQZF1DXcBWIGoYBM5M"
		if enc != want {
			t.Errorf("expected %q, got %q", want, enc)
		}
	})

	t.Run("type", func(t *testing.T) {
		p := &SpotifyPlaylistPayload{PlaylistID: "abc"}
		if p.Type() != "spotify_playlist" {
			t.Errorf("expected %q, got %q", "spotify_playlist", p.Type())
		}
	})

	t.Run("size positive", func(t *testing.T) {
		p := &SpotifyPlaylistPayload{PlaylistID: "abc"}
		if p.Size() <= 0 {
			t.Errorf("expected positive size, got %d", p.Size())
		}
	})

	t.Run("validate empty playlist ID", func(t *testing.T) {
		p := &SpotifyPlaylistPayload{}
		if err := p.Validate(); err == nil {
			t.Error("expected error for empty playlist ID")
		}
	})
}

func TestAppleMusicTrackPayload(t *testing.T) {
	t.Run("valid without storefront", func(t *testing.T) {
		p := &AppleMusicTrackPayload{AlbumID: "123456", SongID: "789012"}
		enc, err := p.Encode()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.HasPrefix(enc, "https://music.apple.com/album/") {
			t.Errorf("unexpected URL: %s", enc)
		}
		if !strings.Contains(enc, "123456") {
			t.Errorf("should contain album ID: %s", enc)
		}
		if !strings.Contains(enc, "i=789012") {
			t.Errorf("should contain song ID: %s", enc)
		}
	})

	t.Run("valid with storefront", func(t *testing.T) {
		p := &AppleMusicTrackPayload{AlbumID: "123456", SongID: "789012", StoreFront: "us"}
		enc, err := p.Encode()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.HasPrefix(enc, "https://music.apple.com/us/album/") {
			t.Errorf("should include storefront in URL: %s", enc)
		}
	})

	t.Run("type", func(t *testing.T) {
		p := &AppleMusicTrackPayload{AlbumID: "a", SongID: "b"}
		if p.Type() != "apple_music" {
			t.Errorf("expected %q, got %q", "apple_music", p.Type())
		}
	})

	t.Run("size positive", func(t *testing.T) {
		p := &AppleMusicTrackPayload{AlbumID: "a", SongID: "b"}
		if p.Size() <= 0 {
			t.Errorf("expected positive size, got %d", p.Size())
		}
	})

	t.Run("validate empty album ID", func(t *testing.T) {
		p := &AppleMusicTrackPayload{SongID: "789"}
		if err := p.Validate(); err == nil {
			t.Error("expected error for empty album ID")
		}
	})

	t.Run("validate empty song ID", func(t *testing.T) {
		p := &AppleMusicTrackPayload{AlbumID: "123"}
		if err := p.Validate(); err == nil {
			t.Error("expected error for empty song ID")
		}
	})

	t.Run("validate both empty", func(t *testing.T) {
		p := &AppleMusicTrackPayload{}
		if err := p.Validate(); err == nil {
			t.Error("expected error for empty album and song ID")
		}
	})
}

func TestYouTubeVideoPayload(t *testing.T) {
	t.Run("valid encode", func(t *testing.T) {
		p := &YouTubeVideoPayload{VideoID: "dQw4w9WgXcQ"}
		enc, err := p.Encode()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := "https://www.youtube.com/watch?v=dQw4w9WgXcQ"
		if enc != want {
			t.Errorf("expected %q, got %q", want, enc)
		}
	})

	t.Run("type", func(t *testing.T) {
		p := &YouTubeVideoPayload{VideoID: "abc"}
		if p.Type() != "youtube_video" {
			t.Errorf("expected %q, got %q", "youtube_video", p.Type())
		}
	})

	t.Run("size positive", func(t *testing.T) {
		p := &YouTubeVideoPayload{VideoID: "abc"}
		if p.Size() <= 0 {
			t.Errorf("expected positive size, got %d", p.Size())
		}
	})

	t.Run("validate empty video ID", func(t *testing.T) {
		p := &YouTubeVideoPayload{}
		if err := p.Validate(); err == nil {
			t.Error("expected error for empty video ID")
		}
	})
}

// =============================================================================
// NTP Locale Payload
// =============================================================================

func TestNTPLocalePayload(t *testing.T) {
	t.Run("host only", func(t *testing.T) {
		p := &NTPLocalePayload{Host: "time.google.com"}
		enc, err := p.Encode()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := "ntp://time.google.com"
		if enc != want {
			t.Errorf("expected %q, got %q", want, enc)
		}
	})

	t.Run("host with custom port", func(t *testing.T) {
		p := &NTPLocalePayload{Host: "time.google.com", Port: "9999"}
		enc, err := p.Encode()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.HasPrefix(enc, "ntp://time.google.com:9999") {
			t.Errorf("should contain custom port: %s", enc)
		}
	})

	t.Run("default port 123 is omitted", func(t *testing.T) {
		p := &NTPLocalePayload{Host: "time.google.com", Port: "123"}
		enc, err := p.Encode()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if strings.Contains(enc, ":123") {
			t.Errorf("port 123 should be omitted: %s", enc)
		}
	})

	t.Run("with description", func(t *testing.T) {
		p := &NTPLocalePayload{Host: "ntp.example.com", Description: "Pool Server"}
		enc, err := p.Encode()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.HasPrefix(enc, "ntp://ntp.example.com#") {
			t.Errorf("should contain description fragment: %s", enc)
		}
		if !strings.Contains(enc, "Pool") {
			t.Errorf("should contain description text: %s", enc)
		}
	})

	t.Run("full payload", func(t *testing.T) {
		p := &NTPLocalePayload{Host: "ntp.example.com", Port: "9999", Description: "Test"}
		enc, err := p.Encode()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.HasPrefix(enc, "ntp://ntp.example.com:9999#") {
			t.Errorf("unexpected full URL: %s", enc)
		}
	})

	t.Run("type", func(t *testing.T) {
		p := &NTPLocalePayload{Host: "h"}
		if p.Type() != "ntp" {
			t.Errorf("expected %q, got %q", "ntp", p.Type())
		}
	})

	t.Run("size positive", func(t *testing.T) {
		p := &NTPLocalePayload{Host: "h"}
		if p.Size() <= 0 {
			t.Errorf("expected positive size, got %d", p.Size())
		}
	})

	t.Run("validate empty host", func(t *testing.T) {
		p := &NTPLocalePayload{}
		if err := p.Validate(); err == nil {
			t.Error("expected error for empty host")
		}
	})

	t.Run("validate non-numeric port", func(t *testing.T) {
		p := &NTPLocalePayload{Host: "ntp.example.com", Port: "abc"}
		if err := p.Validate(); err == nil {
			t.Error("expected error for non-numeric port")
		}
	})

	t.Run("validate port out of range zero", func(t *testing.T) {
		p := &NTPLocalePayload{Host: "ntp.example.com", Port: "0"}
		if err := p.Validate(); err == nil {
			t.Error("expected error for port 0")
		}
	})

	t.Run("validate port out of range high", func(t *testing.T) {
		p := &NTPLocalePayload{Host: "ntp.example.com", Port: "99999"}
		if err := p.Validate(); err == nil {
			t.Error("expected error for port > 65535")
		}
	})

	t.Run("validate port at max 65535", func(t *testing.T) {
		p := &NTPLocalePayload{Host: "ntp.example.com", Port: "65535"}
		if err := p.Validate(); err != nil {
			t.Errorf("port 65535 should be valid: %v", err)
		}
	})

	t.Run("validate port at min 1", func(t *testing.T) {
		p := &NTPLocalePayload{Host: "ntp.example.com", Port: "1"}
		if err := p.Validate(); err != nil {
			t.Errorf("port 1 should be valid: %v", err)
		}
	})

	t.Run("validate invalid version", func(t *testing.T) {
		p := &NTPLocalePayload{Host: "ntp.example.com", Version: 5}
		if err := p.Validate(); err == nil {
			t.Error("expected error for version 5")
		}
	})

	t.Run("validate version 3 valid", func(t *testing.T) {
		p := &NTPLocalePayload{Host: "ntp.example.com", Version: 3}
		if err := p.Validate(); err != nil {
			t.Errorf("version 3 should be valid: %v", err)
		}
	})

	t.Run("validate version 4 valid", func(t *testing.T) {
		p := &NTPLocalePayload{Host: "ntp.example.com", Version: 4}
		if err := p.Validate(); err != nil {
			t.Errorf("version 4 should be valid: %v", err)
		}
	})

	t.Run("validate version 0 valid", func(t *testing.T) {
		p := &NTPLocalePayload{Host: "ntp.example.com", Version: 0}
		if err := p.Validate(); err != nil {
			t.Errorf("version 0 should be valid: %v", err)
		}
	})
}

func TestNTPLocalePayloadString(t *testing.T) {
	t.Run("host only", func(t *testing.T) {
		p := &NTPLocalePayload{Host: "time.google.com"}
		s := p.String()
		if s != "NTP://time.google.com" {
			t.Errorf("expected %q, got %q", "NTP://time.google.com", s)
		}
	})

	t.Run("host with port", func(t *testing.T) {
		p := &NTPLocalePayload{Host: "time.google.com", Port: "9999"}
		s := p.String()
		if s != "NTP://time.google.com:9999" {
			t.Errorf("expected %q, got %q", "NTP://time.google.com:9999", s)
		}
	})

	t.Run("port 123 omitted", func(t *testing.T) {
		p := &NTPLocalePayload{Host: "time.google.com", Port: "123"}
		s := p.String()
		if strings.Contains(s, ":123") {
			t.Errorf("port 123 should be omitted in String(): %s", s)
		}
	})

	t.Run("with version 4", func(t *testing.T) {
		p := &NTPLocalePayload{Host: "time.google.com", Version: 4}
		s := p.String()
		if !strings.Contains(s, "(v4)") {
			t.Errorf("should contain version: %s", s)
		}
	})

	t.Run("version 0 omitted", func(t *testing.T) {
		p := &NTPLocalePayload{Host: "time.google.com", Version: 0}
		s := p.String()
		if strings.Contains(s, "(v") {
			t.Errorf("version 0 should be omitted in String(): %s", s)
		}
	})

	t.Run("full string", func(t *testing.T) {
		p := &NTPLocalePayload{Host: "ntp.example.com", Port: "9999", Version: 3}
		s := p.String()
		if s != "NTP://ntp.example.com:9999 (v3)" {
			t.Errorf("expected %q, got %q", "NTP://ntp.example.com:9999 (v3)", s)
		}
	})
}

// =============================================================================
// PID Payload
// =============================================================================

func TestPIDPayload(t *testing.T) {
	t.Run("valid with QRR default type", func(t *testing.T) {
		p := &PIDPayload{IBAN: "CH44 3199 9123 0008 8901 2"}
		enc, err := p.Encode()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.HasPrefix(enc, "PID:QRR|") {
			t.Errorf("should default to QRR: %s", enc)
		}
		// IBAN spaces should be removed
		if strings.Contains(enc, " ") {
			t.Errorf("IBAN spaces should be removed: %s", enc)
		}
	})

	t.Run("explicit SCOR type", func(t *testing.T) {
		p := &PIDPayload{PIDType: "SCOR", IBAN: "CH9300762011623852957"}
		enc, err := p.Encode()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.HasPrefix(enc, "PID:SCOR|") {
			t.Errorf("should use SCOR type: %s", enc)
		}
	})

	t.Run("NON type", func(t *testing.T) {
		p := &PIDPayload{PIDType: "NON", IBAN: "CH9300762011623852957"}
		enc, err := p.Encode()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.HasPrefix(enc, "PID:NON|") {
			t.Errorf("should use NON type: %s", enc)
		}
	})

	t.Run("with all fields", func(t *testing.T) {
		p := &PIDPayload{
			PIDType: "QRR", CreditorName: "Corp", IBAN: "CH93 0076 2011 6238 5295 7",
			Reference: "RF1234", Amount: "100.00", Currency: "CHF",
			DebtorName: "John", RemittanceInfo: "Invoice 42",
		}
		enc, err := p.Encode()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !containsStr(enc, "Corp") {
			t.Errorf("should contain creditor: %s", enc)
		}
		if !containsStr(enc, "100.00") {
			t.Errorf("should contain amount: %s", enc)
		}
		if !containsStr(enc, "Invoice 42") {
			t.Errorf("should contain remittance info: %s", enc)
		}
	})

	t.Run("IBAN spaces removed", func(t *testing.T) {
		p := &PIDPayload{IBAN: "CH 44 3199 9123 0008 8901 2"}
		enc, _ := p.Encode()
		if strings.Contains(enc, " ") {
			t.Errorf("IBAN spaces should be stripped: %s", enc)
		}
		if !containsStr(enc, "CH4431999123000889012") {
			t.Errorf("should contain compacted IBAN: %s", enc)
		}
	})

	t.Run("type", func(t *testing.T) {
		p := &PIDPayload{IBAN: "CH9300762011623852957"}
		if p.Type() != "pid" {
			t.Errorf("expected %q, got %q", "pid", p.Type())
		}
	})

	t.Run("size positive", func(t *testing.T) {
		p := &PIDPayload{IBAN: "CH9300762011623852957"}
		if p.Size() <= 0 {
			t.Errorf("expected positive size, got %d", p.Size())
		}
	})

	t.Run("validate empty IBAN", func(t *testing.T) {
		p := &PIDPayload{}
		if err := p.Validate(); err == nil {
			t.Error("expected error for empty IBAN")
		}
	})

	t.Run("validate invalid type", func(t *testing.T) {
		p := &PIDPayload{PIDType: "INVALID", IBAN: "CH9300762011623852957"}
		if err := p.Validate(); err == nil {
			t.Error("expected error for invalid PID type")
		}
	})
}

// =============================================================================
// Zoom Payload
// =============================================================================

func TestZoomPayload(t *testing.T) {
	t.Run("valid meeting ID only", func(t *testing.T) {
		p := &ZoomPayload{MeetingID: "1234567890"}
		enc, err := p.Encode()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := "https://zoom.us/j/1234567890"
		if enc != want {
			t.Errorf("expected %q, got %q", want, enc)
		}
	})

	t.Run("with password", func(t *testing.T) {
		p := &ZoomPayload{MeetingID: "1234567890", Password: "abc123"}
		enc, err := p.Encode()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(enc, "?pwd=") {
			t.Errorf("should contain password param: %s", enc)
		}
	})

	t.Run("with display name", func(t *testing.T) {
		p := &ZoomPayload{MeetingID: "1234567890", DisplayName: "John Doe"}
		enc, err := p.Encode()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(enc, "?uname=") {
			t.Errorf("should contain uname param: %s", enc)
		}
	})

	t.Run("with password and display name", func(t *testing.T) {
		p := &ZoomPayload{MeetingID: "1234567890", Password: "pw", DisplayName: "User"}
		enc, err := p.Encode()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(enc, "pwd=") {
			t.Errorf("should contain pwd: %s", enc)
		}
		if !strings.Contains(enc, "uname=") {
			t.Errorf("should contain uname: %s", enc)
		}
	})

	t.Run("type", func(t *testing.T) {
		p := &ZoomPayload{MeetingID: "123"}
		if p.Type() != "zoom" {
			t.Errorf("expected %q, got %q", "zoom", p.Type())
		}
	})

	t.Run("size positive", func(t *testing.T) {
		p := &ZoomPayload{MeetingID: "123"}
		if p.Size() <= 0 {
			t.Errorf("expected positive size, got %d", p.Size())
		}
	})

	t.Run("validate empty meeting ID", func(t *testing.T) {
		p := &ZoomPayload{}
		if err := p.Validate(); err == nil {
			t.Error("expected error for empty meeting ID")
		}
	})
}

// =============================================================================
// MeCard Payload
// =============================================================================

func TestMeCardPayload(t *testing.T) {
	t.Run("name only", func(t *testing.T) {
		p := &MeCardPayload{Name: "John Doe"}
		enc, err := p.Encode()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.HasPrefix(enc, "MECARD:N:John Doe;;") {
			t.Errorf("unexpected mecard output: %s", enc)
		}
	})

	t.Run("with all fields", func(t *testing.T) {
		p := &MeCardPayload{
			Name: "John", Phone: "+1234", Email: "j@doe.com",
			URL: "https://example.com", Birthday: "19900101",
			Note: "Dev", Address: "123 Main St", Nickname: "Johnny",
		}
		enc, err := p.Encode()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !containsStr(enc, "TEL:+1234;") {
			t.Errorf("should contain phone: %s", enc)
		}
		if !containsStr(enc, "EMAIL:j@doe.com;") {
			t.Errorf("should contain email: %s", enc)
		}
		if !containsStr(enc, "URL:https") {
			t.Errorf("should contain URL: %s", enc)
		}
		if !containsStr(enc, "BDAY:19900101;") {
			t.Errorf("should contain birthday: %s", enc)
		}
		if !containsStr(enc, "NOTE:Dev;") {
			t.Errorf("should contain note: %s", enc)
		}
		if !containsStr(enc, "ADR:123 Main St;") {
			t.Errorf("should contain address: %s", enc)
		}
		if !containsStr(enc, "NICKNAME:Johnny;") {
			t.Errorf("should contain nickname: %s", enc)
		}
		// Should end with double semicolon (trailing ";")
		if !strings.HasSuffix(enc, ";;") {
			t.Errorf("should end with double semicolon: %s", enc)
		}
	})

	t.Run("empty optional fields omitted", func(t *testing.T) {
		p := &MeCardPayload{Name: "Jane"}
		enc, _ := p.Encode()
		if containsStr(enc, "TEL:") {
			t.Errorf("empty phone should be omitted: %s", enc)
		}
		if containsStr(enc, "EMAIL:") {
			t.Errorf("empty email should be omitted: %s", enc)
		}
	})

	t.Run("type", func(t *testing.T) {
		p := &MeCardPayload{Name: "John"}
		if p.Type() != "mecard" {
			t.Errorf("expected %q, got %q", "mecard", p.Type())
		}
	})

	t.Run("size positive", func(t *testing.T) {
		p := &MeCardPayload{Name: "John"}
		if p.Size() <= 0 {
			t.Errorf("expected positive size, got %d", p.Size())
		}
	})

	t.Run("validate empty name", func(t *testing.T) {
		p := &MeCardPayload{}
		if err := p.Validate(); err == nil {
			t.Error("expected error for empty name")
		}
	})
}

func TestEscapeMeCard(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"no special chars", "hello", "hello"},
		{"backslash", `a\b`, `a\\b`},
		{"semicolon", "a;b", `a\;b`},
		{"colon", "a:b", `a\:b`},
		{"multiple specials", "a;b:c\\d", `a\;b\:c\\d`},
		{"only backslash", `\`, `\\`},
		{"only semicolon", `;`, `\;`},
		{"only colon", `:`, `\:`},
		{"empty string", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := escapeMeCard(tt.input)
			if got != tt.want {
				t.Errorf("escapeMeCard(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// =============================================================================
// Event Payload
// =============================================================================

func TestEventPayload(t *testing.T) {
	t.Run("minimal event ID only", func(t *testing.T) {
		p := &EventPayload{EventID: "EVT-001"}
		enc, err := p.Encode()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := "EVENT-TICKET:EVT-001"
		if enc != want {
			t.Errorf("expected %q, got %q", want, enc)
		}
	})

	t.Run("with all fields", func(t *testing.T) {
		start := time.Date(2026, 7, 15, 20, 0, 0, 0, time.UTC)
		p := &EventPayload{
			EventID: "EVT-100", EventName: "Concert", Venue: "Madison Square Garden",
			StartTime: start, Category: "Music", Seat: "A-12",
			Organizer: "LiveNation", Description: "Rock Show",
			URL: "https://example.com/event",
		}
		enc, err := p.Encode()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.HasPrefix(enc, "EVENT-TICKET:EVT-100|") {
			t.Errorf("unexpected event prefix: %s", enc)
		}
		if !containsStr(enc, "Concert") {
			t.Errorf("should contain event name: %s", enc)
		}
		if !containsStr(enc, "Madison Square Garden") {
			t.Errorf("should contain venue: %s", enc)
		}
		if !containsStr(enc, start.UTC().Format(dateTimeLayout)) {
			t.Errorf("should contain formatted start time: %s", enc)
		}
		if !containsStr(enc, "Music") {
			t.Errorf("should contain category: %s", enc)
		}
		if !containsStr(enc, "A-12") {
			t.Errorf("should contain seat: %s", enc)
		}
		if !containsStr(enc, "LiveNation") {
			t.Errorf("should contain organizer: %s", enc)
		}
		if !containsStr(enc, "Rock Show") {
			t.Errorf("should contain description: %s", enc)
		}
		if !containsStr(enc, "https://example.com/event") {
			t.Errorf("should contain URL: %s", enc)
		}
	})

	t.Run("zero start time omitted", func(t *testing.T) {
		p := &EventPayload{EventID: "E1", StartTime: time.Time{}}
		enc, _ := p.Encode()
		if strings.Contains(enc, "|202") {
			t.Errorf("zero start time year should be omitted: %s", enc)
		}
		// Verify no pipe-delimited time field (aside from EVENT-TICKET prefix)
		parts := strings.Split(enc, "|")
		if len(parts) != 1 {
			t.Errorf("zero start time should produce no pipe-delimited fields, got %d parts: %s", len(parts), enc)
		}
	})

	t.Run("partial fields", func(t *testing.T) {
		p := &EventPayload{EventID: "E2", EventName: "Show", Venue: "Hall"}
		enc, _ := p.Encode()
		parts := strings.Split(enc, "|")
		// EVENT-TICKET:E2|Show|Hall
		if len(parts) != 3 {
			t.Errorf("expected 3 pipe-separated parts, got %d: %s", len(parts), enc)
		}
	})

	t.Run("type", func(t *testing.T) {
		p := &EventPayload{EventID: "E1"}
		if p.Type() != "event" {
			t.Errorf("expected %q, got %q", "event", p.Type())
		}
	})

	t.Run("size positive", func(t *testing.T) {
		p := &EventPayload{EventID: "E1"}
		if p.Size() <= 0 {
			t.Errorf("expected positive size, got %d", p.Size())
		}
	})

	t.Run("validate empty event ID", func(t *testing.T) {
		p := &EventPayload{}
		if err := p.Validate(); err == nil {
			t.Error("expected error for empty event ID")
		}
	})

	t.Run("start time formatted with dateTimeLayout", func(t *testing.T) {
		start := time.Date(2026, 12, 25, 14, 30, 45, 0, time.UTC)
		p := &EventPayload{EventID: "E3", StartTime: start}
		enc, _ := p.Encode()
		expectedTime := start.UTC().Format(dateTimeLayout)
		if !containsStr(enc, expectedTime) {
			t.Errorf("expected time %q in %s", expectedTime, enc)
		}
	})
}

// =============================================================================
// WhatsApp Payload
// =============================================================================

func TestWhatsAppPayload(t *testing.T) {
	t.Run("phone only", func(t *testing.T) {
		p := &WhatsAppPayload{Phone: "+1234567890"}
		enc, err := p.Encode()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := "https://wa.me/1234567890"
		if enc != want {
			t.Errorf("expected %q, got %q", want, enc)
		}
	})

	t.Run("with message", func(t *testing.T) {
		p := &WhatsAppPayload{Phone: "+1234567890", Message: "Hello World"}
		enc, err := p.Encode()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.HasPrefix(enc, "https://wa.me/1234567890?text=") {
			t.Errorf("unexpected URL: %s", enc)
		}
	})

	t.Run("phone with formatting stripped", func(t *testing.T) {
		p := &WhatsAppPayload{Phone: "+1 (234) 567-890"}
		enc, err := p.Encode()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(enc, "1234567890") {
			t.Errorf("non-digits should be stripped: %s", enc)
		}
		if strings.Contains(enc, "(") || strings.Contains(enc, ")") || strings.Contains(enc, "-") {
			t.Errorf("formatting chars should be stripped: %s", enc)
		}
	})

	t.Run("type", func(t *testing.T) {
		p := &WhatsAppPayload{Phone: "123"}
		if p.Type() != "whatsapp" {
			t.Errorf("expected %q, got %q", "whatsapp", p.Type())
		}
	})

	t.Run("size positive", func(t *testing.T) {
		p := &WhatsAppPayload{Phone: "123"}
		if p.Size() <= 0 {
			t.Errorf("expected positive size, got %d", p.Size())
		}
	})

	t.Run("validate empty phone", func(t *testing.T) {
		p := &WhatsAppPayload{}
		if err := p.Validate(); err == nil {
			t.Error("expected error for empty phone")
		}
	})

	t.Run("validate phone without digits", func(t *testing.T) {
		p := &WhatsAppPayload{Phone: "abc-xyz"}
		if err := p.Validate(); err == nil {
			t.Error("expected error for phone without digits")
		}
	})
}

func TestCleanPhoneNumber(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"digits only", "1234567890", "1234567890"},
		{"with plus", "+1234567890", "1234567890"},
		{"with dashes", "123-456-7890", "1234567890"},
		{"with parens", "(123) 456-7890", "1234567890"},
		{"with spaces", "+1 234 567 890", "1234567890"},
		{"mixed formatting", "+1 (234) 567-890", "1234567890"},
		{"no digits", "abc-def-ghi", ""},
		{"empty string", "", ""},
		{"single digit", "a1b", "1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cleanPhoneNumber(tt.input)
			if got != tt.want {
				t.Errorf("cleanPhoneNumber(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
