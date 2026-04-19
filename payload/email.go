package payload

import (
	"fmt"
	"net/url"
	"strings"
)

// EmailPayload encodes an email message as a mailto: URI per RFC 6068.
// The recipient, subject, body, and CC addresses are encoded as URL parameters.
// Most smartphone QR readers will open the device's email client with these
// fields pre-populated.
//
// Example encoded output:
//
//	mailto:alice@example.com?subject=Hello&body=World&cc=bob@example.com
type EmailPayload struct {
	// To is the recipient email address.
	To string
	// Subject is the email subject line.
	Subject string
	// Body is the email body text.
	Body string
	// CC is a list of carbon-copy recipients.
	CC []string
}

// Encode returns a mailto: URI with optional subject, body, and cc parameters.
// All parameter values are URL-encoded using url.QueryEscape.
func (e *EmailPayload) Encode() (string, error) {
	if err := e.Validate(); err != nil {
		return "", err
	}
	var b strings.Builder
	b.WriteString("mailto:")
	b.WriteString(e.To)
	params := []string{}
	if e.Subject != "" {
		params = append(params, "subject="+url.QueryEscape(e.Subject))
	}
	if e.Body != "" {
		params = append(params, "body="+url.QueryEscape(e.Body))
	}
	for _, cc := range e.CC {
		if cc != "" {
			params = append(params, "cc="+url.QueryEscape(cc))
		}
	}
	if len(params) > 0 {
		b.WriteString("?")
		b.WriteString(strings.Join(params, "&"))
	}
	return b.String(), nil
}

// Validate checks that the recipient (To) address is non-empty.
func (e *EmailPayload) Validate() error {
	if e.To == "" {
		return fmt.Errorf("email payload: recipient (To) must not be empty")
	}
	return nil
}

// Type returns "email".
func (e *EmailPayload) Type() string {
	return "email"
}

// Size returns the byte length of the encoded mailto URI.
func (e *EmailPayload) Size() int {
	encoded, _ := e.Encode() //nolint:errcheck // Size returns 0 on encode error
	return len(encoded)
}
