package payload

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

type NTPLocalePayload struct {
	Host        string
	Port        string
	Version     int
	Description string
}

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

func (n *NTPLocalePayload) Validate() error {
	if n.Host == "" {
		return errors.New("ntp payload: host must not be empty")
	}
	if n.Port != "" {
		port := 0
		for i := range len(n.Port) {
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

func (*NTPLocalePayload) Type() string {
	return "ntp"
}

func (n *NTPLocalePayload) Size() int {
	encoded, _ := n.Encode()
	return len(encoded)
}

func (n *NTPLocalePayload) String() string {
	var b strings.Builder
	b.WriteString("NTP://" + n.Host)
	if n.Port != "" && n.Port != "123" {
		b.WriteString(":")
		b.WriteString(n.Port)
	}
	if n.Version != 0 {
		fmt.Fprintf(&b, " (v%d)", n.Version)
	}
	return b.String()
}
