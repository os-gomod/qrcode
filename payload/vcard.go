package payload

import (
	"errors"
	"fmt"
	"strings"
)

type VCardPayload struct {
	Version      string
	FirstName    string
	LastName     string
	Phone        string
	Email        string
	Organization string
	Title        string
	URL          string
	Address      string
	Note         string
}

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

func (v *VCardPayload) Validate() error {
	if v.FirstName == "" && v.LastName == "" {
		return errors.New("vcard payload: at least FirstName or LastName must be set")
	}
	ver := v.version()
	if ver != "2.1" && ver != "3.0" && ver != "4.0" {
		return fmt.Errorf("vcard payload: unsupported version %q, must be 2.1, 3.0, or 4.0", v.Version)
	}
	return nil
}

func (*VCardPayload) Type() string {
	return "vcard"
}

func (v *VCardPayload) Size() int {
	encoded, _ := v.Encode()
	return len(encoded)
}

func (v *VCardPayload) version() string {
	if v.Version == "" {
		return "3.0"
	}
	return v.Version
}
