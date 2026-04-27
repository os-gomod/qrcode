package payload

import (
	"fmt"
	"strings"
)

type GeoPayload struct {
	Latitude  float64
	Longitude float64
}

func (g *GeoPayload) Encode() (string, error) {
	if err := g.Validate(); err != nil {
		return "", err
	}
	return fmt.Sprintf("geo:%s,%s",
		formatCoord(g.Latitude),
		formatCoord(g.Longitude),
	), nil
}

func (g *GeoPayload) Validate() error {
	return validateLatLong(g.Latitude, g.Longitude, "geo payload")
}

func (*GeoPayload) Type() string {
	return "geo"
}

func (g *GeoPayload) Size() int {
	encoded, _ := g.Encode()
	return len(encoded)
}

func formatCoord(c float64) string {
	s := strings.TrimRight(
		strings.TrimRight(fmt.Sprintf("%f", c), "0"),
		".",
	)
	return s
}

// validateLatLong checks that latitude and longitude are within valid ranges.
func validateLatLong(lat, lng float64, prefix string) error {
	if lat < -90 || lat > 90 {
		return fmt.Errorf("%s: latitude %f is out of range [-90, 90]", prefix, lat)
	}
	if lng < -180 || lng > 180 {
		return fmt.Errorf("%s: longitude %f is out of range [-180, 180]", prefix, lng)
	}
	return nil
}
