package payload

import (
	"errors"
	"fmt"
)

type PhonePayload struct {
	Number string
}

func (p *PhonePayload) Encode() (string, error) {
	if err := p.Validate(); err != nil {
		return "", err
	}
	return "tel:" + p.Number, nil
}

func (p *PhonePayload) Validate() error {
	if p.Number == "" {
		return errors.New("phone payload: number must not be empty")
	}
	if !containsDigit(p.Number) {
		return fmt.Errorf("phone payload: number %q must contain at least one digit", p.Number)
	}
	return nil
}

func (*PhonePayload) Type() string {
	return "phone"
}

func (p *PhonePayload) Size() int {
	encoded, _ := p.Encode()
	return len(encoded)
}

func containsDigit(s string) bool {
	for i := range len(s) {
		if s[i] >= '0' && s[i] <= '9' {
			return true
		}
	}
	return false
}
