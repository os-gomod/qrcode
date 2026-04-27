package payload

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

type CalendarPayload struct {
	Title       string
	Description string
	Location    string
	Start       time.Time
	End         time.Time
	AllDay      bool
}

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

func (c *CalendarPayload) Validate() error {
	if c.Title == "" {
		return errors.New("calendar payload: title must not be empty")
	}
	if !c.End.After(c.Start) {
		return errors.New("calendar payload: end time must be after start time")
	}
	return nil
}

func (*CalendarPayload) Type() string {
	return "calendar"
}

func (c *CalendarPayload) Size() int {
	encoded, _ := c.Encode()
	return len(encoded)
}

const (
	dateTimeLayout = "20060102T150405Z"
	dateLayout     = "20060102"
)
