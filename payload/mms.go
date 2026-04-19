package payload

import (
	"fmt"
	"strings"
)

// MMSPayload encodes an MMS message using the mms: URI scheme.
// The phone number and optional subject/body are encoded as a URI with
// query parameters. This format is primarily supported by Android devices.
//
// Example encoded output:
//
//	mms:+14155552671?subject=Check%20this&body=Hello!
type MMSPayload struct {
	// Phone is the recipient phone number.
	Phone string
	// Message is the MMS body text.
	Message string
	// Subject is the MMS subject line.
	Subject string
}

// Encode returns an mms: URI with optional subject and body parameters.
// Format: mms:<phone>[?subject=...&body=...].
func (m *MMSPayload) Encode() (string, error) {
	if err := m.Validate(); err != nil {
		return "", err
	}
	result := "mms:" + m.Phone
	params := []string{}
	if m.Subject != "" {
		params = append(params, "subject="+m.Subject)
	}
	if m.Message != "" {
		params = append(params, "body="+m.Message)
	}
	if len(params) > 0 {
		result += "?" + strings.Join(params, "&")
	}
	return result, nil
}

// Validate checks that the phone number is non-empty and contains at least
// one digit (formatting characters are allowed).
func (m *MMSPayload) Validate() error {
	if m.Phone == "" {
		return fmt.Errorf("mms payload: phone number must not be empty")
	}
	if !containsDigit(m.Phone) {
		return fmt.Errorf("mms payload: phone number %q must contain at least one digit", m.Phone)
	}
	return nil
}

// Type returns "mms".
func (m *MMSPayload) Type() string {
	return "mms"
}

// Size returns the byte length of the encoded MMS URI.
func (m *MMSPayload) Size() int {
	encoded, _ := m.Encode() //nolint:errcheck // Size returns 0 on encode error
	return len(encoded)
}
