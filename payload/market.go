package payload

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

const (
	MarketGooglePlay = "google"
	MarketAppleApp   = "apple"
)

type MarketPayload struct {
	Platform  string
	PackageID string
	AppName   string
	Campaign  string
}

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
			storeURL = "https://apps.apple.com/app/id" + m.PackageID
		} else if m.AppName != "" {
			storeURL = "https://apps.apple.com/search?term=" + url.QueryEscape(m.AppName)
		}
	default:
		storeURL = "https://play.google.com/store/apps/details?id=" + url.QueryEscape(m.PackageID)
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

func (m *MarketPayload) Validate() error {
	if m.PackageID == "" && m.AppName == "" {
		return errors.New("market payload: at least PackageID or AppName must be set")
	}
	if m.Platform != "" && m.Platform != MarketGooglePlay && m.Platform != MarketAppleApp {
		return fmt.Errorf("market payload: unsupported platform %q, must be %q or %q", m.Platform, MarketGooglePlay, MarketAppleApp)
	}
	return nil
}

func (*MarketPayload) Type() string {
	return "market"
}

func (m *MarketPayload) Size() int {
	encoded, _ := m.Encode()
	return len(encoded)
}
