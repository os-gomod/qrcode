package payload

import (
	"errors"
	"fmt"
)

type TextPayload struct {
	Text string
}

func (t *TextPayload) Encode() (string, error) {
	if err := t.Validate(); err != nil {
		return "", err
	}
	return t.Text, nil
}

func (t *TextPayload) Validate() error {
	if t.Text == "" {
		return errors.New("text payload: text must not be empty")
	}
	if len(t.Text) > maxTextLength {
		return fmt.Errorf("text payload: text length %d exceeds maximum of %d characters", len(t.Text), maxTextLength)
	}
	return nil
}

func (*TextPayload) Type() string {
	return "text"
}

func (t *TextPayload) Size() int {
	return len(t.Text)
}

const maxTextLength = 4296
