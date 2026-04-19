package payload

import (
	"fmt"
	"net/url"
)

// DefaultPayPalCurrency is the default currency for PayPal payment links.
// Used when the Currency field is empty. Value: "USD".
const DefaultPayPalCurrency = "USD"

// PayPalPayload encodes a PayPal.me payment link. When scanned, the QR code
// opens the PayPal payment page pre-filled with the specified username, amount,
// and currency. The currency defaults to USD if not specified.
//
// Example encoded output:
//
//	https://paypal.me/johndoe/25.00/USD&note=Invoice%2042
type PayPalPayload struct {
	// Username is the PayPal.me username.
	Username string
	// Amount is the payment amount.
	Amount string
	// Currency is the three-letter currency code (defaults to USD).
	Currency string
	// Reference is an optional payment reference note.
	Reference string
}

// Encode returns a paypal.me URL with amount, currency (defaults to USD),
// and optional reference note. Format: https://paypal.me/<user>/<amount>/<currency>.
func (p *PayPalPayload) Encode() (string, error) {
	if err := p.Validate(); err != nil {
		return "", err
	}
	currency := p.Currency
	if currency == "" {
		currency = DefaultPayPalCurrency
	}
	result := fmt.Sprintf("https://paypal.me/%s/%s/%s",
		url.PathEscape(p.Username), url.PathEscape(p.Amount), url.QueryEscape(currency))
	if p.Reference != "" {
		result += "&note=" + url.QueryEscape(p.Reference)
	}
	return result, nil
}

// Validate checks that the username and amount are non-empty.
func (p *PayPalPayload) Validate() error {
	if p.Username == "" {
		return fmt.Errorf("paypal payload: username must not be empty")
	}
	if p.Amount == "" {
		return fmt.Errorf("paypal payload: amount must not be empty")
	}
	return nil
}

// Type returns "paypal".
func (p *PayPalPayload) Type() string {
	return "paypal"
}

// Size returns the byte length of the encoded PayPal URL.
func (p *PayPalPayload) Size() int {
	encoded, _ := p.Encode() //nolint:errcheck // Size returns 0 on encode error
	return len(encoded)
}
