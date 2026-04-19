package payload

import (
	"fmt"
	"net/url"
	"strings"
)

// NTPLocalePayload encodes an NTP time server configuration as an ntp:// URI.
// This is used in QR codes for configuring network time synchronization on
// devices. The default NTP port (123) is omitted from the URI.
//
// Example encoded output:
//
//	ntp://time.google.com#NIST%20Time%20Server
type NTPLocalePayload struct {
	// Host is the NTP server hostname or IP address.
	Host string
	// Port is the server port (defaults to 123).
	Port string
	// Version is the NTP protocol version (3 or 4).
	Version int
	// Description is an optional human-readable description.
	Description string
}

// Encode returns an ntp:// URI for the time server. The default port 123
// is omitted. An optional description is appended as a URL fragment.
func (n *NTPLocalePayload) Encode() (string, error) {
	if err := n.Validate(); err != nil {
		return "", err
	}
	encoded := "ntp://" + n.Host
	if n.Port != "" && n.Port != "123" {
		encoded += ":" + url.PathEscape(n.Port)
	}
	if n.Description != "" {
		encoded += "#" + url.QueryEscape(n.Description)
	}
	return encoded, nil
}

// Validate checks that the host is non-empty, the port (if set) is a valid
// number in [1, 65535], and the version (if set) is 3 or 4.
func (n *NTPLocalePayload) Validate() error {
	if n.Host == "" {
		return fmt.Errorf("ntp payload: host must not be empty")
	}
	if n.Port != "" {
		port := 0
		for i := 0; i < len(n.Port); i++ {
			c := n.Port[i]
			if c < '0' || c > '9' {
				return fmt.Errorf("ntp payload: port %q must be numeric", n.Port)
			}
			port = port*10 + int(c-'0')
		}
		if port < 1 || port > 65535 {
			return fmt.Errorf("ntp payload: port %d is out of range [1, 65535]", port)
		}
	}
	if n.Version != 0 && n.Version != 3 && n.Version != 4 {
		return fmt.Errorf("ntp payload: unsupported version %d, must be 3 or 4", n.Version)
	}
	return nil
}

// Type returns "ntp".
func (n *NTPLocalePayload) Type() string {
	return "ntp"
}

// Size returns the byte length of the encoded ntp URI.
func (n *NTPLocalePayload) Size() int {
	encoded, _ := n.Encode() //nolint:errcheck // Size returns 0 on encode error
	return len(encoded)
}

// String returns a human-readable representation of the NTP configuration
// including the protocol version if set.
func (n *NTPLocalePayload) String() string {
	var b strings.Builder
	b.WriteString("NTP://")
	b.WriteString(n.Host)
	if n.Port != "" && n.Port != "123" {
		b.WriteString(":")
		b.WriteString(n.Port)
	}
	if n.Version != 0 {
		fmt.Fprintf(&b, " (v%d)", n.Version)
	}
	return b.String()
}
