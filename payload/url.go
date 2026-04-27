package payload

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

type URLPayload struct {
	URL   string
	Title string
}

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

func (u *URLPayload) Validate() error {
	if u.URL == "" {
		return errors.New("url payload: URL must not be empty")
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

func (*URLPayload) Type() string {
	return "url"
}

func (u *URLPayload) Size() int {
	encoded, _ := u.Encode()
	return len(encoded)
}

func normalizeURL(rawURL string) string {
	if strings.HasPrefix(rawURL, "http://") || strings.HasPrefix(rawURL, "https://") {
		return rawURL
	}
	if strings.HasPrefix(rawURL, "//") {
		return "https:" + rawURL
	}
	return "https://" + rawURL
}
