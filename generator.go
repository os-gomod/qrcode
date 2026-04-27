package qrcode

import (
	"context"
	"io"
	"strconv"
	"sync"

	qrerrors "github.com/os-gomod/qrcode/v2/errors"
	"github.com/os-gomod/qrcode/v2/internal/encoding"
	"github.com/os-gomod/qrcode/v2/internal/hash"
	"github.com/os-gomod/qrcode/v2/internal/lifecycle"
	"github.com/os-gomod/qrcode/v2/internal/pool"
	"github.com/os-gomod/qrcode/v2/internal/renderer"
	"github.com/os-gomod/qrcode/v2/internal/singleflight"
	"github.com/os-gomod/qrcode/v2/internal/storage"
	"github.com/os-gomod/qrcode/v2/internal/workerpool"
	"github.com/os-gomod/qrcode/v2/payload"
)

type generator struct {
	mu         sync.RWMutex
	config     *Config
	lifecycle  *lifecycle.Guard
	sf         *singleflight.Group
	bufferPool *pool.BufferPool
	store      storage.Storage
}

func newGenerator(cfg *Config) *generator {
	return &generator{
		config:     cfg,
		lifecycle:  lifecycle.New(),
		sf:         singleflight.NewGroup(),
		bufferPool: pool.NewBufferPool(),
		store:      storage.NewFileSystem(),
	}
}

func (g *generator) Generate(ctx context.Context, p payload.Payload) (*encoding.QRCode, error) {
	g.mu.RLock()
	cfg := g.config.Clone()
	g.mu.RUnlock()
	return g.generate(ctx, p, cfg)
}

func (g *generator) generate(ctx context.Context, p payload.Payload, cfg *Config) (*encoding.QRCode, error) {
	if g.lifecycle.IsClosed() {
		return nil, qrerrors.New(qrerrors.ErrCodeClosed, "client is closed")
	}
	select {
	case <-ctx.Done():
		return nil, qrerrors.Wrap(qrerrors.ErrCodeClosed, "context cancelled", ctx.Err())
	default:
	}
	if validateErr := p.Validate(); validateErr != nil {
		return nil, qrerrors.Wrap(qrerrors.ErrCodePayload, "payload validation failed", validateErr)
	}
	data, err := p.Encode()
	if err != nil {
		return nil, qrerrors.Wrap(qrerrors.ErrCodePayload, "payload encode failed", err)
	}
	ecLevel := parseECLevel(cfg.DefaultECLevel)
	if ecLevel < 0 {
		ecLevel = 1
	}
	cacheKey := strconv.FormatUint(hash.Hash(data), 10)
	val, _, sfErr := g.sf.Do(cacheKey, func() (any, error) {
		return encoding.Encode([]byte(data), ecLevel, cfg.DefaultVersion)
	})
	if sfErr != nil {
		return nil, qrerrors.Wrap(qrerrors.ErrCodeEncoding, "QR encoding failed", sfErr)
	}
	return val.(*encoding.QRCode), nil
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
	data, err := g.Render(ctx, p, format)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

// Render encodes a payload and returns the rendered output as bytes in the given format.
func (g *generator) Render(ctx context.Context, p payload.Payload, format Format) ([]byte, error) {
	qr, err := g.Generate(ctx, p)
	if err != nil {
		return nil, err
	}

	g.mu.RLock()
	renderOpts := []renderer.RenderOption{
		renderer.WithWidth(g.config.DefaultSize),
		renderer.WithHeight(g.config.DefaultSize),
		renderer.WithQuietZone(g.config.QuietZone),
		renderer.WithForegroundColor(g.config.ForegroundColor),
		renderer.WithBackgroundColor(g.config.BackgroundColor),
	}
	g.mu.RUnlock()

	r, err := renderer.GetRenderer(renderer.Format(format))
	if err != nil {
		return nil, qrerrors.Wrap(qrerrors.ErrCodeRendering, "renderer lookup failed", err)
	}
	return r.Render(ctx, qr, renderOpts...)
}

// Save encodes a payload, renders it to the format inferred from the file extension,
// and writes the result to the file system using the configured storage backend.
func (g *generator) Save(ctx context.Context, p payload.Payload, path string) error {
	format := FormatFromPath(path)
	data, err := g.Render(ctx, p, format)
	if err != nil {
		return err
	}
	return g.store.Save(ctx, path, data, 0o644)
}

func (g *generator) Batch(ctx context.Context, payloads []payload.Payload, opts ...Option) ([]*encoding.QRCode, error) {
	if len(payloads) == 0 {
		return nil, nil
	}
	g.mu.RLock()
	workers := g.config.WorkerCount
	g.mu.RUnlock()

	type indexedPayload struct {
		idx int
		pl  payload.Payload
	}
	type batchResult struct {
		qr *encoding.QRCode
	}
	jobs := make([]indexedPayload, len(payloads))
	for i, p := range payloads {
		jobs[i] = indexedPayload{idx: i, pl: p}
	}

	wp := workerpool.New[indexedPayload, batchResult](workers)
	poolResults, _ := wp.Process(ctx, jobs, func(ctx context.Context, job indexedPayload) (batchResult, error) {
		qr, err := g.GenerateWithOptions(ctx, job.pl, opts...)
		return batchResult{qr: qr}, err
	})

	results := make([]*encoding.QRCode, len(poolResults))
	batchErr := qrerrors.NewBatchError(len(poolResults))
	for i, pr := range poolResults {
		results[i] = pr.Value.qr
		if pr.Err != nil {
			batchErr.Errors[i] = pr.Err
		}
	}
	if len(batchErr.Errors) == 0 {
		return results, nil
	}
	return results, qrerrors.Wrap(qrerrors.ErrCodeBatch, "batch generation completed with errors", batchErr)
}

func (g *generator) Close() error {
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
		return qrerrors.New(qrerrors.ErrCodeClosed, "cannot set options on closed client")
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

// Compile-time interface satisfaction checks.
var _ Client = (*generator)(nil)
