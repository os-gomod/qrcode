package payload

import (
	"fmt"
	"strings"
)

// TwitterPayload encodes a link to a Twitter (X) user profile.
// The username should not include the @ prefix.
//
// Example encoded output:
//
//	https://twitter.com/golang
type TwitterPayload struct {
	// Username is the Twitter username (without the @ prefix).
	Username string
}

// Encode returns a twitter.com profile URL.
func (t *TwitterPayload) Encode() (string, error) {
	if err := t.Validate(); err != nil {
		return "", err
	}
	return "https://twitter.com/" + t.Username, nil
}

// Validate checks that the username is non-empty.
func (t *TwitterPayload) Validate() error {
	if t.Username == "" {
		return fmt.Errorf("twitter payload: username must not be empty")
	}
	return nil
}

// Type returns "twitter".
func (t *TwitterPayload) Type() string {
	return "twitter"
}

// Size returns the byte length of the encoded URL.
func (t *TwitterPayload) Size() int {
	encoded, _ := t.Encode() //nolint:errcheck // Size returns 0 on encode error
	return len(encoded)
}

// TwitterFollowPayload encodes a link to a Twitter (X) user profile on
// x.com (the current Twitter domain). The screen name should not include
// the @ prefix.
//
// Example encoded output:
//
//	https://x.com/golang
type TwitterFollowPayload struct {
	// ScreenName is the screen name of the user to follow.
	ScreenName string
}

// Encode returns an x.com profile URL.
func (t *TwitterFollowPayload) Encode() (string, error) {
	if err := t.Validate(); err != nil {
		return "", err
	}
	return "https://x.com/" + t.ScreenName, nil
}

// Validate checks that the screen name is non-empty.
func (t *TwitterFollowPayload) Validate() error {
	if t.ScreenName == "" {
		return fmt.Errorf("twitter_follow payload: screen name must not be empty")
	}
	return nil
}

// Type returns "twitter_follow".
func (t *TwitterFollowPayload) Type() string {
	return "twitter_follow"
}

// Size returns the byte length of the encoded URL.
func (t *TwitterFollowPayload) Size() int {
	encoded, _ := t.Encode() //nolint:errcheck // Size returns 0 on encode error
	return len(encoded)
}

// LinkedInPayload encodes a link to a LinkedIn profile.
// The ProfileURL must be the full HTTPS URL to the profile page.
//
// Example encoded output:
//
//	https://www.linkedin.com/in/johndoe
type LinkedInPayload struct {
	// ProfileURL is the full HTTPS URL to the LinkedIn profile.
	ProfileURL string
}

// Encode returns the LinkedIn profile URL.
func (l *LinkedInPayload) Encode() (string, error) {
	if err := l.Validate(); err != nil {
		return "", err
	}
	return l.ProfileURL, nil
}

// Validate checks that the profile URL is non-empty and starts with https://.
func (l *LinkedInPayload) Validate() error {
	if l.ProfileURL == "" {
		return fmt.Errorf("linkedin payload: profile URL must not be empty")
	}
	if !strings.HasPrefix(l.ProfileURL, "https://") {
		return fmt.Errorf("linkedin payload: profile URL must start with https://")
	}
	return nil
}

// Type returns "linkedin".
func (l *LinkedInPayload) Type() string {
	return "linkedin"
}

// Size returns the byte length of the encoded URL.
func (l *LinkedInPayload) Size() int {
	encoded, _ := l.Encode() //nolint:errcheck // Size returns 0 on encode error
	return len(encoded)
}

// InstagramPayload encodes a link to an Instagram profile.
//
// Example encoded output:
//
//	https://instagram.com/natgeo
type InstagramPayload struct {
	// Username is the Instagram username.
	Username string
}

// Encode returns an instagram.com profile URL.
func (i *InstagramPayload) Encode() (string, error) {
	if err := i.Validate(); err != nil {
		return "", err
	}
	return "https://instagram.com/" + i.Username, nil
}

// Validate checks that the username is non-empty.
func (i *InstagramPayload) Validate() error {
	if i.Username == "" {
		return fmt.Errorf("instagram payload: username must not be empty")
	}
	return nil
}

// Type returns "instagram".
func (i *InstagramPayload) Type() string {
	return "instagram"
}

// Size returns the byte length of the encoded URL.
func (i *InstagramPayload) Size() int {
	encoded, _ := i.Encode() //nolint:errcheck // Size returns 0 on encode error
	return len(encoded)
}

// FacebookPayload encodes a link to a Facebook page.
// The PageURL must be the full HTTPS URL to the Facebook page.
//
// Example encoded output:
//
//	https://www.facebook.com/golang
type FacebookPayload struct {
	// PageURL is the full HTTPS URL to the Facebook page.
	PageURL string
}

// Encode returns the Facebook page URL.
func (f *FacebookPayload) Encode() (string, error) {
	if err := f.Validate(); err != nil {
		return "", err
	}
	return f.PageURL, nil
}

// Validate checks that the page URL is non-empty and starts with https://.
func (f *FacebookPayload) Validate() error {
	if f.PageURL == "" {
		return fmt.Errorf("facebook payload: page URL must not be empty")
	}
	if !strings.HasPrefix(f.PageURL, "https://") {
		return fmt.Errorf("facebook payload: page URL must start with https://")
	}
	return nil
}

// Type returns "facebook".
func (f *FacebookPayload) Type() string {
	return "facebook"
}

// Size returns the byte length of the encoded URL.
func (f *FacebookPayload) Size() int {
	encoded, _ := f.Encode() //nolint:errcheck // Size returns 0 on encode error
	return len(encoded)
}

// YouTubeChannelPayload encodes a link to a YouTube channel.
//
// Example encoded output:
//
//	https://youtube.com/channel/UC_x5XG1OV2P6uZZ5FSM9Ttw
type YouTubeChannelPayload struct {
	// ChannelID is the YouTube channel ID.
	ChannelID string
}

// Encode returns a youtube.com/channel/ URL.
func (y *YouTubeChannelPayload) Encode() (string, error) {
	if err := y.Validate(); err != nil {
		return "", err
	}
	return "https://youtube.com/channel/" + y.ChannelID, nil
}

// Validate checks that the channel ID is non-empty.
func (y *YouTubeChannelPayload) Validate() error {
	if y.ChannelID == "" {
		return fmt.Errorf("youtube_channel payload: channel ID must not be empty")
	}
	return nil
}

// Type returns "youtube_channel".
func (y *YouTubeChannelPayload) Type() string {
	return "youtube_channel"
}

// Size returns the byte length of the encoded URL.
func (y *YouTubeChannelPayload) Size() int {
	encoded, _ := y.Encode() //nolint:errcheck // Size returns 0 on encode error
	return len(encoded)
}

// TelegramPayload encodes a link to a Telegram user or group.
// The username should not include the @ prefix.
//
// Example encoded output:
//
//	https://t.me/golang
type TelegramPayload struct {
	// Username is the Telegram username (without the @ prefix).
	Username string
}

// Encode returns a t.me profile URL.
func (t *TelegramPayload) Encode() (string, error) {
	if err := t.Validate(); err != nil {
		return "", err
	}
	return "https://t.me/" + t.Username, nil
}

// Validate checks that the username is non-empty.
func (t *TelegramPayload) Validate() error {
	if t.Username == "" {
		return fmt.Errorf("telegram payload: username must not be empty")
	}
	return nil
}

// Type returns "telegram".
func (t *TelegramPayload) Type() string {
	return "telegram"
}

// Size returns the byte length of the encoded URL.
func (t *TelegramPayload) Size() int {
	encoded, _ := t.Encode() //nolint:errcheck // Size returns 0 on encode error
	return len(encoded)
}
