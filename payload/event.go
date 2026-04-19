package payload

import (
	"fmt"
	"strings"
	"time"
)

// EventPayload encodes a generic event ticket using the EVENT-TICKET
// format. Fields are pipe-separated after the "EVENT-TICKET:" prefix.
// Only non-empty fields are included. This format is useful for
// conference tickets, concert passes, and similar event admissions.
//
// Example encoded output:
//
//	EVENT-TICKET:evt-001|GopherCon 2025|San Francisco|20250715T090000Z|Conference|A12|Go Team
type EventPayload struct {
	// EventID is the unique event identifier.
	EventID string
	// EventName is the name of the event.
	EventName string
	// Venue is the event venue.
	Venue string
	// StartTime is the event start time.
	StartTime time.Time
	// Category is the event category.
	Category string
	// Seat is the seat assignment.
	Seat string
	// Organizer is the event organizer.
	Organizer string
	// Description is the event description.
	Description string
	// URL is a link to the event page.
	URL string
}

// Encode returns an EVENT-TICKET: string with pipe-separated fields.
// Only non-empty fields (EventID, EventName, Venue, StartTime, Category,
// Seat, Organizer, Description, URL) are appended.
func (e *EventPayload) Encode() (string, error) {
	if err := e.Validate(); err != nil {
		return "", err
	}
	fields := []string{e.EventID}
	if e.EventName != "" {
		fields = append(fields, e.EventName)
	}
	if e.Venue != "" {
		fields = append(fields, e.Venue)
	}
	if !e.StartTime.IsZero() {
		fields = append(fields, e.StartTime.UTC().Format(dateTimeLayout))
	}
	if e.Category != "" {
		fields = append(fields, e.Category)
	}
	if e.Seat != "" {
		fields = append(fields, e.Seat)
	}
	if e.Organizer != "" {
		fields = append(fields, e.Organizer)
	}
	if e.Description != "" {
		fields = append(fields, e.Description)
	}
	if e.URL != "" {
		fields = append(fields, e.URL)
	}
	return "EVENT-TICKET:" + strings.Join(fields, "|"), nil
}

// Validate checks that the event ID is non-empty.
func (e *EventPayload) Validate() error {
	if e.EventID == "" {
		return fmt.Errorf("event payload: event ID must not be empty")
	}
	return nil
}

// Type returns "event".
func (e *EventPayload) Type() string {
	return "event"
}

// Size returns the byte length of the encoded event string.
func (e *EventPayload) Size() int {
	encoded, _ := e.Encode() //nolint:errcheck // Size returns 0 on encode error
	return len(encoded)
}
