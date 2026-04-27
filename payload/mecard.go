package payload

import (
	"errors"
	"strings"
)

type MeCardPayload struct {
	Name     string
	Phone    string
	Email    string
	URL      string
	Birthday string
	Note     string
	Address  string
	Nickname string
}

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

func (m *MeCardPayload) Validate() error {
	if m.Name == "" {
		return errors.New("mecard payload: name must not be empty")
	}
	return nil
}

func (*MeCardPayload) Type() string {
	return "mecard"
}

func (m *MeCardPayload) Size() int {
	encoded, _ := m.Encode()
	return len(encoded)
}

func escapeMeCard(s string) string {
	var b strings.Builder
	for i := range len(s) {
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
