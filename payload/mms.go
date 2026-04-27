package payload

import (
	"errors"
	"fmt"
	"strings"
)

type MMSPayload struct {
	Phone   string
	Message string
	Subject string
}

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

func (m *MMSPayload) Validate() error {
	if m.Phone == "" {
		return errors.New("mms payload: phone number must not be empty")
	}
	if !containsDigit(m.Phone) {
		return fmt.Errorf("mms payload: phone number %q must contain at least one digit", m.Phone)
	}
	return nil
}

func (*MMSPayload) Type() string {
	return "mms"
}

func (m *MMSPayload) Size() int {
	encoded, _ := m.Encode()
	return len(encoded)
}
