package payload

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
)

const (
	TravelModeDriving   = "driving"
	TravelModeWalking   = "walking"
	TravelModeBicycling = "bicycling"
	TravelModeTransit   = "transit"
)

type GoogleMapsPayload struct {
	Latitude  float64
	Longitude float64
	Query     string
	Zoom      int
}

func (g *GoogleMapsPayload) Encode() (string, error) {
	if err := g.Validate(); err != nil {
		return "", err
	}
	var result string
	if g.Query != "" {
		result = "https://maps.google.com/?q=" + url.QueryEscape(g.Query)
	} else {
		result = fmt.Sprintf("https://maps.google.com/?q=%s,%s",
			formatCoord(g.Latitude), formatCoord(g.Longitude))
	}
	if g.Zoom > 0 {
		result += "&zoom=" + strconv.Itoa(g.Zoom)
	}
	return result, nil
}

func (g *GoogleMapsPayload) Validate() error {
	if g.Query != "" {
		return nil
	}
	return validateLatLong(g.Latitude, g.Longitude, "google_maps payload")
}

func (*GoogleMapsPayload) Type() string {
	return "google_maps"
}

func (g *GoogleMapsPayload) Size() int {
	encoded, _ := g.Encode()
	return len(encoded)
}

type GoogleMapsPlacePayload struct {
	PlaceName string
}

func (g *GoogleMapsPlacePayload) Encode() (string, error) {
	if err := g.Validate(); err != nil {
		return "", err
	}
	return "https://www.google.com/maps/place/" + url.QueryEscape(g.PlaceName), nil
}

func (g *GoogleMapsPlacePayload) Validate() error {
	if g.PlaceName == "" {
		return errors.New("google_maps_place payload: place name must not be empty")
	}
	return nil
}

func (*GoogleMapsPlacePayload) Type() string {
	return "google_maps_place"
}

func (g *GoogleMapsPlacePayload) Size() int {
	encoded, _ := g.Encode()
	return len(encoded)
}

type GoogleMapsDirectionsPayload struct {
	Origin      string
	Destination string
	TravelMode  string
}

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

func (g *GoogleMapsDirectionsPayload) Validate() error {
	if g.Origin == "" {
		return errors.New("google_maps_directions payload: origin must not be empty")
	}
	if g.Destination == "" {
		return errors.New("google_maps_directions payload: destination must not be empty")
	}
	return nil
}

func (*GoogleMapsDirectionsPayload) Type() string {
	return "google_maps_directions"
}

func (g *GoogleMapsDirectionsPayload) Size() int {
	encoded, _ := g.Encode()
	return len(encoded)
}

type AppleMapsPayload struct {
	Latitude  float64
	Longitude float64
	Query     string
	Zoom      int
}

func (a *AppleMapsPayload) Encode() (string, error) {
	if err := a.Validate(); err != nil {
		return "", err
	}
	result := fmt.Sprintf("https://maps.apple.com/?ll=%s,%s",
		formatCoord(a.Latitude), formatCoord(a.Longitude))
	if a.Query != "" {
		result += "&q=" + url.QueryEscape(a.Query)
	}
	if a.Zoom > 0 {
		result += "&t=" + strconv.Itoa(a.Zoom)
	}
	return result, nil
}

func (a *AppleMapsPayload) Validate() error {
	return validateLatLong(a.Latitude, a.Longitude, "apple_maps payload")
}

func (*AppleMapsPayload) Type() string {
	return "apple_maps"
}

func (a *AppleMapsPayload) Size() int {
	encoded, _ := a.Encode()
	return len(encoded)
}
