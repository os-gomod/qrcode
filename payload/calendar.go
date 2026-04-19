package payload

import (
	"fmt"
	"strings"
	"time"
)

// CalendarPayload encodes a calendar event using the iCalendar VEVENT
// format (RFC 5545). The output is a self-contained VEVENT block with
// SUMMARY, DESCRIPTION, LOCATION, and DTSTART/DTEND properties.
//
// Dates are always encoded in UTC. For all-day events (AllDay=true), the
// date-only format YYYYMMDD is used. For timed events, the full
// YYYYMMDDTHHMMSSZ format is used.
//
// Example encoded output:
//
//	BEGIN:VEVENT\r\nSUMMARY:Team Standup\r\nLOCATION:Room 42\r\nDTSTART:20250715T090000Z\r\nDTEND:20250715T100000Z\r\nEND:VEVENT
type CalendarPayload struct {
	// Title is the event summary.
	Title string
	// Description is the event description.
	Description string
	// Location is the event location.
	Location string
	// Start is the event start time.
	Start time.Time
	// End is the event end time.
	End time.Time
	// AllDay indicates whether this is an all-day event.
	AllDay bool
}

// Encode returns an iCalendar VEVENT string with CRLF line endings.
// Date-time values are in UTC using YYYYMMDDTHHMMSSZ format, or
// YYYYMMDD for all-day events.
func (c *CalendarPayload) Encode() (string, error) {
	if err := c.Validate(); err != nil {
		return "", err
	}
	var b strings.Builder
	b.WriteString("BEGIN:VEVENT\r\n")
	fmt.Fprintf(&b, "SUMMARY:%s\r\n", c.Title)
	if c.Description != "" {
		fmt.Fprintf(&b, "DESCRIPTION:%s\r\n", c.Description)
	}
	if c.Location != "" {
		fmt.Fprintf(&b, "LOCATION:%s\r\n", c.Location)
	}
	if c.AllDay {
		fmt.Fprintf(&b, "DTSTART:%s\r\n", c.Start.UTC().Format(dateLayout))
		fmt.Fprintf(&b, "DTEND:%s\r\n", c.End.UTC().Format(dateLayout))
	} else {
		fmt.Fprintf(&b, "DTSTART:%s\r\n", c.Start.UTC().Format(dateTimeLayout))
		fmt.Fprintf(&b, "DTEND:%s\r\n", c.End.UTC().Format(dateTimeLayout))
	}
	b.WriteString("END:VEVENT")
	return b.String(), nil
}

// Validate checks that the title is non-empty and that the end time
// is strictly after the start time.
func (c *CalendarPayload) Validate() error {
	if c.Title == "" {
		return fmt.Errorf("calendar payload: title must not be empty")
	}
	if !c.End.After(c.Start) {
		return fmt.Errorf("calendar payload: end time must be after start time")
	}
	return nil
}

// Type returns "calendar".
func (c *CalendarPayload) Type() string {
	return "calendar"
}

// Size returns the byte length of the encoded iCalendar string.
func (c *CalendarPayload) Size() int {
	encoded, _ := c.Encode() //nolint:errcheck // Size returns 0 on encode error
	return len(encoded)
}

const (
	dateTimeLayout = "20060102T150405Z"
	dateLayout     = "20060102"
)
