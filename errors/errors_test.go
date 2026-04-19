package errors

import (
	"errors"
	"testing"
)

func TestNew(t *testing.T) {
	e := New(ErrCodeValidation, "test error")
	if e == nil {
		t.Fatal("New returned nil")
	}
	if e.Code != ErrCodeValidation {
		t.Errorf("Code = %q, want %q", e.Code, ErrCodeValidation)
	}
	if e.Message != "test error" {
		t.Errorf("Message = %q", e.Message)
	}
	if e.Cause != nil {
		t.Error("Cause should be nil")
	}
	want := "[VALIDATION] test error"
	if e.Error() != want {
		t.Errorf("Error() = %q, want %q", e.Error(), want)
	}
}

func TestWrap(t *testing.T) {
	inner := errors.New("inner error")
	e := Wrap(ErrCodeEncoding, "wrapper", inner)
	if e == nil {
		t.Fatal("Wrap returned nil")
	}
	if e.Code != ErrCodeEncoding {
		t.Errorf("Code = %q", e.Code)
	}
	if e.Cause != inner {
		t.Error("Cause should be inner error")
	}
	if !errors.Is(e, inner) {
		t.Error("errors.Is should find inner error")
	}
	if e.Unwrap() != inner {
		t.Error("Unwrap() should return inner")
	}
	msg := e.Error()
	if msg == "" {
		t.Error("Error() should return non-empty string")
	}
}

func TestIsCode(t *testing.T) {
	e := New(ErrCodeValidation, "validation failed")
	if !IsCode(e, ErrCodeValidation) {
		t.Error("IsCode should return true for matching code")
	}
	if IsCode(e, ErrCodeEncoding) {
		t.Error("IsCode should return false for non-matching code")
	}
	if IsCode(errors.New("plain error"), ErrCodeValidation) {
		t.Error("IsCode should return false for non-QRCodeError")
	}
	if IsCode(nil, ErrCodeValidation) {
		t.Error("IsCode should return false for nil")
	}
}

func TestAs(t *testing.T) {
	e := New(ErrCodeRendering, "render error")
	var target *QRCodeError
	if !As(e, &target) {
		t.Error("As should return true for QRCodeError")
	}
	if target.Code != ErrCodeRendering {
		t.Errorf("target.Code = %q, want %q", target.Code, ErrCodeRendering)
	}

	_ = As(errors.New("plain"), &target)
}

func TestJoinErrors(t *testing.T) {
	tests := []struct {
		name string
		errs []error
		want bool
	}{
		{name: "nil errors", errs: nil, want: false},
		{name: "empty errors", errs: []error{}, want: false},
		{name: "all nil", errs: []error{nil, nil}, want: false},
		{name: "single error", errs: []error{errors.New("e1")}, want: true},
		{name: "multiple errors", errs: []error{errors.New("e1"), errors.New("e2")}, want: true},
		{name: "mixed", errs: []error{nil, errors.New("e1"), nil}, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := JoinErrors(tt.errs...)
			if tt.want && got == nil {
				t.Error("expected non-nil error")
			}
			if !tt.want && got != nil {
				t.Errorf("expected nil error, got %v", got)
			}
		})
	}
}

func TestBatchError(t *testing.T) {
	be := NewBatchError()
	if be == nil {
		t.Fatal("NewBatchError returned nil")
	}
	if len(be.Errors) != 0 {
		t.Errorf("Errors should be empty, got %d", len(be.Errors))
	}
	be.Errors[0] = errors.New("first")
	be.Errors[2] = errors.New("third")
	msg := be.Error()
	if msg == "" {
		t.Error("Error() should return non-empty string")
	}
}

func TestWithMeta(t *testing.T) {
	e := New(ErrCodeValidation, "test")
	e2 := e.WithMeta("key1", "value1")
	if e2 == nil {
		t.Fatal("WithMeta returned nil")
	}
	if e2.Code != ErrCodeValidation {
		t.Errorf("Code = %q", e2.Code)
	}
	if e2.Meta == nil {
		t.Fatal("Meta should not be nil")
	}
	// Original gets Meta set (side effect of pointer receiver), but e2's Meta is independent
	if e2.Meta["key1"] != "value1" {
		t.Errorf("Meta[key1] = %v", e2.Meta["key1"])
	}
	// Verify e2's Meta is independent from e's Meta
	e2.Meta["different"] = "val"
	if e.Meta["different"] == "val" {
		t.Error("e2.Meta should be independent from e.Meta")
	}

	// Chain WithMeta
	e3 := e2.WithMeta("key2", 42)
	if len(e3.Meta) != 3 {
		t.Errorf("expected 3 meta entries, got %d", len(e3.Meta))
	}
	if e3.Meta["key2"] != 42 {
		t.Errorf("Meta[key2] = %v", e3.Meta["key2"])
	}
}

func TestErrorCodeConstants(t *testing.T) {
	codes := []ErrorCode{
		ErrCodeUnknown, ErrCodeValidation, ErrCodeEncoding,
		ErrCodeRendering, ErrCodeTimeout, ErrCodeClosed,
		ErrCodePayload, ErrCodeBatch, ErrCodeDataTooLong,
		ErrCodeFileWrite,
	}
	for _, code := range codes {
		if string(code) == "" {
			t.Errorf("error code constant is empty")
		}
	}
}
