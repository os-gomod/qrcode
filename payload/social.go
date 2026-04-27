package payload

import (
	"errors"
	"strings"
)

type TwitterPayload struct {
	Username string
}

func (t *TwitterPayload) Encode() (string, error) {
	if err := t.Validate(); err != nil {
		return "", err
	}
	return "https://twitter.com/" + t.Username, nil
}

func (t *TwitterPayload) Validate() error {
	if t.Username == "" {
		return errors.New("twitter payload: username must not be empty")
	}
	return nil
}

func (*TwitterPayload) Type() string {
	return "twitter"
}

func (t *TwitterPayload) Size() int {
	encoded, _ := t.Encode()
	return len(encoded)
}

type TwitterFollowPayload struct {
	ScreenName string
}

func (t *TwitterFollowPayload) Encode() (string, error) {
	if err := t.Validate(); err != nil {
		return "", err
	}
	return "https://twitter.com/intent/follow?screen_name=" + t.ScreenName, nil
}

func (t *TwitterFollowPayload) Validate() error {
	if t.ScreenName == "" {
		return errors.New("twitter_follow payload: screen name must not be empty")
	}
	return nil
}

func (*TwitterFollowPayload) Type() string {
	return "twitter_follow"
}

func (t *TwitterFollowPayload) Size() int {
	encoded, _ := t.Encode()
	return len(encoded)
}

type LinkedInPayload struct {
	ProfileURL string
}

func (l *LinkedInPayload) Encode() (string, error) {
	if err := l.Validate(); err != nil {
		return "", err
	}
	return l.ProfileURL, nil
}

func (l *LinkedInPayload) Validate() error {
	if l.ProfileURL == "" {
		return errors.New("linkedin payload: profile URL must not be empty")
	}
	if !strings.HasPrefix(l.ProfileURL, "https://") {
		return errors.New("linkedin payload: profile URL must start with https://")
	}
	return nil
}

func (*LinkedInPayload) Type() string {
	return "linkedin"
}

func (l *LinkedInPayload) Size() int {
	encoded, _ := l.Encode()
	return len(encoded)
}

type InstagramPayload struct {
	Username string
}

func (i *InstagramPayload) Encode() (string, error) {
	if err := i.Validate(); err != nil {
		return "", err
	}
	return "https://instagram.com/" + i.Username, nil
}

func (i *InstagramPayload) Validate() error {
	if i.Username == "" {
		return errors.New("instagram payload: username must not be empty")
	}
	return nil
}

func (*InstagramPayload) Type() string {
	return "instagram"
}

func (i *InstagramPayload) Size() int {
	encoded, _ := i.Encode()
	return len(encoded)
}

type FacebookPayload struct {
	PageURL string
}

func (f *FacebookPayload) Encode() (string, error) {
	if err := f.Validate(); err != nil {
		return "", err
	}
	return f.PageURL, nil
}

func (f *FacebookPayload) Validate() error {
	if f.PageURL == "" {
		return errors.New("facebook payload: page URL must not be empty")
	}
	if !strings.HasPrefix(f.PageURL, "https://") {
		return errors.New("facebook payload: page URL must start with https://")
	}
	return nil
}

func (*FacebookPayload) Type() string {
	return "facebook"
}

func (f *FacebookPayload) Size() int {
	encoded, _ := f.Encode()
	return len(encoded)
}

type YouTubeChannelPayload struct {
	ChannelID string
}

func (y *YouTubeChannelPayload) Encode() (string, error) {
	if err := y.Validate(); err != nil {
		return "", err
	}
	return "https://www.youtube.com/channel/" + y.ChannelID, nil
}

func (y *YouTubeChannelPayload) Validate() error {
	if y.ChannelID == "" {
		return errors.New("youtube_channel payload: channel ID must not be empty")
	}
	return nil
}

func (*YouTubeChannelPayload) Type() string {
	return "youtube_channel"
}

func (y *YouTubeChannelPayload) Size() int {
	encoded, _ := y.Encode()
	return len(encoded)
}

type TelegramPayload struct {
	Username string
}

func (t *TelegramPayload) Encode() (string, error) {
	if err := t.Validate(); err != nil {
		return "", err
	}
	return "https://t.me/" + t.Username, nil
}

func (t *TelegramPayload) Validate() error {
	if t.Username == "" {
		return errors.New("telegram payload: username must not be empty")
	}
	return nil
}

func (*TelegramPayload) Type() string {
	return "telegram"
}

func (t *TelegramPayload) Size() int {
	encoded, _ := t.Encode()
	return len(encoded)
}
