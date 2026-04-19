package payload

import (
	"fmt"
	"net/url"
	"strings"
)

// ZoomPayload encodes a Zoom meeting join link. When scanned, the QR code
// opens the Zoom app or web client and prompts the user to join the specified
// meeting. An optional password and display name can be pre-filled.
//
// Example encoded output:
//
//	https://zoom.us/j/1234567890?pwd=secret&uname=Jane%20Doe
type ZoomPayload struct {
	// MeetingID is the Zoom meeting ID.
	MeetingID string
	// Password is the optional meeting password.
	Password string
	// DisplayName is the optional display name for the joiner.
	DisplayName string
}

// Encode returns a zoom.us/j/ URL with optional password (pwd) and display
// name (uname) query parameters.
func (z *ZoomPayload) Encode() (string, error) {
	if err := z.Validate(); err != nil {
		return "", err
	}
	result := "https://zoom.us/j/" + z.MeetingID
	params := []string{}
	if z.Password != "" {
		params = append(params, "pwd="+url.QueryEscape(z.Password))
	}
	if z.DisplayName != "" {
		params = append(params, "uname="+url.QueryEscape(z.DisplayName))
	}
	if len(params) > 0 {
		result += "?" + strings.Join(params, "&")
	}
	return result, nil
}

// Validate checks that the meeting ID is non-empty.
func (z *ZoomPayload) Validate() error {
	if z.MeetingID == "" {
		return fmt.Errorf("zoom payload: meeting ID must not be empty")
	}
	return nil
}

// Type returns "zoom".
func (z *ZoomPayload) Type() string {
	return "zoom"
}

// Size returns the byte length of the encoded Zoom URL.
func (z *ZoomPayload) Size() int {
	encoded, _ := z.Encode() //nolint:errcheck // Size returns 0 on encode error
	return len(encoded)
}
