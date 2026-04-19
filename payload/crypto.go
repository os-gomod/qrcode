package payload

import (
	"fmt"
	"strings"
)

// CryptoBTC represents Bitcoin.
// The encoded URI scheme is "bitcoin:".
// Used as the CryptoType field in CryptoPayload.
const CryptoBTC = "BTC"

// CryptoETH represents Ethereum.
// The encoded URI scheme is "ethereum:".
// Used as the CryptoType field in CryptoPayload.
const CryptoETH = "ETH"

// CryptoLTC represents Litecoin.
// The encoded URI scheme is "litecoin:".
// Used as the CryptoType field in CryptoPayload.
const CryptoLTC = "LTC"

// CryptoPayload encodes a cryptocurrency payment request as a URI following
// the BIP21 Bitcoin URI scheme (adapted for ETH and LTC).
// When scanned by a crypto wallet app, the user is prompted to send the
// specified amount to the given address.
//
// Example encoded output:
//
//	bitcoin:1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa?amount=0.01&label=Tip&message=Thanks
type CryptoPayload struct {
	// Address is the wallet address.
	Address string
	// Amount is the optional payment amount.
	Amount string
	// Label is an optional label for the transaction.
	Label string
	// Message is an optional message for the transaction.
	Message string
	// CryptoType is the cryptocurrency type (BTC, ETH, or LTC).
	CryptoType string
}

// Encode returns a cryptocurrency URI with optional amount, label, and
// message query parameters. Format: <scheme>:<address>[?amount=...&label=...&message=...].
func (c *CryptoPayload) Encode() (string, error) {
	if err := c.Validate(); err != nil {
		return "", err
	}
	scheme := cryptoScheme(c.CryptoType)
	var b strings.Builder
	b.WriteString(scheme)
	b.WriteString(":")
	b.WriteString(c.Address)
	params := []string{}
	if c.Amount != "" {
		params = append(params, "amount="+c.Amount)
	}
	if c.Label != "" {
		params = append(params, "label="+c.Label)
	}
	if c.Message != "" {
		params = append(params, "message="+c.Message)
	}
	if len(params) > 0 {
		b.WriteString("?")
		b.WriteString(strings.Join(params, "&"))
	}
	return b.String(), nil
}

// Validate checks that the address is non-empty and the crypto type is one of
// BTC, ETH, or LTC.
func (c *CryptoPayload) Validate() error {
	if c.Address == "" {
		return fmt.Errorf("crypto payload: address must not be empty")
	}
	if !isValidCryptoType(c.CryptoType) {
		return fmt.Errorf("crypto payload: unsupported crypto type %q, must be one of BTC, ETH, LTC", c.CryptoType)
	}
	return nil
}

// Type returns "crypto".
func (c *CryptoPayload) Type() string {
	return "crypto"
}

// Size returns the byte length of the encoded crypto URI.
func (c *CryptoPayload) Size() int {
	encoded, _ := c.Encode() //nolint:errcheck // Size returns 0 on encode error
	return len(encoded)
}

func isValidCryptoType(ct string) bool {
	switch ct {
	case CryptoBTC, CryptoETH, CryptoLTC:
		return true
	default:
		return false
	}
}

func cryptoScheme(ct string) string {
	switch ct {
	case CryptoBTC:
		return "bitcoin"
	case CryptoETH:
		return "ethereum"
	case CryptoLTC:
		return "litecoin"
	default:
		return strings.ToLower(ct)
	}
}
