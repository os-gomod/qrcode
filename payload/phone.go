package payload

import "fmt"

// PhonePayload encodes a phone number as a tel: URI (RFC 3966).
// When scanned by a smartphone QR reader, the device will typically prompt
// the user to initiate a phone call to the encoded number.
//
// Example encoded output:
//
//	tel:+1-555-0123
type PhonePayload struct {
	// Number is the phone number to encode.
	Number string
}

// Encode returns a tel: URI for the phone number. The number is included
// verbatim (no digit stripping or normalization is performed).
func (p *PhonePayload) Encode() (string, error) {
	if err := p.Validate(); err != nil {
		return "", err
	}
	return "tel:" + p.Number, nil
}

// Validate checks that the number is non-empty and contains at least one digit.
// Formatting characters (spaces, dashes, parentheses, +) are allowed.
func (p *PhonePayload) Validate() error {
	if p.Number == "" {
		return fmt.Errorf("phone payload: number must not be empty")
	}
	if !containsDigit(p.Number) {
		return fmt.Errorf("phone payload: number %q must contain at least one digit", p.Number)
	}
	return nil
}

// Type returns "phone".
func (p *PhonePayload) Type() string {
	return "phone"
}

// Size returns the byte length of the encoded tel URI.
func (p *PhonePayload) Size() int {
	encoded, _ := p.Encode() //nolint:errcheck // Size returns 0 on encode error
	return len(encoded)
}

func containsDigit(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] >= '0' && s[i] <= '9' {
			return true
		}
	}
	return false
}
