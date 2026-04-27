package qrcode

import (
	"context"
	"io"

	qrerrors "github.com/os-gomod/qrcode/v2/errors"
	"github.com/os-gomod/qrcode/v2/internal/encoding"
	"github.com/os-gomod/qrcode/v2/payload"
)

// ---------------------------------------------------------------------------
// Client — the canonical public interface for QR code generation.
// ---------------------------------------------------------------------------

// Client is the primary interface for QR code generation and rendering.
// It is the stable public contract that all implementations must satisfy.
// Use New() or NewClient() to create instances; avoid depending on the
// concrete generator type directly.
//
// All methods accept context.Context as the first parameter, enabling
// cancellation, timeout, and observability propagation.
type Client interface {
	// Generate encodes a payload into a QR code matrix without rendering.
	// Returns the raw QR code structure (modules, version, size).
	Generate(ctx context.Context, p payload.Payload) (*encoding.QRCode, error)

	// GenerateWithOptions encodes a payload with per-call option overrides.
	// Options are applied to a copy of the client's config — the client's
	// default configuration is not modified.
	GenerateWithOptions(ctx context.Context, p payload.Payload, opts ...Option) (*encoding.QRCode, error)

	// GenerateToWriter encodes a payload, renders it, and writes the result
	// to w in the specified format. This is the streaming variant of Render
	// for callers who want to pipe output directly (e.g., HTTP responses).
	GenerateToWriter(ctx context.Context, p payload.Payload, w io.Writer, format Format) error

	// Render encodes a payload and returns the rendered output as bytes.
	// This is the unified replacement for the former GeneratePNG/SVG/ASCII/Base64 functions.
	Render(ctx context.Context, p payload.Payload, format Format) ([]byte, error)

	// Save encodes a payload, renders it, and writes the result to a file.
	// The output format is inferred from the file extension (.png, .svg, .pdf, .txt, .b64).
	Save(ctx context.Context, p payload.Payload, path string) error

	// Batch generates multiple QR codes concurrently with bounded worker count.
	// Results are returned in the same order as the input payloads.
	// Optional per-call options are applied uniformly to all items.
	Batch(ctx context.Context, payloads []payload.Payload, opts ...Option) ([]*encoding.QRCode, error)

	// Close releases all resources held by the client.
	// After Close is called, all other methods return ErrCodeClosed errors.
	Close() error

	// SetOptions updates the client's default configuration.
	// The new configuration is validated before being applied.
	// This is safe for concurrent use.
	SetOptions(opts ...Option) error

	// Closed reports whether the client has been closed.
	Closed() bool
}

// ---------------------------------------------------------------------------
// Constructors
// ---------------------------------------------------------------------------

// New creates a new Client with the given options.
// It validates the configuration before returning; an error is returned
// for invalid configurations (e.g., size outside 100-4000, worker count
// outside 1-64).
//
// This is the primary constructor for the QR code library.
func New(opts ...Option) (Client, error) {
	cfg := defaultConfig()
	for _, opt := range opts {
		opt(cfg)
	}
	if err := cfg.Validate(); err != nil {
		return nil, qrerrors.Wrap(qrerrors.ErrCodeValidation, "invalid configuration", err)
	}
	return newGenerator(cfg), nil
}

// MustNew creates a new Client, panicking on invalid configuration.
// Use this only in package-level init functions or tests where failure is fatal.
func MustNew(opts ...Option) Client {
	client, err := New(opts...)
	if err != nil {
		panic("qrcode: MustNew: " + err.Error())
	}
	return client
}

// NewClient creates a new Client with the given options.
// This is the recommended constructor name for v2 code.
func NewClient(opts ...Option) (Client, error) {
	return New(opts...)
}

// MustNewClient creates a new Client, panicking on invalid configuration.
func MustNewClient(opts ...Option) Client {
	return MustNew(opts...)
}

// ---------------------------------------------------------------------------
// Context helpers — store and retrieve a Client from context.Context
// ---------------------------------------------------------------------------

type contextKey struct{}

// ContextWithQR stores a Client in a context value.
func ContextWithQR(ctx context.Context, client Client) context.Context {
	return context.WithValue(ctx, contextKey{}, client)
}

// QRFromContext retrieves a Client from a context value.
func QRFromContext(ctx context.Context) (Client, bool) {
	client, ok := ctx.Value(contextKey{}).(Client)
	return client, ok
}
