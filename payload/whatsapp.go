package payload

import (
	"errors"
	"net/url"
	"strings"
)

type WhatsAppPayload struct {
	Phone   string
	Message string
}

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

func (w *WhatsAppPayload) Validate() error {
	if w.Phone == "" {
		return errors.New("whatsapp payload: phone number must not be empty")
	}
	cleaned := cleanPhoneNumber(w.Phone)
	if cleaned == "" {
		return errors.New("whatsapp payload: phone number must contain at least one digit")
	}
	return nil
}

func (*WhatsAppPayload) Type() string {
	return "whatsapp"
}

func (w *WhatsAppPayload) Size() int {
	encoded, _ := w.Encode()
	return len(encoded)
}

func cleanPhoneNumber(s string) string {
	var b strings.Builder
	for i := range len(s) {
		c := s[i]
		if c >= '0' && c <= '9' {
			b.WriteByte(c)
		}
	}
	return b.String()
}
