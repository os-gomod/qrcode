package payload

import (
	"fmt"
	"strings"
)

// IBeaconPayload encodes an iBeacon Bluetooth Low Energy (BLE) beacon
// configuration as a URL following the open beacon registry format
// (https://github.com/google/beacon).
// The URL points to <manufacturer>.github.io/beacon/ with UUID, major,
// and minor as query parameters. When scanned, the beacon data can be
// extracted and used by beacon-aware applications.
//
// The UUID must be in the standard format: XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX.
// Major and Minor are 16-bit unsigned integers used for proximity identification.
//
// Example encoded output:
//
//	https://beacon.github.io/beacon/?uuid=A1B2C3D4-E5F6-7890-ABCD-EF1234567890&major=1&minor=42
type IBeaconPayload struct {
	// UUID is the iBeacon proximity UUID.
	UUID string
	// Major is the iBeacon major value.
	Major uint16
	// Minor is the iBeacon minor value.
	Minor uint16
	// Manufacturer is the optional manufacturer identifier (defaults to "beacon").
	Manufacturer string
}

// Encode returns an iBeacon URL following the beacon registry format.
// Format: https://<manufacturer>.github.io/beacon/?uuid=<UUID>&major=<N>&minor=<N>
// The manufacturer defaults to "beacon" if not specified.
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

// Validate checks that the UUID is non-empty and matches the format
// XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX (32 hex digits with hyphens).
func (ib *IBeaconPayload) Validate() error {
	if ib.UUID == "" {
		return fmt.Errorf("ibeacon payload: UUID must not be empty")
	}
	if !isValidUUID(ib.UUID) {
		return fmt.Errorf("ibeacon payload: invalid UUID format %q, expected format XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX", ib.UUID)
	}
	return nil
}

// Type returns "ibeacon".
func (ib *IBeaconPayload) Type() string {
	return "ibeacon"
}

// Size returns the byte length of the encoded iBeacon URL.
func (ib *IBeaconPayload) Size() int {
	encoded, _ := ib.Encode() //nolint:errcheck // Size returns 0 on encode error
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
	for i := 0; i < len(clean); i++ {
		c := clean[i]
		if !isHexDigit(c) {
			return false
		}
	}
	return true
}
