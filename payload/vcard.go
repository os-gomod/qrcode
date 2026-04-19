package payload

import (
	"fmt"
	"strings"
)

// VCardPayload encodes a contact as a vCard electronic business card.
// By default, version 3.0 is used, which offers broad compatibility across
// QR readers, contact apps, and email clients. Supported versions are:
//   - "2.1" — legacy vCard format
//   - "3.0" — RFC 6350 predecessor (default)
//   - "4.0" — RFC 6350 current standard
//
// The encoded output uses CRLF line endings and includes BEGIN:VCARD /
// END:VCARD delimiters. Optional fields (phone, email, organization, etc.)
// are included only when non-empty.
//
// Example encoded output:
//
//	BEGIN:VCARD\r\nVERSION:3.0\r\nN:Doe;Jane\r\nFN:Jane Doe\r\nEND:VCARD
type VCardPayload struct {
	// Version is the vCard version (defaults to "3.0").
	Version string
	// FirstName is the given name.
	FirstName string
	// LastName is the family name.
	LastName string
	// Phone is the telephone number.
	Phone string
	// Email is the email address.
	Email string
	// Organization is the company or organization name.
	Organization string
	// Title is the job title.
	Title string
	// URL is the website address.
	URL string
	// Address is the postal address.
	Address string
	// Note is a free-text note.
	Note string
}

// Encode returns a vCard formatted string with CRLF line endings.
// Only non-empty optional fields (TEL, EMAIL, ORG, TITLE, URL, ADR, NOTE)
// are included in the output.
func (v *VCardPayload) Encode() (string, error) {
	if err := v.Validate(); err != nil {
		return "", err
	}
	var b strings.Builder
	b.WriteString("BEGIN:VCARD\r\n")
	fmt.Fprintf(&b, "VERSION:%s\r\n", v.version())
	lastName := v.LastName
	fmt.Fprintf(&b, "N:%s;%s\r\n", lastName, v.FirstName)
	fn := strings.TrimSpace(v.FirstName + " " + v.LastName)
	if fn != "" {
		fmt.Fprintf(&b, "FN:%s\r\n", fn)
	}
	if v.Phone != "" {
		fmt.Fprintf(&b, "TEL:%s\r\n", v.Phone)
	}
	if v.Email != "" {
		fmt.Fprintf(&b, "EMAIL:%s\r\n", v.Email)
	}
	if v.Organization != "" {
		fmt.Fprintf(&b, "ORG:%s\r\n", v.Organization)
	}
	if v.Title != "" {
		fmt.Fprintf(&b, "TITLE:%s\r\n", v.Title)
	}
	if v.URL != "" {
		fmt.Fprintf(&b, "URL:%s\r\n", v.URL)
	}
	if v.Address != "" {
		fmt.Fprintf(&b, "ADR:;;%s;;;;\r\n", v.Address)
	}
	if v.Note != "" {
		fmt.Fprintf(&b, "NOTE:%s\r\n", v.Note)
	}
	b.WriteString("END:VCARD")
	return b.String(), nil
}

// Validate checks that at least FirstName or LastName is provided and that
// the version (if set) is one of "2.1", "3.0", or "4.0".
func (v *VCardPayload) Validate() error {
	if v.FirstName == "" && v.LastName == "" {
		return fmt.Errorf("vcard payload: at least FirstName or LastName must be set")
	}
	ver := v.version()
	if ver != "2.1" && ver != "3.0" && ver != "4.0" {
		return fmt.Errorf("vcard payload: unsupported version %q, must be 2.1, 3.0, or 4.0", v.Version)
	}
	return nil
}

// Type returns "vcard".
func (v *VCardPayload) Type() string {
	return "vcard"
}

// Size returns the byte length of the encoded vCard.
func (v *VCardPayload) Size() int {
	encoded, _ := v.Encode() //nolint:errcheck // Size returns 0 on encode error
	return len(encoded)
}

func (v *VCardPayload) version() string {
	if v.Version == "" {
		return "3.0"
	}
	return v.Version
}
