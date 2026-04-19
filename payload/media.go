package payload

import (
	"fmt"
	"net/url"
)

// SpotifyTrackPayload encodes a link to a Spotify track using the
// open.spotify.com/track/ deep link format.
//
// Example encoded output:
//
//	https://open.spotify.com/track/4cOdK2wGLETKBW3PvgPWqT
type SpotifyTrackPayload struct {
	// TrackID is the Spotify track ID.
	TrackID string
}

// Encode returns a Spotify open.spotify.com/track/ URL.
func (s *SpotifyTrackPayload) Encode() (string, error) {
	if err := s.Validate(); err != nil {
		return "", err
	}
	return "https://open.spotify.com/track/" + s.TrackID, nil
}

// Validate checks that the track ID is non-empty.
func (s *SpotifyTrackPayload) Validate() error {
	if s.TrackID == "" {
		return fmt.Errorf("spotify_track payload: track ID must not be empty")
	}
	return nil
}

// Type returns "spotify_track".
func (s *SpotifyTrackPayload) Type() string {
	return "spotify_track"
}

// Size returns the byte length of the encoded URL.
func (s *SpotifyTrackPayload) Size() int {
	encoded, _ := s.Encode() //nolint:errcheck // Size returns 0 on encode error
	return len(encoded)
}

// SpotifyPlaylistPayload encodes a link to a Spotify playlist using the
// open.spotify.com/playlist/ deep link format.
//
// Example encoded output:
//
//	https://open.spotify.com/playlist/37i9dQZF1DXcBWIGoYBM5M
type SpotifyPlaylistPayload struct {
	// PlaylistID is the Spotify playlist ID.
	PlaylistID string
}

// Encode returns a Spotify open.spotify.com/playlist/ URL.
func (s *SpotifyPlaylistPayload) Encode() (string, error) {
	if err := s.Validate(); err != nil {
		return "", err
	}
	return "https://open.spotify.com/playlist/" + s.PlaylistID, nil
}

// Validate checks that the playlist ID is non-empty.
func (s *SpotifyPlaylistPayload) Validate() error {
	if s.PlaylistID == "" {
		return fmt.Errorf("spotify_playlist payload: playlist ID must not be empty")
	}
	return nil
}

// Type returns "spotify_playlist".
func (s *SpotifyPlaylistPayload) Type() string {
	return "spotify_playlist"
}

// Size returns the byte length of the encoded URL.
func (s *SpotifyPlaylistPayload) Size() int {
	encoded, _ := s.Encode() //nolint:errcheck // Size returns 0 on encode error
	return len(encoded)
}

// AppleMusicTrackPayload encodes a link to an Apple Music track.
// The URL requires an AlbumID and SongID. An optional StoreFront
// parameter specifies the regional storefront (e.g. "us").
//
// Example encoded output:
//
//	https://music.apple.com/album/1234567890?i=1234567891
type AppleMusicTrackPayload struct {
	// AlbumID is the Apple Music album ID.
	AlbumID string
	// SongID is the Apple Music song ID.
	SongID string
	// StoreFront is the optional regional storefront identifier.
	StoreFront string
}

// Encode returns an Apple Music album URL with the song parameter.
// If StoreFront is set, the regional storefront path segment is included.
// AlbumID and SongID are path/query-encoded respectively.
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

// Validate checks that both album ID and song ID are non-empty.
func (a *AppleMusicTrackPayload) Validate() error {
	if a.AlbumID == "" {
		return fmt.Errorf("apple_music payload: album ID must not be empty")
	}
	if a.SongID == "" {
		return fmt.Errorf("apple_music payload: song ID must not be empty")
	}
	return nil
}

// Type returns "apple_music".
func (a *AppleMusicTrackPayload) Type() string {
	return "apple_music"
}

// Size returns the byte length of the encoded URL.
func (a *AppleMusicTrackPayload) Size() int {
	encoded, _ := a.Encode() //nolint:errcheck // Size returns 0 on encode error
	return len(encoded)
}

// YouTubeVideoPayload encodes a link to a YouTube video.
//
// Example encoded output:
//
//	https://youtube.com/watch?v=dQw4w9WgXcQ
type YouTubeVideoPayload struct {
	// VideoID is the YouTube video ID.
	VideoID string
}

// Encode returns a YouTube watch URL.
func (y *YouTubeVideoPayload) Encode() (string, error) {
	if err := y.Validate(); err != nil {
		return "", err
	}
	return "https://youtube.com/watch?v=" + y.VideoID, nil
}

// Validate checks that the video ID is non-empty.
func (y *YouTubeVideoPayload) Validate() error {
	if y.VideoID == "" {
		return fmt.Errorf("youtube_video payload: video ID must not be empty")
	}
	return nil
}

// Type returns "youtube_video".
func (y *YouTubeVideoPayload) Type() string {
	return "youtube_video"
}

// Size returns the byte length of the encoded URL.
func (y *YouTubeVideoPayload) Size() int {
	encoded, _ := y.Encode() //nolint:errcheck // Size returns 0 on encode error
	return len(encoded)
}
