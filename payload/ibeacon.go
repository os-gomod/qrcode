package payload

import (
	"errors"
	"fmt"
	"strings"
)

type IBeaconPayload struct {
	UUID         string
	Major        uint16
	Minor        uint16
	Manufacturer string
}

func (ib *IBeaconPayload) Encode() (string, error) {
	if err := ib.Validate(); err != nil {
		return "", err
	}
	mfr := ib.Manufacturer
	if mfr == "" {
		mfr = "beacon"
	}
	var b strings.Builder
	b.WriteString("https://")
	b.WriteString(strings.ToLower(mfr))
	b.WriteString(".github.io/beacon/")
	b.WriteString("?uuid=")
	b.WriteString(strings.ToUpper(ib.UUID))
	b.WriteString("&major=")
	fmt.Fprintf(&b, "%d", ib.Major)
	b.WriteString("&minor=")
	fmt.Fprintf(&b, "%d", ib.Minor)
	return b.String(), nil
}

func (ib *IBeaconPayload) Validate() error {
	if ib.UUID == "" {
		return errors.New("ibeacon payload: UUID must not be empty")
	}
	if !isValidUUID(ib.UUID) {
		return fmt.Errorf("ibeacon payload: invalid UUID format %q, expected format XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX", ib.UUID)
	}
	return nil
}

func (*IBeaconPayload) Type() string {
	return "ibeacon"
}

func (ib *IBeaconPayload) Size() int {
	encoded, _ := ib.Encode()
	return len(encoded)
}

func isHexDigit(c byte) bool {
	return (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')
}

func isValidUUID(s string) bool {
	clean := strings.ReplaceAll(s, "-", "")
	if len(clean) != 32 {
		return false
	}
	for i := range len(clean) {
		c := clean[i]
		if !isHexDigit(c) {
			return false
		}
	}
	return true
}
