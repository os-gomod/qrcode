package payload

import (
	"fmt"
	"strings"
)

// MeCardPayload encodes a contact in the MeCard format, a lightweight
// alternative to vCard originally developed by NTT DoCoMo for mobile phones.
// MeCard is widely supported by Japanese QR readers and many international
// scanner apps. It uses a compact semicolon-delimited format.
//
// Special characters (\, ;, :) in field values are backslash-escaped.
//
// Example encoded output:
//
//	MECARD:N:Jane Doe;TEL:+1-555-0123;EMAIL:jane@example.com;;
type MeCardPayload struct {
	// Name is the contact's full name.
	Name string
	// Phone is the phone number.
	Phone string
	// Email is the email address.
	Email string
	// URL is the website address.
	URL string
	// Birthday is the birthday in YYYYMMDD format.
	Birthday string
	// Note is a free-text note.
	Note string
	// Address is the postal address.
	Address string
	// Nickname is the contact's nickname.
	Nickname string
}

// Encode returns a MECARD: formatted string with semicolon-delimited fields.
// Special characters in values are backslash-escaped.
func (m *MeCardPayload) Encode() (string, error) {
	if err := m.Validate(); err != nil {
		return "", err
	}
	var b strings.Builder
	b.WriteString("MECARD:")
	b.WriteString("N:")
	b.WriteString(escapeMeCard(m.Name))
	b.WriteString(";")
	if m.Phone != "" {
		b.WriteString("TEL:")
		b.WriteString(escapeMeCard(m.Phone))
		b.WriteString(";")
	}
	if m.Email != "" {
		b.WriteString("EMAIL:")
		b.WriteString(escapeMeCard(m.Email))
		b.WriteString(";")
	}
	if m.URL != "" {
		b.WriteString("URL:")
		b.WriteString(escapeMeCard(m.URL))
		b.WriteString(";")
	}
	if m.Birthday != "" {
		b.WriteString("BDAY:")
		b.WriteString(escapeMeCard(m.Birthday))
		b.WriteString(";")
	}
	if m.Note != "" {
		b.WriteString("NOTE:")
		b.WriteString(escapeMeCard(m.Note))
		b.WriteString(";")
	}
	if m.Address != "" {
		b.WriteString("ADR:")
		b.WriteString(escapeMeCard(m.Address))
		b.WriteString(";")
	}
	if m.Nickname != "" {
		b.WriteString("NICKNAME:")
		b.WriteString(escapeMeCard(m.Nickname))
		b.WriteString(";")
	}
	b.WriteString(";")
	return b.String(), nil
}

// Validate checks that the contact name is non-empty.
func (m *MeCardPayload) Validate() error {
	if m.Name == "" {
		return fmt.Errorf("mecard payload: name must not be empty")
	}
	return nil
}

// Type returns "mecard".
func (m *MeCardPayload) Type() string {
	return "mecard"
}

// Size returns the byte length of the encoded MeCard string.
func (m *MeCardPayload) Size() int {
	encoded, _ := m.Encode() //nolint:errcheck // Size returns 0 on encode error
	return len(encoded)
}

func escapeMeCard(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch c {
		case '\\', ';', ':':
			b.WriteByte('\\')
			b.WriteByte(c)
		default:
			b.WriteByte(c)
		}
	}
	return b.String()
}
