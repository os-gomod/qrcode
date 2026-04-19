package qrcode

import "context"

// contextKey is the unexported type used for context value keys.
type contextKey struct{}

// ContextWithQR stores a Generator in the given context using an unexported
// context key. The stored Generator can later be retrieved with QRFromContext.
// This is useful in HTTP middleware or request-scoped dependency injection.
func ContextWithQR(ctx context.Context, gen Generator) context.Context {
	return context.WithValue(ctx, contextKey{}, gen)
}

// QRFromContext retrieves a Generator previously stored via ContextWithQR.
// Returns the Generator and true if found, or nil and false otherwise.
func QRFromContext(ctx context.Context) (Generator, bool) {
	gen, ok := ctx.Value(contextKey{}).(Generator)
	return gen, ok
}
