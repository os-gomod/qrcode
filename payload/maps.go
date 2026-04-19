package payload

import (
	"fmt"
	"net/url"
	"strconv"
)

// TravelModeDriving represents driving directions.
// Used as the TravelMode field in GoogleMapsDirectionsPayload.
const TravelModeDriving = "driving"

// TravelModeWalking represents walking directions.
// Used as the TravelMode field in GoogleMapsDirectionsPayload.
const TravelModeWalking = "walking"

// TravelModeBicycling represents bicycling directions.
// Used as the TravelMode field in GoogleMapsDirectionsPayload.
const TravelModeBicycling = "bicycling"

// TravelModeTransit represents public transit directions.
// Used as the TravelMode field in GoogleMapsDirectionsPayload.
const TravelModeTransit = "transit"

// GoogleMapsPayload encodes a Google Maps location or search query.
// When a Query is set, it takes precedence over coordinates and produces a
// maps.google.com/maps?q=<query> URL. Otherwise, the coordinates are
// encoded as a loc:<lat>,<lng> parameter.
//
// Example encoded output (coordinates):
//
//	https://maps.google.com/maps?q=loc:37.7749,-122.4194
//
// Example encoded output (query):
//
//	https://maps.google.com/maps?q=coffee+shop
type GoogleMapsPayload struct {
	// Latitude is the center latitude.
	Latitude float64
	// Longitude is the center longitude.
	Longitude float64
	// Query is an optional search query (overrides coordinates when set).
	Query string
	// Zoom is an optional zoom level.
	Zoom int
}

// Encode returns a Google Maps URL for the location or query.
// If Query is non-empty, it is used as the q parameter. Otherwise,
// coordinates are formatted as loc:<lat>,<lng>. An optional Zoom level
// is appended as &zoom=N.
func (g *GoogleMapsPayload) Encode() (string, error) {
	if err := g.Validate(); err != nil {
		return "", err
	}
	var result string
	if g.Query != "" {
		result = "https://maps.google.com/maps?q=" + url.QueryEscape(g.Query)
	} else {
		result = fmt.Sprintf("https://maps.google.com/maps?q=loc:%s,%s",
			formatCoord(g.Latitude), formatCoord(g.Longitude))
	}
	if g.Zoom > 0 {
		result += "&zoom=" + strconv.Itoa(g.Zoom)
	}
	return result, nil
}

// Validate checks that either a query is non-empty or valid coordinate
// ranges are used (latitude in [-90, 90], longitude in [-180, 180]).
func (g *GoogleMapsPayload) Validate() error {
	if g.Query != "" {
		return nil
	}
	if g.Latitude < -90 || g.Latitude > 90 {
		return fmt.Errorf("google_maps payload: latitude %f is out of range [-90, 90]", g.Latitude)
	}
	if g.Longitude < -180 || g.Longitude > 180 {
		return fmt.Errorf("google_maps payload: longitude %f is out of range [-180, 180]", g.Longitude)
	}
	return nil
}

// Type returns "google_maps".
func (g *GoogleMapsPayload) Type() string {
	return "google_maps"
}

// Size returns the byte length of the encoded URL.
func (g *GoogleMapsPayload) Size() int {
	encoded, _ := g.Encode() //nolint:errcheck // Size returns 0 on encode error
	return len(encoded)
}

// GoogleMapsPlacePayload encodes a Google Maps place search query.
// The PlaceName is URL-encoded and used as the q parameter in the
// maps.google.com/maps?q= URL.
//
// Example encoded output:
//
//	https://maps.google.com/maps?q=Eiffel+Tower
type GoogleMapsPlacePayload struct {
	// PlaceName is the place or business name to search for.
	PlaceName string
}

// Encode returns a Google Maps search URL for the place name.
// The PlaceName is URL-encoded as the q query parameter.
func (g *GoogleMapsPlacePayload) Encode() (string, error) {
	if err := g.Validate(); err != nil {
		return "", err
	}
	return "https://maps.google.com/maps?q=" + url.QueryEscape(g.PlaceName), nil
}

// Validate checks that the place name is non-empty.
func (g *GoogleMapsPlacePayload) Validate() error {
	if g.PlaceName == "" {
		return fmt.Errorf("google_maps_place payload: place name must not be empty")
	}
	return nil
}

// Type returns "google_maps_place".
func (g *GoogleMapsPlacePayload) Type() string {
	return "google_maps_place"
}

// Size returns the byte length of the encoded URL.
func (g *GoogleMapsPlacePayload) Size() int {
	encoded, _ := g.Encode() //nolint:errcheck // Size returns 0 on encode error
	return len(encoded)
}

// GoogleMapsDirectionsPayload encodes Google Maps turn-by-turn
// directions between two points. It uses the Google Maps Directions API
// URL format with origin, destination, and optional travel mode.
//
// Supported travel modes: driving, walking, bicycling, transit.
//
// Example encoded output:
//
//	https://maps.google.com/maps/dir/?api=1&origin=San+Francisco&destination=Los+Angeles&travelmode=driving
type GoogleMapsDirectionsPayload struct {
	// Origin is the starting location.
	Origin string
	// Destination is the ending location.
	Destination string
	// TravelMode is the mode of transportation (defaults to "driving").
	TravelMode string
}

// Encode returns a Google Maps directions URL with origin, destination,
// and travel mode. Origin and destination are URL-encoded.
func (g *GoogleMapsDirectionsPayload) Encode() (string, error) {
	if err := g.Validate(); err != nil {
		return "", err
	}
	travelMode := g.TravelMode
	if travelMode == "" {
		travelMode = TravelModeDriving
	}
	result := fmt.Sprintf("https://maps.google.com/maps/dir/?api=1&origin=%s&destination=%s&travelmode=%s",
		url.QueryEscape(g.Origin), url.QueryEscape(g.Destination), url.QueryEscape(travelMode))
	return result, nil
}

// Validate checks that both origin and destination are non-empty.
func (g *GoogleMapsDirectionsPayload) Validate() error {
	if g.Origin == "" {
		return fmt.Errorf("google_maps_directions payload: origin must not be empty")
	}
	if g.Destination == "" {
		return fmt.Errorf("google_maps_directions payload: destination must not be empty")
	}
	return nil
}

// Type returns "google_maps_directions".
func (g *GoogleMapsDirectionsPayload) Type() string {
	return "google_maps_directions"
}

// Size returns the byte length of the encoded URL.
func (g *GoogleMapsDirectionsPayload) Size() int {
	encoded, _ := g.Encode() //nolint:errcheck // Size returns 0 on encode error
	return len(encoded)
}

// AppleMapsPayload encodes an Apple Maps location or search query.
// Coordinates are encoded as the ll (latitude/longitude) query parameter.
// An optional search query can be appended as the q parameter, and an
// optional zoom level as the t parameter.
//
// Example encoded output:
//
//	https://maps.apple.com/maps?ll=37.7749,-122.4194
type AppleMapsPayload struct {
	// Latitude is the center latitude.
	Latitude float64
	// Longitude is the center longitude.
	Longitude float64
	// Query is an optional search query.
	Query string
	// Zoom is an optional zoom level.
	Zoom int
}

// Encode returns an Apple Maps URL with coordinates (ll parameter),
// optional search query (q parameter), and optional zoom (t parameter).
func (a *AppleMapsPayload) Encode() (string, error) {
	if err := a.Validate(); err != nil {
		return "", err
	}
	result := fmt.Sprintf("https://maps.apple.com/maps?ll=%s,%s",
		formatCoord(a.Latitude), formatCoord(a.Longitude))
	if a.Query != "" {
		result += "&q=" + url.QueryEscape(a.Query)
	}
	if a.Zoom > 0 {
		result += "&t=" + strconv.Itoa(a.Zoom)
	}
	return result, nil
}

// Validate checks that the coordinates are within valid ranges:
// latitude in [-90, 90], longitude in [-180, 180].
func (a *AppleMapsPayload) Validate() error {
	if a.Latitude < -90 || a.Latitude > 90 {
		return fmt.Errorf("apple_maps payload: latitude %f is out of range [-90, 90]", a.Latitude)
	}
	if a.Longitude < -180 || a.Longitude > 180 {
		return fmt.Errorf("apple_maps payload: longitude %f is out of range [-180, 180]", a.Longitude)
	}
	return nil
}

// Type returns "apple_maps".
func (a *AppleMapsPayload) Type() string {
	return "apple_maps"
}

// Size returns the byte length of the encoded URL.
func (a *AppleMapsPayload) Size() int {
	encoded, _ := a.Encode() //nolint:errcheck // Size returns 0 on encode error
	return len(encoded)
}
