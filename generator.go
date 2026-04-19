package qrcode

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/os-gomod/qrcode/encoding"
	qrerrors "github.com/os-gomod/qrcode/errors"
	"github.com/os-gomod/qrcode/internal/hash"
	"github.com/os-gomod/qrcode/internal/lifecycle"
	"github.com/os-gomod/qrcode/internal/pool"
	"github.com/os-gomod/qrcode/internal/singleflight"
	"github.com/os-gomod/qrcode/payload"
	"github.com/os-gomod/qrcode/renderer"
)

type generator struct {
	mu         sync.RWMutex
	config     *Config
	lifecycle  *lifecycle.Guard
	sf         *singleflight.Group
	bufferPool *pool.BufferPool
}

//nolint:unparam // error return reserved for future initialization logic
func newGenerator(cfg *Config) (*generator, error) {
	g := &generator{
		config:     cfg,
		lifecycle:  lifecycle.New(),
		sf:         singleflight.NewGroup(),
		bufferPool: pool.NewBufferPool(),
	}
	return g, nil
}

func (g *generator) Generate(ctx context.Context, p payload.Payload) (*encoding.QRCode, error) {
	g.mu.RLock()
	cfg := g.config.Clone()
	g.mu.RUnlock()
	return g.generate(ctx, p, cfg)
}

func (g *generator) generate(_ context.Context, p payload.Payload, cfg *Config) (*encoding.QRCode, error) {
	if g.lifecycle.IsClosed() {
		return nil, qrerrors.New(qrerrors.ErrCodeClosed, "generator is closed")
	}
	data, err := p.Encode()
	if err != nil {
		return nil, qrerrors.Wrap(qrerrors.ErrCodePayload, "payload encode failed", err)
	}
	if validateErr := p.Validate(); validateErr != nil {
		return nil, qrerrors.Wrap(qrerrors.ErrCodePayload, "payload validation failed", validateErr)
	}
	ecLevel := parseECLevel(cfg.DefaultECLevel)
	if ecLevel < 0 {
		ecLevel = 1
	}
	version := cfg.DefaultVersion
	if version == 0 && cfg.AutoSize {
		version = 0
	}
	cacheKey := fmt.Sprintf("%d", hash.Hash(data))
	val, _, sfErr := g.sf.Do(cacheKey, func() (any, error) {
		return encoding.Encode([]byte(data), ecLevel, version)
	})
	if sfErr != nil {
		return nil, qrerrors.Wrap(qrerrors.ErrCodeEncoding, "QR encoding failed", sfErr)
	}
	return val.(*encoding.QRCode), nil //nolint:errcheck // singleflight.Do always returns the expected type
}

func (g *generator) GenerateWithOptions(ctx context.Context, p payload.Payload, opts ...Option) (*encoding.QRCode, error) {
	g.mu.RLock()
	cfg := g.config.Clone()
	g.mu.RUnlock()
	for _, opt := range opts {
		opt(cfg)
	}
	if err := cfg.Validate(); err != nil {
		return nil, qrerrors.Wrap(qrerrors.ErrCodeValidation, "invalid per-call options", err)
	}
	return g.generate(ctx, p, cfg)
}

func (g *generator) GenerateToWriter(ctx context.Context, p payload.Payload, w io.Writer, format Format) error {
	g.mu.RLock()
	renderOpts := []renderer.RenderOption{
		renderer.WithWidth(g.config.DefaultSize),
		renderer.WithHeight(g.config.DefaultSize),
		renderer.WithQuietZone(g.config.QuietZone),
		renderer.WithForegroundColor(g.config.ForegroundColor),
		renderer.WithBackgroundColor(g.config.BackgroundColor),
	}
	g.mu.RUnlock()
	qr, err := g.Generate(ctx, p)
	if err != nil {
		return err
	}
	return renderToWriter(ctx, qr, w, format, renderOpts)
}

func renderToWriter(ctx context.Context, qr *encoding.QRCode, w io.Writer, format Format, opts []renderer.RenderOption) error {
	switch format {
	case FormatPNG:
		r := renderer.NewPNGRenderer()
		return r.Render(ctx, qr, w, opts...)
	case FormatSVG:
		r := renderer.NewSVGRenderer()
		return r.Render(ctx, qr, w, opts...)
	case FormatTerminal:
		r := renderer.NewTerminalRenderer()
		return r.Render(ctx, qr, w, opts...)
	case FormatPDF:
		r := renderer.NewPDFRenderer()
		return r.Render(ctx, qr, w, opts...)
	case FormatBase64:
		r := renderer.NewBase64Renderer()
		return r.Render(ctx, qr, w, opts...)
	default:
		return qrerrors.New(qrerrors.ErrCodeRendering, fmt.Sprintf("unsupported format: %v", format))
	}
}

func (g *generator) Batch(ctx context.Context, payloads []payload.Payload, opts ...Option) ([]*encoding.QRCode, error) {
	if len(payloads) == 0 {
		return nil, nil
	}
	results := make([]*encoding.QRCode, len(payloads))
	errs := make([]error, len(payloads))
	var wg sync.WaitGroup
	wg.Add(len(payloads))
	for i, p := range payloads {
		go func(idx int, pl payload.Payload) {
			defer wg.Done()
			qr, err := g.GenerateWithOptions(ctx, pl, opts...)
			results[idx] = qr
			errs[idx] = err
		}(i, p)
	}
	wg.Wait()
	batchErr := qrerrors.NewBatchError()
	for i, err := range errs {
		if err != nil {
			batchErr.Errors[i] = err
		}
	}
	if len(batchErr.Errors) == 0 {
		return results, nil
	}
	return results, qrerrors.Wrap(qrerrors.ErrCodeBatch, "batch generation completed with errors", batchErr)
}

func (g *generator) Close(_ context.Context) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if err := g.lifecycle.Close(); err != nil {
		return qrerrors.Wrap(qrerrors.ErrCodeClosed, "close failed", err)
	}
	return nil
}

func (g *generator) SetOptions(opts ...Option) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.lifecycle.IsClosed() {
		return qrerrors.New(qrerrors.ErrCodeClosed, "cannot set options on closed generator")
	}
	newCfg := g.config.Clone()
	for _, opt := range opts {
		opt(newCfg)
	}
	if err := newCfg.Validate(); err != nil {
		return qrerrors.Wrap(qrerrors.ErrCodeValidation, "invalid options for SetOptions", err)
	}
	g.config = newCfg
	return nil
}

func (g *generator) Closed() bool {
	return g.lifecycle.IsClosed()
}

var _ Generator = (*generator)(nil)
