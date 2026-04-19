package payload

import (
	"fmt"
	"strings"
)

// GeoPayload encodes a geographic location as a geo: URI per RFC 5870.
// The URI contains latitude and longitude in decimal degrees. When scanned
// by a mapping application, the device will typically open a pin at the
// specified coordinates.
//
// Example encoded output:
//
//	geo:37.7749,-122.4194
type GeoPayload struct {
	// Latitude is the latitude in decimal degrees (-90 to 90).
	Latitude float64
	// Longitude is the longitude in decimal degrees (-180 to 180).
	Longitude float64
}

// Encode returns a geo: URI with latitude and longitude coordinates.
// Trailing zeros in the decimal representation are trimmed for compactness.
// Format: geo:<lat>,<lng>.
func (g *GeoPayload) Encode() (string, error) {
	if err := g.Validate(); err != nil {
		return "", err
	}
	return fmt.Sprintf("geo:%s,%s",
		formatCoord(g.Latitude),
		formatCoord(g.Longitude),
	), nil
}

// Validate checks that the latitude is within [-90, 90] and the longitude
// is within [-180, 180].
func (g *GeoPayload) Validate() error {
	if g.Latitude < -90 || g.Latitude > 90 {
		return fmt.Errorf("geo payload: latitude %f is out of range [-90, 90]", g.Latitude)
	}
	if g.Longitude < -180 || g.Longitude > 180 {
		return fmt.Errorf("geo payload: longitude %f is out of range [-180, 180]", g.Longitude)
	}
	return nil
}

// Type returns "geo".
func (g *GeoPayload) Type() string {
	return "geo"
}

// Size returns the byte length of the encoded geo URI.
func (g *GeoPayload) Size() int {
	encoded, _ := g.Encode() //nolint:errcheck // Size returns 0 on encode error
	return len(encoded)
}

func formatCoord(c float64) string {
	s := strings.TrimRight(
		strings.TrimRight(fmt.Sprintf("%f", c), "0"),
		".",
	)
	return s
}
