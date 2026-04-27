package payload

import (
	"errors"
	"fmt"
	"net/url"
)

type SpotifyTrackPayload struct {
	TrackID string
}

func (s *SpotifyTrackPayload) Encode() (string, error) {
	if err := s.Validate(); err != nil {
		return "", err
	}
	return "https://open.spotify.com/track/" + s.TrackID, nil
}

func (s *SpotifyTrackPayload) Validate() error {
	if s.TrackID == "" {
		return errors.New("spotify_track payload: track ID must not be empty")
	}
	return nil
}

func (*SpotifyTrackPayload) Type() string {
	return "spotify_track"
}

func (s *SpotifyTrackPayload) Size() int {
	encoded, _ := s.Encode()
	return len(encoded)
}

type SpotifyPlaylistPayload struct {
	PlaylistID string
}

func (s *SpotifyPlaylistPayload) Encode() (string, error) {
	if err := s.Validate(); err != nil {
		return "", err
	}
	return "https://open.spotify.com/playlist/" + s.PlaylistID, nil
}

func (s *SpotifyPlaylistPayload) Validate() error {
	if s.PlaylistID == "" {
		return errors.New("spotify_playlist payload: playlist ID must not be empty")
	}
	return nil
}

func (*SpotifyPlaylistPayload) Type() string {
	return "spotify_playlist"
}

func (s *SpotifyPlaylistPayload) Size() int {
	encoded, _ := s.Encode()
	return len(encoded)
}

type AppleMusicTrackPayload struct {
	AlbumID    string
	SongID     string
	StoreFront string
}

func (a *AppleMusicTrackPayload) Encode() (string, error) {
	if err := a.Validate(); err != nil {
		return "", err
	}
	var result string
	if a.StoreFront != "" {
		result = fmt.Sprintf("https://music.apple.com/%s/album/%s?i=%s",
			url.PathEscape(a.StoreFront), url.PathEscape(a.AlbumID), url.QueryEscape(a.SongID))
	} else {
		result = fmt.Sprintf("https://music.apple.com/album/%s?i=%s",
			url.PathEscape(a.AlbumID), url.QueryEscape(a.SongID))
	}
	return result, nil
}

func (a *AppleMusicTrackPayload) Validate() error {
	if a.AlbumID == "" {
		return errors.New("apple_music payload: album ID must not be empty")
	}
	if a.SongID == "" {
		return errors.New("apple_music payload: song ID must not be empty")
	}
	return nil
}

func (*AppleMusicTrackPayload) Type() string {
	return "apple_music"
}

func (a *AppleMusicTrackPayload) Size() int {
	encoded, _ := a.Encode()
	return len(encoded)
}

type YouTubeVideoPayload struct {
	VideoID string
}

func (y *YouTubeVideoPayload) Encode() (string, error) {
	if err := y.Validate(); err != nil {
		return "", err
	}
	return "https://www.youtube.com/watch?v=" + y.VideoID, nil
}

func (y *YouTubeVideoPayload) Validate() error {
	if y.VideoID == "" {
		return errors.New("youtube_video payload: video ID must not be empty")
	}
	return nil
}

func (*YouTubeVideoPayload) Type() string {
	return "youtube_video"
}

func (y *YouTubeVideoPayload) Size() int {
	encoded, _ := y.Encode()
	return len(encoded)
}
