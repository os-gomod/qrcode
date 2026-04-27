package payload

import (
	"errors"
)

type SMSPayload struct {
	Phone   string
	Message string
}

func (s *SMSPayload) Encode() (string, error) {
	if err := s.Validate(); err != nil {
		return "", err
	}
	result := "smsto:" + s.Phone
	if s.Message != "" {
		result += ":" + s.Message
	}
	return result, nil
}

func (s *SMSPayload) Validate() error {
	if s.Phone == "" {
		return errors.New("sms payload: phone number must not be empty")
	}
	return nil
}

func (*SMSPayload) Type() string {
	return "sms"
}

func (s *SMSPayload) Size() int {
	encoded, _ := s.Encode()
	return len(encoded)
}
