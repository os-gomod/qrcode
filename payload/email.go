package payload

import (
	"errors"
	"net/url"
	"strings"
)

type EmailPayload struct {
	To      string
	Subject string
	Body    string
	CC      []string
}

func (e *EmailPayload) Encode() (string, error) {
	if err := e.Validate(); err != nil {
		return "", err
	}
	var b strings.Builder
	b.WriteString("mailto:")
	b.WriteString(e.To)
	params := []string{}
	if e.Subject != "" {
		params = append(params, "subject="+url.QueryEscape(e.Subject))
	}
	if e.Body != "" {
		params = append(params, "body="+url.QueryEscape(e.Body))
	}
	for _, cc := range e.CC {
		if cc != "" {
			params = append(params, "cc="+url.QueryEscape(cc))
		}
	}
	if len(params) > 0 {
		b.WriteString("?")
		b.WriteString(strings.Join(params, "&"))
	}
	return b.String(), nil
}

func (e *EmailPayload) Validate() error {
	if e.To == "" {
		return errors.New("email payload: recipient (To) must not be empty")
	}
	return nil
}

func (*EmailPayload) Type() string {
	return "email"
}

func (e *EmailPayload) Size() int {
	encoded, _ := e.Encode()
	return len(encoded)
}
