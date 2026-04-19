package payload

import (
	"fmt"
	"net/url"
	"strings"
)

// URLPayload encodes a URL into a QR code.
// The URL is normalized to HTTPS if no scheme is provided. An optional Title
// field is appended as a URL fragment (#title) to provide a human-readable
// label in some QR reader apps.
//
// Supported schemes are http and https. If the input has no scheme, https is
// automatically prepended.
type URLPayload struct {
	// URL is the target URL to encode.
	URL string
	// Title is an optional fragment title appended to the URL.
	Title string
}

// Encode returns a normalized HTTPS URL, optionally with a title fragment.
// If the URL has no scheme, https:// is prepended. If Title is set, it is
// URL-encoded and appended as a fragment (#title).
func (u *URLPayload) Encode() (string, error) {
	if err := u.Validate(); err != nil {
		return "", err
	}
	normalized := normalizeURL(u.URL)
	if u.Title != "" {
		normalized += "#" + url.QueryEscape(u.Title)
	}
	return normalized, nil
}

// Validate checks that the URL is non-empty and has a supported scheme.
// Only http and https schemes are accepted.
func (u *URLPayload) Validate() error {
	if u.URL == "" {
		return fmt.Errorf("url payload: URL must not be empty")
	}
	parsed, err := url.Parse(u.URL)
	if err != nil {
		return fmt.Errorf("url payload: invalid URL %q: %w", u.URL, err)
	}
	if parsed.Scheme != "" && parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("url payload: unsupported scheme %q, must be http or https", parsed.Scheme)
	}
	return nil
}

// Type returns "url".
func (u *URLPayload) Type() string {
	return "url"
}

// Size returns the byte length of the encoded URL.
func (u *URLPayload) Size() int {
	encoded, _ := u.Encode() //nolint:errcheck // Size returns 0 on encode error
	return len(encoded)
}

func normalizeURL(rawURL string) string {
	if strings.HasPrefix(rawURL, "http://") || strings.HasPrefix(rawURL, "https://") {
		return rawURL
	}
	if strings.HasPrefix(rawURL, "//") {
		return "https://" + rawURL
	}
	return "https://" + rawURL
}
