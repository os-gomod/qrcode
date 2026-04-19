package payload

import (
	"fmt"
	"strings"
)

// EncryptionWEP represents WEP encryption.
const EncryptionWEP = "WEP"

// EncryptionWPA represents WPA encryption.
const EncryptionWPA = "WPA"

// EncryptionWPA2 represents WPA2 encryption.
const EncryptionWPA2 = "WPA2"

// EncryptionWPA3 represents WPA3 encryption.
const EncryptionWPA3 = "WPA3"

// EncryptionSAE represents SAE (Simultaneous Authentication of Equals) encryption.
const EncryptionSAE = "SAE"

// EncryptionNoPass indicates an open network with no password.
const EncryptionNoPass = "nopass"

// WiFiPayload encodes a WiFi network configuration into a QR code using the
// standard WIFI:T:...;S:...;P:...;; format. This format is recognized by
// Android (since Android 10), iOS (since iOS 11), and many other QR readers.
//
// Special characters in the SSID and password (\, ;, ,, ", :) are escaped
// using backslash-hex notation (\XX) per the barcode specification.
//
// The following encryption types are supported as constants:
//   - EncryptionWEP    ("WEP")
//   - EncryptionWPA    ("WPA")
//   - EncryptionWPA2   ("WPA2")
//   - EncryptionWPA3   ("WPA3")
//   - EncryptionSAE    ("SAE")
//   - EncryptionNoPass ("nopass") — for open networks
//
// Example encoded output:
//
//	WIFI:T:WPA2;S:MyNetwork;P:s\3Acret;;
type WiFiPayload struct {
	// SSID is the network name.
	SSID string
	// Password is the network passphrase.
	Password string
	// Encryption is the encryption type (WEP, WPA, WPA2, WPA3, SAE, nopass).
	Encryption string
	// Hidden indicates whether the SSID is hidden.
	Hidden bool
}

// Encode returns the WIFI:... string representation of the network configuration.
// Special characters in SSID and Password are hex-escaped. The password field
// is omitted when encryption is "nopass".
func (w *WiFiPayload) Encode() (string, error) {
	if err := w.Validate(); err != nil {
		return "", err
	}
	var b strings.Builder
	b.WriteString("WIFI:T:")
	b.WriteString(w.Encryption)
	b.WriteString(";S:")
	b.WriteString(escapeWiFi(w.SSID))
	if w.Encryption != EncryptionNoPass {
		b.WriteString(";P:")
		b.WriteString(escapeWiFi(w.Password))
	}
	if w.Hidden {
		b.WriteString(";H:true")
	}
	b.WriteString(";;")
	return b.String(), nil
}

// Validate checks that the SSID is non-empty, the encryption type is one of the
// supported values, and a password is provided when encryption is not "nopass".
func (w *WiFiPayload) Validate() error {
	if w.SSID == "" {
		return fmt.Errorf("wifi payload: SSID must not be empty")
	}
	if !isValidEncryption(w.Encryption) {
		return fmt.Errorf("wifi payload: invalid encryption type %q, must be one of WEP, WPA, WPA2, WPA3, SAE, nopass", w.Encryption)
	}
	if w.Encryption != EncryptionNoPass && w.Password == "" {
		return fmt.Errorf("wifi payload: password is required for %s encryption", w.Encryption)
	}
	return nil
}

// Type returns "wifi".
func (w *WiFiPayload) Type() string {
	return "wifi"
}

// Size returns the byte length of the encoded WiFi string.
func (w *WiFiPayload) Size() int {
	encoded, _ := w.Encode() //nolint:errcheck // Size returns 0 on encode error
	return len(encoded)
}

func isValidEncryption(enc string) bool {
	switch enc {
	case EncryptionWEP, EncryptionWPA, EncryptionWPA2, EncryptionWPA3, EncryptionSAE, EncryptionNoPass:
		return true
	default:
		return false
	}
}

func escapeWiFi(s string) string {
	special := map[byte]bool{
		'\\': true,
		';':  true,
		',':  true,
		'"':  true,
		':':  true,
	}
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		if special[c] {
			fmt.Fprintf(&b, "\\%02X", c)
		} else {
			b.WriteByte(c)
		}
	}
	return b.String()
}
