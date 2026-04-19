package payload

import (
	"fmt"
	"net/url"
	"strings"
)

// MarketGooglePlay represents the Google Play Store.
// Used as the Platform field in MarketPayload.
const MarketGooglePlay = "google"

// MarketAppleApp represents the Apple App Store.
// Used as the Platform field in MarketPayload.
const MarketAppleApp = "apple"

// MarketPayload encodes an app store download link for either Google Play or
// Apple App Store. If PackageID is set, a direct app page URL is produced.
// If only AppName is set, a store search URL is used as a fallback.
//
// An optional Campaign field appends UTM tracking parameters
// (utm_source=qr&utm_medium=scan&utm_campaign=...) to the URL.
//
// Example encoded output (Google Play, package ID):
//
//	https://play.google.com/store/apps/details?id=com.example.app
//
// Example encoded output (Apple, with campaign):
//
//	https://apps.apple.com/app/1234567890?utm_source=qr&utm_medium=scan&utm_campaign=launch
type MarketPayload struct {
	// Platform is the store platform ("google" or "apple").
	Platform string
	// PackageID is the app's package ID or App Store ID.
	PackageID string
	// AppName is the app name (used as a search fallback).
	AppName string
	// Campaign is an optional UTM campaign parameter.
	Campaign string
}

// Encode returns a store URL for the specified platform and app.
// For Google Play: play.google.com/store/apps/details?id=... or /search?q=...
// For Apple: apps.apple.com/app/... or /search?term=...
// UTM campaign parameters are appended if Campaign is set.
func (m *MarketPayload) Encode() (string, error) {
	if err := m.Validate(); err != nil {
		return "", err
	}
	var storeURL string
	switch m.Platform {
	case MarketGooglePlay:
		if m.PackageID != "" {
			storeURL = "https://play.google.com/store/apps/details?id=" + url.QueryEscape(m.PackageID)
		} else if m.AppName != "" {
			storeURL = "https://play.google.com/store/search?q=" + url.QueryEscape(m.AppName)
		}
	case MarketAppleApp:
		if m.PackageID != "" {
			storeURL = "https://apps.apple.com/app/" + url.PathEscape(m.PackageID)
		} else if m.AppName != "" {
			storeURL = "https://apps.apple.com/search?term=" + url.QueryEscape(m.AppName)
		}
	default:
		storeURL = "https://play.google.com/store/search?q=" + url.QueryEscape(m.AppName)
	}
	if m.Campaign != "" {
		sep := "?"
		if strings.Contains(storeURL, "?") {
			sep = "&"
		}
		storeURL += sep + "utm_source=qr&utm_medium=scan&utm_campaign=" + url.QueryEscape(m.Campaign)
	}
	return storeURL, nil
}

// Validate checks that at least PackageID or AppName is set, and that
// Platform (if set) is either "google" or "apple".
func (m *MarketPayload) Validate() error {
	if m.PackageID == "" && m.AppName == "" {
		return fmt.Errorf("market payload: at least PackageID or AppName must be set")
	}
	if m.Platform != "" && m.Platform != MarketGooglePlay && m.Platform != MarketAppleApp {
		return fmt.Errorf("market payload: unsupported platform %q, must be %q or %q", m.Platform, MarketGooglePlay, MarketAppleApp)
	}
	return nil
}

// Type returns "market".
func (m *MarketPayload) Type() string {
	return "market"
}

// Size returns the byte length of the encoded store URL.
func (m *MarketPayload) Size() int {
	encoded, _ := m.Encode() //nolint:errcheck // Size returns 0 on encode error
	return len(encoded)
}
