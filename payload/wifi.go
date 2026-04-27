package payload

import (
	"errors"
	"fmt"
	"strings"
)

const (
	EncryptionWEP    = "WEP"
	EncryptionWPA    = "WPA"
	EncryptionWPA2   = "WPA2"
	EncryptionWPA3   = "WPA3"
	EncryptionSAE    = "SAE"
	EncryptionNoPass = "nopass"
)

type WiFiPayload struct {
	SSID       string
	Password   string
	Encryption string
	Hidden     bool
}

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

func (w *WiFiPayload) Validate() error {
	if w.SSID == "" {
		return errors.New("wifi payload: SSID must not be empty")
	}
	if !isValidEncryption(w.Encryption) {
		return fmt.Errorf("wifi payload: invalid encryption type %q, must be one of WEP, WPA, WPA2, WPA3, SAE, nopass", w.Encryption)
	}
	if w.Encryption != EncryptionNoPass && w.Password == "" {
		return fmt.Errorf("wifi payload: password is required for %s encryption", w.Encryption)
	}
	return nil
}

func (*WiFiPayload) Type() string {
	return "wifi"
}

func (w *WiFiPayload) Size() int {
	encoded, _ := w.Encode()
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
	for i := range len(s) {
		c := s[i]
		if special[c] {
			fmt.Fprintf(&b, "\\%02X", c)
		} else {
			b.WriteByte(c)
		}
	}
	return b.String()
}
