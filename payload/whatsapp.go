package payload

import (
	"fmt"
	"net/url"
	"strings"
)

// WhatsAppPayload encodes a WhatsApp chat link using the wa.me deep link API.
// The phone number is cleaned to digits only before being appended to the URL.
// An optional pre-filled message is added as a ?text= query parameter.
// When scanned, the QR code opens a WhatsApp chat with the specified number.
//
// Example encoded output:
//
//	https://wa.me/14155552671?text=Hello%20there
type WhatsAppPayload struct {
	// Phone is the phone number in international format.
	Phone string
	// Message is an optional pre-filled message.
	Message string
}

// Encode returns a wa.me link with an optional text parameter.
// The phone number is stripped of non-digit characters. The message, if
// present, is URL-encoded and appended as ?text=...
func (w *WhatsAppPayload) Encode() (string, error) {
	if err := w.Validate(); err != nil {
		return "", err
	}
	cleaned := cleanPhoneNumber(w.Phone)
	result := "https://wa.me/" + cleaned
	if w.Message != "" {
		result += "?text=" + url.QueryEscape(w.Message)
	}
	return result, nil
}

// Validate checks that the phone number is non-empty and contains at least
// one digit after cleaning (non-digit characters are stripped).
func (w *WhatsAppPayload) Validate() error {
	if w.Phone == "" {
		return fmt.Errorf("whatsapp payload: phone number must not be empty")
	}
	cleaned := cleanPhoneNumber(w.Phone)
	if cleaned == "" {
		return fmt.Errorf("whatsapp payload: phone number must contain at least one digit")
	}
	return nil
}

// Type returns "whatsapp".
func (w *WhatsAppPayload) Type() string {
	return "whatsapp"
}

// Size returns the byte length of the encoded wa.me link.
func (w *WhatsAppPayload) Size() int {
	encoded, _ := w.Encode() //nolint:errcheck // Size returns 0 on encode error
	return len(encoded)
}

func cleanPhoneNumber(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= '0' && c <= '9' {
			b.WriteByte(c)
		}
	}
	return b.String()
}
