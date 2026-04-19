// Package errors provides structured error types for the QR code library.
//
// All errors produced by the library implement the [QRCodeError] type,
// which carries a machine-readable [ErrorCode] and supports wrapping
// underlying causes. Use [IsCode] to programmatically inspect errors
// and [As] to extract the full [QRCodeError] value.
//
// # Error Codes
//
// The package defines a set of string-based error codes (ErrCode*) that
// classify failures into categories such as validation, encoding, rendering,
// batch processing, and I/O.
//
//	func handle(err error) {
//	    if errors.IsCode(err, errors.ErrCodeValidation) {
//	        // handle invalid input
//	    }
//	    if errors.IsCode(err, errors.ErrCodeDataTooLong) {
//	        // data exceeds QR code capacity
//	    }
//	}
package errors

import (
	"errors"
	"fmt"
)

// ErrorCode is a string-based error code for QR code errors.
//
// Error codes provide a stable, machine-readable classification of failures
// that can be used for programmatic error handling and logging.
type ErrorCode string

const (
	// ErrCodeUnknown is a generic or unclassified error.
	ErrCodeUnknown ErrorCode = "UNKNOWN"
	// ErrCodeValidation indicates input validation failure (empty data, invalid parameters, etc.).
	ErrCodeValidation ErrorCode = "VALIDATION"
	// ErrCodeEncoding indicates a failure during QR code data encoding.
	ErrCodeEncoding ErrorCode = "ENCODING"
	// ErrCodeRendering indicates a failure during image output rendering.
	ErrCodeRendering ErrorCode = "RENDERING"
	// ErrCodeTimeout indicates an operation exceeded its deadline.
	ErrCodeTimeout ErrorCode = "TIMEOUT"
	// ErrCodeClosed indicates an operation was attempted on a closed generator.
	ErrCodeClosed ErrorCode = "CLOSED"
	// ErrCodePayload indicates an invalid or unsupported payload type.
	ErrCodePayload ErrorCode = "PAYLOAD"
	// ErrCodeBatch indicates an error occurred during batch processing.
	ErrCodeBatch ErrorCode = "BATCH"
	// ErrCodeDataTooLong indicates the input data exceeds the maximum QR code capacity.
	ErrCodeDataTooLong ErrorCode = "DATA_TOO_LONG"
	// ErrCodeFileWrite indicates a file I/O error during output.
	ErrCodeFileWrite ErrorCode = "FILE_WRITE"
)

// QRCodeError is a typed error with an associated error code.
//
// It implements the error interface and supports cause chaining via
// [Unwrap]. The optional Meta map holds arbitrary key-value pairs for
// structured logging and diagnostics.
type QRCodeError struct {
	// Code is the machine-readable error classification.
	Code ErrorCode
	// Message is a human-readable description of the error.
	Message string
	// Cause is the underlying error that triggered this error (may be nil).
	Cause error
	// Meta holds optional key-value diagnostic metadata.
	Meta map[string]any
}

func (e *QRCodeError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *QRCodeError) Unwrap() error {
	return e.Cause
}

// WithMeta returns a copy of the error with the given key-value pair added to
// the Meta map. The original error is not modified.
func (e *QRCodeError) WithMeta(key string, value any) *QRCodeError {
	if e.Meta == nil {
		e.Meta = make(map[string]any)
	}
	ne := *e
	ne.Meta = make(map[string]any, len(e.Meta)+1)
	for k, v := range e.Meta {
		ne.Meta[k] = v
	}
	ne.Meta[key] = value
	return &ne
}

// New creates a new [QRCodeError] with the given code and message.
// Use [Wrap] to attach an underlying cause error.
func New(code ErrorCode, message string) *QRCodeError {
	return &QRCodeError{
		Code:    code,
		Message: message,
	}
}

// Wrap creates a [QRCodeError] wrapping an underlying cause.
// The cause is accessible via [errors.Unwrap] and [QRCodeError.Unwrap].
func Wrap(code ErrorCode, message string, cause error) *QRCodeError {
	return &QRCodeError{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

// IsCode reports whether err (or any error in its chain) is a
// [QRCodeError] with the given error code.
func IsCode(err error, code ErrorCode) bool {
	var qe *QRCodeError
	if As(err, &qe) {
		return qe.Code == code
	}
	return false
}

// As attempts to unwrap err into a *[QRCodeError] using [errors.As].
// Returns true if a matching error was found.
func As(err error, target **QRCodeError) bool {
	return errors.As(err, target)
}

// JoinErrors combines multiple errors into one. Nil errors are filtered out.
// If no non-nil errors remain, it returns nil. If only one remains, it is
// returned directly. Otherwise, the errors are joined into a single message.
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
	return fmt.Errorf("%v (and %d more errors)", nonNil[0], len(nonNil)-1)
}

// BatchError aggregates errors from a batch of QR code generations.
//
// The Errors map keys are the indices of the items that failed.
type BatchError struct {
	// Errors maps item indices to their corresponding errors.
	Errors map[int]error
}

func (be *BatchError) Error() string {
	return fmt.Sprintf("batch operation failed: %d of %d items had errors", len(be.Errors), len(be.Errors))
}

// NewBatchError creates a new empty [BatchError] ready to accumulate
// per-item errors by index.
func NewBatchError() *BatchError {
	return &BatchError{
		Errors: make(map[int]error),
	}
}
