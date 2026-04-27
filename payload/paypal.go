package payload

import (
	"errors"
	"fmt"
	"net/url"
)

const DefaultPayPalCurrency = "USD"

type PayPalPayload struct {
	Username  string
	Amount    string
	Currency  string
	Reference string
}

func (p *PayPalPayload) Encode() (string, error) {
	if err := p.Validate(); err != nil {
		return "", err
	}
	currency := p.Currency
	if currency == "" {
		currency = DefaultPayPalCurrency
	}
	result := fmt.Sprintf("https://www.paypal.me/%s/%s/%s",
		url.PathEscape(p.Username), url.PathEscape(p.Amount), url.QueryEscape(currency))
	if p.Reference != "" {
		result += "&note=" + url.QueryEscape(p.Reference)
	}
	return result, nil
}

func (p *PayPalPayload) Validate() error {
	if p.Username == "" {
		return errors.New("paypal payload: username must not be empty")
	}
	if p.Amount == "" {
		return errors.New("paypal payload: amount must not be empty")
	}
	return nil
}

func (*PayPalPayload) Type() string {
	return "paypal"
}

func (p *PayPalPayload) Size() int {
	encoded, _ := p.Encode()
	return len(encoded)
}
