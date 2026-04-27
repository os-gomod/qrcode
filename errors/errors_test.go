package errors

import (
	"errors"
	"fmt"
	"testing"
)

// ---------------------------------------------------------------------------
// DomainError interface tests
// ---------------------------------------------------------------------------

func TestQRCodeError_ImplementsDomainError(t *testing.T) {
	var _ DomainError = (*QRCodeError)(nil)
	var _ DomainError = New(ErrCodeValidation, "test")
}

func TestQRCodeError_Code(t *testing.T) {
	e := New(ErrCodeValidation, "test")
	if e.Code() != ErrCodeValidation {
		t.Errorf("expected %q, got %q", ErrCodeValidation, e.Code())
	}
}

func TestQRCodeError_Metadata_Nil(t *testing.T) {
	e := New(ErrCodeValidation, "test")
	if e.Metadata() != nil {
		t.Error("expected nil metadata for error without meta")
	}
}

func TestQRCodeError_Metadata_Copy(t *testing.T) {
	e := New(ErrCodeValidation, "test").WithMeta("key", "value")
	meta := e.Metadata()
	if meta["key"] != "value" {
		t.Errorf("expected key=value, got %v", meta["key"])
	}
	// Mutating the copy should not affect the original.
	meta["key"] = "modified"
	if e.Metadata()["key"] == "modified" {
		t.Error("metadata copy should be independent from original")
	}
}

// ---------------------------------------------------------------------------
// Retryable tests
// ---------------------------------------------------------------------------

func TestRetryable_DefaultCodes(t *testing.T) {
	tests := []struct {
		code ErrorCode
		want bool
	}{
		{ErrCodeTimeout, true},
		{ErrCodeInternal, true},
		{ErrCodeStorage, true},
		{ErrCodeValidation, false},
		{ErrCodeEncoding, false},
		{ErrCodeRendering, false},
		{ErrCodeClosed, false},
		{ErrCodePayload, false},
		{ErrCodeBatch, false},
		{ErrCodeDataTooLong, false},
		{ErrCodeFileWrite, false},
		{ErrCodeConfig, false},
		{ErrCodeUnknown, false},
	}
	for _, tt := range tests {
		t.Run(string(tt.code), func(t *testing.T) {
			e := New(tt.code, "test")
			if got := e.Retryable(); got != tt.want {
				t.Errorf("Retryable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRetryable_Override(t *testing.T) {
	e := New(ErrCodeValidation, "test").WithRetryable(true)
	if !e.Retryable() {
		t.Error("should be retryable after override")
	}

	e2 := New(ErrCodeTimeout, "test").WithRetryable(false)
	if e2.Retryable() {
		t.Error("should not be retryable after override")
	}
}

func TestIsRetryable(t *testing.T) {
	// Direct DomainError.
	e := New(ErrCodeTimeout, "test")
	if !IsRetryable(e) {
		t.Error("timeout errors should be retryable")
	}

	// Wrapped error should also be detected.
	inner := New(ErrCodeInternal, "inner")
	wrapped := Wrap(ErrCodeValidation, "wrapper", inner)
	if IsRetryable(wrapped) {
		t.Error("validation-wrapped internal error: outer code is validation (not retryable)")
	}
	// But IsRetryable checks the outermost DomainError, so validation is not retryable.

	// Non-DomainError should not be retryable.
	if IsRetryable(fmt.Errorf("plain error")) {
		t.Error("plain errors should not be retryable")
	}
}

func TestIsRetryable_WrappedChain(t *testing.T) {
	// The inner error is retryable but it's wrapped by a non-retryable outer.
	inner := New(ErrCodeTimeout, "timeout")
	wrapped := Wrap(ErrCodeValidation, "wrapper", inner)
	// IsRetryable should find the first DomainError in the chain (the outer one).
	if IsRetryable(wrapped) {
		t.Error("outer non-retryable should take precedence")
	}
}

// ---------------------------------------------------------------------------
// New / Wrap tests
// ---------------------------------------------------------------------------

func TestNew(t *testing.T) {
	e := New(ErrCodeValidation, "test error")
	if e.Code() != ErrCodeValidation {
		t.Errorf("expected code %q, got %q", ErrCodeValidation, e.Code())
	}
	if e.Message != "test error" {
		t.Errorf("expected message %q, got %q", "test error", e.Message)
	}
	if e.Cause != nil {
		t.Errorf("expected nil cause, got %v", e.Cause)
	}
}

func TestWrap(t *testing.T) {
	inner := fmt.Errorf("inner failure")
	e := Wrap(ErrCodeEncoding, "encoding failed", inner)

	if e.Code() != ErrCodeEncoding {
		t.Errorf("expected code %q, got %q", ErrCodeEncoding, e.Code())
	}
	if e.Message != "encoding failed" {
		t.Errorf("expected message %q, got %q", "encoding failed", e.Message)
	}
	if e.Cause != inner {
		t.Error("cause should point to inner error")
	}

	// Verify Error() includes both message and cause.
	errStr := e.Error()
	if !contains(errStr, "encoding failed") || !contains(errStr, "inner failure") {
		t.Errorf("Error() should include message and cause, got: %s", errStr)
	}
}

func TestWrap_NilCause(t *testing.T) {
	e := Wrap(ErrCodeRendering, "render failed", nil)
	if e.Cause != nil {
		t.Error("expected nil cause")
	}
	// Error() should omit cause portion.
	errStr := e.Error()
	if contains(errStr, ": ") && len(errStr) > len("[RENDERING] render failed") {
		t.Errorf("Error() with nil cause should not have ': <cause>' suffix, got: %s", errStr)
	}
}

func TestWrapf(t *testing.T) {
	e := Wrapf(ErrCodePayload, "invalid field %q", "email")
	if e.Message != `invalid field "email"` {
		t.Errorf("Message = %q", e.Message)
	}
}

func TestError_String(t *testing.T) {
	tests := []struct {
		name string
		err  *QRCodeError
		want string
	}{
		{
			name: "no cause",
			err:  New(ErrCodeValidation, "bad input"),
			want: "[VALIDATION] bad input",
		},
		{
			name: "with cause",
			err:  Wrap(ErrCodePayload, "encode fail", fmt.Errorf("data too long")),
			want: "[PAYLOAD] encode fail: data too long",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.want {
				t.Errorf("Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Unwrap / errors.Is / errors.As
// ---------------------------------------------------------------------------

func TestUnwrap(t *testing.T) {
	inner := fmt.Errorf("root cause")
	e := Wrap(ErrCodeClosed, "closed", inner)
	unwrapped := e.Unwrap()
	if unwrapped != inner {
		t.Error("Unwrap() should return the original cause")
	}

	// Nil cause returns nil.
	e2 := New(ErrCodeUnknown, "no cause")
	if e2.Unwrap() != nil {
		t.Error("Unwrap() on nil cause should return nil")
	}
}

func TestUnwrap_ErrorsIs(t *testing.T) {
	inner := fmt.Errorf("root cause")
	e := Wrap(ErrCodeUnknown, "wrapped", inner)
	if !errors.Is(e, inner) {
		t.Error("errors.Is should find the wrapped cause")
	}
}

func TestWithMeta(t *testing.T) {
	e := New(ErrCodeValidation, "base error")

	// WithMeta creates a new QRCodeError with a Meta map.
	e2 := e.WithMeta("key1", "value1")
	if e2.Meta == nil {
		t.Fatal("Meta map should not be nil after WithMeta")
	}
	if e2.Meta["key1"] != "value1" {
		t.Errorf("expected meta key1=value1, got %v", e2.Meta["key1"])
	}

	// Chained WithMeta should accumulate on the new copy.
	e3 := e2.WithMeta("key2", 42)
	if len(e3.Meta) != 2 {
		t.Errorf("expected 2 meta entries, got %d", len(e3.Meta))
	}
	if e3.Meta["key2"] != 42 {
		t.Errorf("expected meta key2=42, got %v", e3.Meta["key2"])
	}

	// e2 should not have key2 (independent copy).
	if _, ok := e2.Meta["key2"]; ok {
		t.Error("e2 should not be mutated by e3's WithMeta")
	}
}

func TestIsCode(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		code   ErrorCode
		expect bool
	}{
		{
			name:   "matching code",
			err:    New(ErrCodeValidation, "bad"),
			code:   ErrCodeValidation,
			expect: true,
		},
		{
			name:   "different code",
			err:    New(ErrCodeValidation, "bad"),
			code:   ErrCodeEncoding,
			expect: false,
		},
		{
			name:   "wrapped error",
			err:    Wrap(ErrCodeBatch, "batch fail", fmt.Errorf("item 3 failed")),
			code:   ErrCodeBatch,
			expect: true,
		},
		{
			name:   "non-QRCodeError",
			err:    fmt.Errorf("plain error"),
			code:   ErrCodeUnknown,
			expect: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsCode(tt.err, tt.code)
			if got != tt.expect {
				t.Errorf("IsCode() = %v, want %v", got, tt.expect)
			}
		})
	}
}

func TestIsCode_WrappedChain(t *testing.T) {
	// IsCode checks the first QRCodeError in the chain.
	inner := New(ErrCodeValidation, "inner")
	wrapped := Wrap(ErrCodeEncoding, "outer", inner)
	if IsCode(wrapped, ErrCodeEncoding) != true {
		t.Error("IsCode should find the outer error code")
	}
	if IsCode(wrapped, ErrCodeValidation) != false {
		t.Error("IsCode should not find the inner error code (outer takes precedence)")
	}
}

func TestAs(t *testing.T) {
	e := New(ErrCodeEncoding, "encoding error")
	var target *QRCodeError
	if !As(e, &target) {
		t.Error("As() should find QRCodeError")
	}
	if target.Code() != ErrCodeEncoding {
		t.Errorf("As() target code = %q, want %q", target.Code(), ErrCodeEncoding)
	}

	// Non-QRCodeError should return false.
	plain := fmt.Errorf("not a QRCodeError")
	if As(plain, &target) {
		t.Error("As() should return false for non-QRCodeError")
	}
}

// ---------------------------------------------------------------------------
// JoinErrors tests
// ---------------------------------------------------------------------------

func TestJoinErrors(t *testing.T) {
	t.Run("nil errors", func(t *testing.T) {
		result := JoinErrors()
		if result != nil {
			t.Errorf("expected nil, got %v", result)
		}
	})

	t.Run("all nil", func(t *testing.T) {
		result := JoinErrors(nil, nil, nil)
		if result != nil {
			t.Errorf("expected nil, got %v", result)
		}
	})

	t.Run("single error", func(t *testing.T) {
		err := fmt.Errorf("single")
		result := JoinErrors(err)
		if result != err {
			t.Errorf("expected same error, got %v", result)
		}
	})

	t.Run("multiple errors", func(t *testing.T) {
		e1 := fmt.Errorf("error 1")
		e2 := fmt.Errorf("error 2")
		result := JoinErrors(e1, e2)
		if result == nil {
			t.Fatal("expected non-nil result")
		}
		str := result.Error()
		if !contains(str, "error 1") || !contains(str, "1 more") {
			t.Errorf("expected summary format, got: %s", str)
		}
	})

	t.Run("mixed nil and error", func(t *testing.T) {
		e := fmt.Errorf("only one")
		result := JoinErrors(nil, e, nil)
		if result != e {
			t.Error("should return the single non-nil error")
		}
	})
}

// ---------------------------------------------------------------------------
// BatchError tests
// ---------------------------------------------------------------------------

func TestBatchError(t *testing.T) {
	t.Run("NewBatchError", func(t *testing.T) {
		be := NewBatchError(5)
		if be.Total != 5 {
			t.Errorf("expected total 5, got %d", be.Total)
		}
		if be.Errors == nil {
			t.Error("Errors map should not be nil")
		}
		if len(be.Errors) != 0 {
			t.Errorf("expected empty Errors map, got %d entries", len(be.Errors))
		}
	})

	t.Run("NewBatchError_zero_or_negative", func(t *testing.T) {
		be := NewBatchError(0)
		if be.Total != 0 {
			t.Errorf("expected total 0, got %d", be.Total)
		}
		be2 := NewBatchError(-3)
		if be2.Total != 0 {
			t.Errorf("expected total 0 for negative input, got %d", be2.Total)
		}
	})

	t.Run("Error_empty", func(t *testing.T) {
		be := NewBatchError(5)
		str := be.Error()
		if !contains(str, "0 of 5") {
			t.Errorf("expected '0 of 5', got: %s", str)
		}
	})

	t.Run("Error_with_failures", func(t *testing.T) {
		be := NewBatchError(10)
		be.Errors[2] = fmt.Errorf("item 2 failed")
		be.Errors[7] = fmt.Errorf("item 7 failed")
		str := be.Error()
		if !contains(str, "2 of 10") {
			t.Errorf("expected '2 of 10', got: %s", str)
		}
	})
}

func TestBatchError_SetGet(t *testing.T) {
	be := NewBatchError(5)
	be.Set(0, fmt.Errorf("err0"))
	be.Set(3, fmt.Errorf("err3"))

	if got := be.Get(0); got == nil || got.Error() != "err0" {
		t.Errorf("Get(0) = %v, want err0", got)
	}
	if got := be.Get(1); got != nil {
		t.Errorf("Get(1) = %v, want nil", got)
	}
	if be.Len() != 2 {
		t.Errorf("Len() = %d, want 2", be.Len())
	}
}

// ---------------------------------------------------------------------------
// Sentinel error tests
// ---------------------------------------------------------------------------

func TestSentinelErrors(t *testing.T) {
	sentinels := []struct {
		name string
		err  *QRCodeError
		code ErrorCode
	}{
		{"ErrClosed", ErrClosed, ErrCodeClosed},
		{"ErrDataTooLong", ErrDataTooLong, ErrCodeDataTooLong},
		{"ErrInvalidConfig", ErrInvalidConfig, ErrCodeConfig},
		{"ErrNilPayload", ErrNilPayload, ErrCodePayload},
	}
	for _, s := range sentinels {
		t.Run(s.name, func(t *testing.T) {
			if s.err.Code() != s.code {
				t.Errorf("%s.Code() = %q, want %q", s.name, s.err.Code(), s.code)
			}
			if s.err.Error() == "" {
				t.Errorf("%s.Error() should not be empty", s.name)
			}
		})
	}
}

func TestSentinel_ErrorsIs(t *testing.T) {
	// Sentinel errors should work with errors.Is.
	wrapped := Wrap(ErrCodeClosed, "wrapped closed", ErrClosed)
	if !errors.Is(wrapped, ErrClosed) {
		t.Error("errors.Is should find sentinel error in chain")
	}
}

// ---------------------------------------------------------------------------
// HTTP status mapping tests
// ---------------------------------------------------------------------------

func TestHTTPStatus_QRCodeError(t *testing.T) {
	tests := []struct {
		code ErrorCode
		want int
	}{
		{ErrCodeValidation, 400},
		{ErrCodeEncoding, 422},
		{ErrCodeTimeout, 504},
		{ErrCodeClosed, 503},
		{ErrCodeFileWrite, 500},
		{ErrCodeConfig, 400},
		{ErrCodeBatch, 207},
	}
	for _, tt := range tests {
		t.Run(string(tt.code), func(t *testing.T) {
			e := New(tt.code, "test")
			if got := e.HTTPStatus(); got != tt.want {
				t.Errorf("HTTPStatus() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestHTTPStatus_Helper(t *testing.T) {
	// DomainError.
	if got := HTTPStatus(New(ErrCodeValidation, "test")); got != 400 {
		t.Errorf("HTTPStatus = %d, want 400", got)
	}
	// Plain error.
	if got := HTTPStatus(fmt.Errorf("plain")); got != 500 {
		t.Errorf("HTTPStatus for plain error = %d, want 500", got)
	}
}

func TestHTTPStatus_UnknownCode(t *testing.T) {
	e := New(ErrorCode("CUSTOM_CODE"), "test")
	if got := e.HTTPStatus(); got != 500 {
		t.Errorf("HTTPStatus for unknown code = %d, want 500", got)
	}
}

// ---------------------------------------------------------------------------
// Recover / SafeExecute tests
// ---------------------------------------------------------------------------

func TestRecover_NoPanic(t *testing.T) {
	err := Recover()
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestSafeExecute(t *testing.T) {
	err := SafeExecute(func() error {
		return nil
	})
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestSafeExecute_Panic(t *testing.T) {
	err := SafeExecute(func() error {
		panic("test panic")
	})
	if err == nil {
		t.Fatal("expected error from panic")
	}
	if !IsCode(err, ErrCodeInternal) {
		t.Errorf("expected ErrCodeInternal, got %v", err)
	}
}

func TestSafeExecute_PanicError(t *testing.T) {
	innerErr := fmt.Errorf("inner panic")
	err := SafeExecute(func() error {
		panic(innerErr)
	})
	if err == nil {
		t.Fatal("expected error from panic")
	}
	if !IsCode(err, ErrCodeInternal) {
		t.Errorf("expected ErrCodeInternal, got %v", err)
	}
	// The cause chain should include the original error.
	if !errors.Is(err, innerErr) {
		t.Error("errors.Is should find the original panic error")
	}
}

// ---------------------------------------------------------------------------
// All error codes test
// ---------------------------------------------------------------------------

func TestAllErrorCodes(t *testing.T) {
	codes := []ErrorCode{
		ErrCodeUnknown,
		ErrCodeValidation,
		ErrCodeEncoding,
		ErrCodeRendering,
		ErrCodeTimeout,
		ErrCodeClosed,
		ErrCodePayload,
		ErrCodeBatch,
		ErrCodeDataTooLong,
		ErrCodeFileWrite,
		ErrCodeStorage,
		ErrCodeConfig,
		ErrCodeInternal,
	}
	for _, code := range codes {
		e := New(code, "test")
		if e.Code() != code {
			t.Errorf("code mismatch: expected %q, got %q", code, e.Code())
		}
	}
}

// ---------------------------------------------------------------------------
// Helper
// ---------------------------------------------------------------------------

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
