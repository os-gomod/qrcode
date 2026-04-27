package payload

import (
	"testing"
	"time"
)

func TestCalendarPayload(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		start := time.Date(2026, 6, 15, 10, 0, 0, 0, time.UTC)
		end := time.Date(2026, 6, 15, 11, 0, 0, 0, time.UTC)
		p := &CalendarPayload{Title: "Meeting", Start: start, End: end}
		enc, err := p.Encode()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p.Type() != "calendar" {
			t.Errorf("expected 'calendar', got %q", p.Type())
		}
		if !containsStr(enc, "BEGIN:VEVENT") {
			t.Error("should have BEGIN:VEVENT")
		}
		if !containsStr(enc, "SUMMARY:Meeting") {
			t.Error("should have SUMMARY:Meeting")
		}
		if !containsStr(enc, "END:VEVENT") {
			t.Error("should have END:VEVENT")
		}
	})
	t.Run("with description and location", func(t *testing.T) {
		start := time.Date(2026, 6, 15, 10, 0, 0, 0, time.UTC)
		end := time.Date(2026, 6, 15, 11, 0, 0, 0, time.UTC)
		p := &CalendarPayload{Title: "Conf", Description: "Desc", Location: "Room 1", Start: start, End: end}
		enc, _ := p.Encode()
		if !containsStr(enc, "DESCRIPTION:Desc") || !containsStr(enc, "LOCATION:Room 1") {
			t.Error("should include description and location")
		}
	})
	t.Run("all day", func(t *testing.T) {
		start := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)
		end := time.Date(2026, 6, 16, 0, 0, 0, 0, time.UTC)
		p := &CalendarPayload{Title: "Conf", Start: start, End: end, AllDay: true}
		enc, _ := p.Encode()
		if !containsStr(enc, "DTSTART:20260615") {
			t.Error("all-day event should use date format")
		}
	})
	t.Run("empty title", func(t *testing.T) {
		p := &CalendarPayload{Start: time.Now(), End: time.Now().Add(time.Hour)}
		if err := p.Validate(); err == nil {
			t.Error("expected error for empty title")
		}
	})
	t.Run("end before start", func(t *testing.T) {
		start := time.Date(2026, 6, 15, 12, 0, 0, 0, time.UTC)
		end := time.Date(2026, 6, 15, 10, 0, 0, 0, time.UTC)
		p := &CalendarPayload{Title: "T", Start: start, End: end}
		if err := p.Validate(); err == nil {
			t.Error("expected error for end before start")
		}
	})
}

func TestPhonePayload(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		p := &PhonePayload{Number: "+1234567890"}
		enc, err := p.Encode()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if enc != "tel:+1234567890" {
			t.Errorf("got %q", enc)
		}
		if p.Type() != "phone" {
			t.Errorf("expected 'phone', got %q", p.Type())
		}
	})
	t.Run("empty number", func(t *testing.T) {
		p := &PhonePayload{Number: ""}
		if err := p.Validate(); err == nil {
			t.Error("expected error for empty number")
		}
	})
	t.Run("no digits", func(t *testing.T) {
		p := &PhonePayload{Number: "abc"}
		if err := p.Validate(); err == nil {
			t.Error("expected error for number without digits")
		}
	})
}

func TestMMSPayload(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		p := &MMSPayload{Phone: "+1234", Subject: "Hi", Message: "Hello"}
		enc, err := p.Encode()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p.Type() != "mms" {
			t.Errorf("expected 'mms', got %q", p.Type())
		}
		if !containsStr(enc, "mms:+1234") {
			t.Errorf("got %q", enc)
		}
	})
	t.Run("phone only", func(t *testing.T) {
		p := &MMSPayload{Phone: "+1234567890"}
		enc, _ := p.Encode()
		if enc != "mms:+1234567890" {
			t.Errorf("got %q", enc)
		}
	})
	t.Run("empty phone", func(t *testing.T) {
		p := &MMSPayload{Phone: ""}
		if err := p.Validate(); err == nil {
			t.Error("expected error for empty phone")
		}
	})
	t.Run("no digits", func(t *testing.T) {
		p := &MMSPayload{Phone: "abc"}
		if err := p.Validate(); err == nil {
			t.Error("expected error for no digits")
		}
	})
}

func TestContainsDigit(t *testing.T) {
	if !containsDigit("123") {
		t.Error("should find digit")
	}
	if !containsDigit("abc1def") {
		t.Error("should find digit in string")
	}
	if containsDigit("abc") {
		t.Error("should not find digit")
	}
	if containsDigit("") {
		t.Error("empty string has no digits")
	}
}

func TestFormatCoord(t *testing.T) {
	c := formatCoord(37.7749)
	if c == "" {
		t.Error("formatCoord should not return empty")
	}
}
