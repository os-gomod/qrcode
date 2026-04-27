// Package errors provides enterprise-grade error handling for the QR code library.
// All errors implement the DomainError interface, supporting structured error
// codes, retryability classification, metadata propagation, and HTTP status mapping.
//
// # Usage
//
//	// Create a new domain error
//	err := errors.New(errors.ErrCodeValidation, "invalid input")
//
//	// Wrap an existing error with context
//	err := errors.Wrap(errors.ErrCodeEncoding, "encode failed", innerErr)
//
//	// Check error codes
//	if errors.IsCode(err, errors.ErrCodeValidation) { ... }
//
//	// Add metadata
//	err = err.WithMeta("field", "email").WithMeta("value", "bad")
//
//	// Check retryability
//	if errors.IsRetryable(err) { ... }
package errors

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
)

// ---------------------------------------------------------------------------
// DomainError interface
// ---------------------------------------------------------------------------

// DomainError is the core error interface for all library errors.
// It extends the standard error interface with structured error classification.
type DomainError interface {
	error
	// Code returns the machine-readable error code.
	Code() ErrorCode
	// Retryable reports whether the operation that caused this error
	// can reasonably be retried. Examples of retryable errors include
	// timeouts, temporary resource exhaustion, and rate limiting.
	// Non-retryable errors include validation failures and encoding errors.
	Retryable() bool
	// Metadata returns a copy of the error's metadata map.
	// Returns nil if no metadata has been attached.
	Metadata() map[string]any
}

// ---------------------------------------------------------------------------
// Error codes
// ---------------------------------------------------------------------------

// ErrorCode is a machine-readable identifier for error categories.
type ErrorCode string

const (
	ErrCodeUnknown     ErrorCode = "UNKNOWN"
	ErrCodeValidation  ErrorCode = "VALIDATION"
	ErrCodeEncoding    ErrorCode = "ENCODING"
	ErrCodeRendering   ErrorCode = "RENDERING"
	ErrCodeTimeout     ErrorCode = "TIMEOUT"
	ErrCodeClosed      ErrorCode = "CLOSED"
	ErrCodePayload     ErrorCode = "PAYLOAD"
	ErrCodeBatch       ErrorCode = "BATCH"
	ErrCodeDataTooLong ErrorCode = "DATA_TOO_LONG"
	ErrCodeFileWrite   ErrorCode = "FILE_WRITE"
	ErrCodeStorage     ErrorCode = "STORAGE"
	ErrCodeConfig      ErrorCode = "CONFIG"
	ErrCodeInternal    ErrorCode = "INTERNAL"
)

// retryableCodes defines which error codes are considered retryable.
var retryableCodes = map[ErrorCode]bool{
	ErrCodeTimeout:  true,
	ErrCodeInternal: true,
	ErrCodeStorage:  true,
}

// httpStatusMapping defines default HTTP status codes for error categories.
var httpStatusMapping = map[ErrorCode]int{
	ErrCodeUnknown:     http.StatusInternalServerError,
	ErrCodeValidation:  http.StatusBadRequest,
	ErrCodeEncoding:    http.StatusUnprocessableEntity,
	ErrCodeRendering:   http.StatusUnprocessableEntity,
	ErrCodeTimeout:     http.StatusGatewayTimeout,
	ErrCodeClosed:      http.StatusServiceUnavailable,
	ErrCodePayload:     http.StatusBadRequest,
	ErrCodeBatch:       http.StatusMultiStatus,
	ErrCodeDataTooLong: http.StatusRequestEntityTooLarge,
	ErrCodeFileWrite:   http.StatusInternalServerError,
	ErrCodeStorage:     http.StatusServiceUnavailable,
	ErrCodeConfig:      http.StatusBadRequest,
	ErrCodeInternal:    http.StatusInternalServerError,
}

// ---------------------------------------------------------------------------
// QRCodeError — the primary domain error
// ---------------------------------------------------------------------------

// QRCodeError is the standard error type for the QR code library.
// It implements the DomainError interface and supports error wrapping
// via errors.Is / errors.As compatibility.
type QRCodeError struct {
	code    ErrorCode
	Message string
	Cause   error
	Meta    map[string]any

	// retryable overrides the default retryability for this error code.
	// When nil, retryability is determined from the code.
	retryable *bool
}

// Compile-time interface compliance check.
var _ DomainError = (*QRCodeError)(nil)

// Error returns a human-readable error string including the code, message,
// and optional cause chain.
func (e *QRCodeError) Error() string {
	var b strings.Builder
	b.WriteString("[")
	b.WriteString(string(e.code))
	b.WriteString("] ")
	b.WriteString(e.Message)
	if e.Cause != nil {
		b.WriteString(": ")
		b.WriteString(e.Cause.Error())
	}
	return b.String()
}

// Unwrap supports error chain traversal via errors.Is / errors.As.
func (e *QRCodeError) Unwrap() error {
	return e.Cause
}

// Code returns the error's category code.
func (e *QRCodeError) Code() ErrorCode {
	return e.code
}

// Retryable reports whether the operation can be retried.
// By default, TIMEOUT, INTERNAL, and STORAGE errors are retryable.
// Use WithRetryable() to override per-instance.
func (e *QRCodeError) Retryable() bool {
	if e.retryable != nil {
		return *e.retryable
	}
	return retryableCodes[e.code]
}

// Metadata returns a shallow copy of the error's metadata, or nil if none exists.
func (e *QRCodeError) Metadata() map[string]any {
	if e.Meta == nil {
		return nil
	}
	cp := make(map[string]any, len(e.Meta))
	for k, v := range e.Meta {
		cp[k] = v
	}
	return cp
}

// WithMeta attaches a key-value pair to the error's metadata.
// It returns a new QRCodeError with the metadata merged; the original is unchanged.
func (e *QRCodeError) WithMeta(key string, value any) *QRCodeError {
	ne := *e
	ne.Meta = make(map[string]any, len(e.Meta)+1)
	for k, v := range e.Meta {
		ne.Meta[k] = v
	}
	ne.Meta[key] = value
	return &ne
}

// WithRetryable sets a per-instance retryability override.
// This overrides the default retryability derived from the error code.
func (e *QRCodeError) WithRetryable(retryable bool) *QRCodeError {
	ne := *e
	ne.retryable = &retryable
	return &ne
}

// HTTPStatus returns the recommended HTTP status code for this error.
func (e *QRCodeError) HTTPStatus() int {
	if status, ok := httpStatusMapping[e.code]; ok {
		return status
	}
	return http.StatusInternalServerError
}

// ---------------------------------------------------------------------------
// Constructors
// ---------------------------------------------------------------------------

// New creates a new QRCodeError with the given code and message.
func New(code ErrorCode, message string) *QRCodeError {
	return &QRCodeError{
		code:    code,
		Message: message,
	}
}

// Wrap creates a new QRCodeError that wraps an existing cause error.
// If cause is nil, it behaves like New.
func Wrap(code ErrorCode, message string, cause error) *QRCodeError {
	return &QRCodeError{
		code:    code,
		Message: message,
		Cause:   cause,
	}
}

// Wrapf creates a new QRCodeError with a formatted message wrapping a cause.
func Wrapf(code ErrorCode, format string, args ...any) *QRCodeError {
	return &QRCodeError{
		code:    code,
		Message: fmt.Sprintf(format, args...),
	}
}

// ---------------------------------------------------------------------------
// Sentinel errors
// ---------------------------------------------------------------------------

var (
	// ErrClosed is returned when operations are attempted on a closed client.
	ErrClosed = New(ErrCodeClosed, "client is closed")

	// ErrDataTooLong is returned when the input data exceeds the maximum
	// capacity for the selected error correction level.
	ErrDataTooLong = New(ErrCodeDataTooLong, "data too long for QR code")

	// ErrInvalidConfig is returned when a configuration fails validation.
	ErrInvalidConfig = New(ErrCodeConfig, "invalid configuration")

	// ErrNilPayload is returned when a nil payload is provided.
	ErrNilPayload = New(ErrCodePayload, "payload is nil")
)

// ---------------------------------------------------------------------------
// Error classification helpers
// ---------------------------------------------------------------------------

// IsCode reports whether err (or any wrapped error in its chain) is a
// QRCodeError with the given code.
func IsCode(err error, code ErrorCode) bool {
	var qe *QRCodeError
	if errors.As(err, &qe) {
		return qe.code == code
	}
	return false
}

// As is a convenience wrapper around errors.As for QRCodeError.
func As(err error, target **QRCodeError) bool {
	return errors.As(err, target)
}

// IsRetryable reports whether the error (or any wrapped DomainError in its chain)
// is retryable.
func IsRetryable(err error) bool {
	var de DomainError
	if errors.As(err, &de) {
		return de.Retryable()
	}
	return false
}

// HTTPStatus returns the recommended HTTP status code for the given error.
// For non-DomainError errors, it returns 500 (Internal Server Error).
func HTTPStatus(err error) int {
	var qe *QRCodeError
	if errors.As(err, &qe) {
		return qe.HTTPStatus()
	}
	return http.StatusInternalServerError
}

// ---------------------------------------------------------------------------
// Error joining and aggregation
// ---------------------------------------------------------------------------

// JoinErrors combines multiple errors into a single error.
// Nil errors are filtered out. If only one non-nil error remains,
// it is returned directly. Otherwise, a joined error is returned.
func JoinErrors(errs ...error) error {
	var nonNil []error
	for _, e := range errs {
		if e != nil {
			nonNil = append(nonNil, e)
		}
	}
	if len(nonNil) == 0 {
		return nil
	}
	if len(nonNil) == 1 {
		return nonNil[0]
	}
	return fmt.Errorf("%w (and %d more errors)", nonNil[0], len(nonNil)-1)
}

// ---------------------------------------------------------------------------
// BatchError — aggregate error for batch operations
// ---------------------------------------------------------------------------

// BatchError collects errors from batch operations, mapping item indices
// to their respective errors.
type BatchError struct {
	Errors map[int]error
	Total  int

	mu sync.RWMutex
}

// Error returns a human-readable summary of the batch errors.
func (be *BatchError) Error() string {
	be.mu.RLock()
	defer be.mu.RUnlock()
	return fmt.Sprintf("batch operation failed: %d of %d items had errors", len(be.Errors), be.Total)
}

// Set records an error for the given item index.
func (be *BatchError) Set(idx int, err error) {
	be.mu.Lock()
	defer be.mu.Unlock()
	be.Errors[idx] = err
}

// Get retrieves the error for the given item index, or nil if none.
func (be *BatchError) Get(idx int) error {
	be.mu.RLock()
	defer be.mu.RUnlock()
	return be.Errors[idx]
}

// Len returns the number of failed items.
func (be *BatchError) Len() int {
	be.mu.RLock()
	defer be.mu.RUnlock()
	return len(be.Errors)
}

// NewBatchError creates a new BatchError for the given total item count.
// If total <= 0, it defaults to 0.
func NewBatchError(total int) *BatchError {
	if total <= 0 {
		total = 0
	}
	return &BatchError{
		Errors: make(map[int]error, total),
		Total:  total,
	}
}

// ---------------------------------------------------------------------------
// Safe error recovery
// ---------------------------------------------------------------------------

// Recover recovers from a panic and returns it as an error.
// Useful in goroutine wrappers to prevent panics from crashing the process.
//
//nolint:revive // recover is intentionally called directly as a utility function
func Recover() error {
	if r := recover(); r != nil {
		switch v := r.(type) {
		case error:
			return Wrap(ErrCodeInternal, "panic recovered", v)
		case string:
			return Wrap(ErrCodeInternal, "panic recovered", fmt.Errorf("%s", v))
		default:
			return Wrap(ErrCodeInternal, "panic recovered", fmt.Errorf("%v", v))
		}
	}
	return nil
}

// SafeExecute runs fn and converts any panic into an error.
func SafeExecute(fn func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			switch v := r.(type) {
			case error:
				err = Wrap(ErrCodeInternal, "panic in SafeExecute", v)
			default:
				err = Wrap(ErrCodeInternal, "panic in SafeExecute", fmt.Errorf("%v", v))
			}
		}
	}()
	return fn()
}
