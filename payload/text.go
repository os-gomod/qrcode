package payload

import "fmt"

// TextPayload encodes arbitrary plain text into a QR code.
// The text is stored as-is with no transformation, making it the simplest
// payload type. The maximum allowed length is 4296 characters, which
// corresponds to the QR code version 40, low error correction capacity.
type TextPayload struct {
	// Text is the content to encode.
	Text string
}

// Encode returns the text content as the QR code data string.
// The returned string is the raw Text value, unmodified.
func (t *TextPayload) Encode() (string, error) {
	if err := t.Validate(); err != nil {
		return "", err
	}
	return t.Text, nil
}

// Validate checks that the text is non-empty and within the maximum length
// of 4296 characters (QR version 40, EC level L).
func (t *TextPayload) Validate() error {
	if t.Text == "" {
		return fmt.Errorf("text payload: text must not be empty")
	}
	if len(t.Text) > maxTextLength {
		return fmt.Errorf("text payload: text length %d exceeds maximum of %d characters", len(t.Text), maxTextLength)
	}
	return nil
}

// Type returns "text".
func (t *TextPayload) Type() string {
	return "text"
}

// Size returns the byte length of the text.
func (t *TextPayload) Size() int {
	return len(t.Text)
}

const maxTextLength = 4296
