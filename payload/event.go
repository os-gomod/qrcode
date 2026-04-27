package payload

import (
	"errors"
	"strings"
	"time"
)

type EventPayload struct {
	EventID     string
	EventName   string
	Venue       string
	StartTime   time.Time
	Category    string
	Seat        string
	Organizer   string
	Description string
	URL         string
}

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

func (e *EventPayload) Validate() error {
	if e.EventID == "" {
		return errors.New("event payload: event ID must not be empty")
	}
	return nil
}

func (*EventPayload) Type() string {
	return "event"
}

func (e *EventPayload) Size() int {
	encoded, _ := e.Encode()
	return len(encoded)
}
