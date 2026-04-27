package payload

import (
	"errors"
	"net/url"
	"strings"
)

type ZoomPayload struct {
	MeetingID   string
	Password    string
	DisplayName string
}

func (z *ZoomPayload) Encode() (string, error) {
	if err := z.Validate(); err != nil {
		return "", err
	}
	result := "https://zoom.us/j/" + z.MeetingID
	params := []string{}
	if z.Password != "" {
		params = append(params, "pwd="+url.QueryEscape(z.Password))
	}
	if z.DisplayName != "" {
		params = append(params, "uname="+url.QueryEscape(z.DisplayName))
	}
	if len(params) > 0 {
		result += "?" + strings.Join(params, "&")
	}
	return result, nil
}

func (z *ZoomPayload) Validate() error {
	if z.MeetingID == "" {
		return errors.New("zoom payload: meeting ID must not be empty")
	}
	return nil
}

func (*ZoomPayload) Type() string {
	return "zoom"
}

func (z *ZoomPayload) Size() int {
	encoded, _ := z.Encode()
	return len(encoded)
}
