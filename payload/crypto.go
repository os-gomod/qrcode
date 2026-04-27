package payload

import (
	"errors"
	"fmt"
	"strings"
)

const (
	CryptoBTC = "BTC"
	CryptoETH = "ETH"
	CryptoLTC = "LTC"
)

type CryptoPayload struct {
	Address    string
	Amount     string
	Label      string
	Message    string
	CryptoType string
}

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

func (c *CryptoPayload) Validate() error {
	if c.Address == "" {
		return errors.New("crypto payload: address must not be empty")
	}
	if !isValidCryptoType(c.CryptoType) {
		return fmt.Errorf("crypto payload: unsupported crypto type %q, must be one of BTC, ETH, LTC", c.CryptoType)
	}
	return nil
}

func (*CryptoPayload) Type() string {
	return "crypto"
}

func (c *CryptoPayload) Size() int {
	encoded, _ := c.Encode()
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
