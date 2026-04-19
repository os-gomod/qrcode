package payload

import "fmt"

// SMSPayload encodes an SMS message using the smsto: URI scheme.
// This is a widely supported format for QR code SMS initiation. When scanned,
// most smartphones will open the SMS composer with the recipient and optional
// message body pre-filled.
//
// Example encoded output:
//
//	smsto:+14155552671:Hi there!
type SMSPayload struct {
	// Phone is the recipient phone number.
	Phone string
	// Message is the SMS body text.
	Message string
}

// Encode returns an smsto: URI with the phone number and optional message.
// Format: smsto:<phone>[:<message>].
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

// Validate checks that the phone number is non-empty.
func (s *SMSPayload) Validate() error {
	if s.Phone == "" {
		return fmt.Errorf("sms payload: phone number must not be empty")
	}
	return nil
}

// Type returns "sms".
func (s *SMSPayload) Type() string {
	return "sms"
}

// Size returns the byte length of the encoded SMS URI.
func (s *SMSPayload) Size() int {
	encoded, _ := s.Encode() //nolint:errcheck // Size returns 0 on encode error
	return len(encoded)
}
