package batch

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	qrcode "github.com/os-gomod/qrcode/v2"
	qrerrors "github.com/os-gomod/qrcode/v2/errors"
	"github.com/os-gomod/qrcode/v2/internal/encoding"
	"github.com/os-gomod/qrcode/v2/internal/storage"
	"github.com/os-gomod/qrcode/v2/internal/workerpool"
	"github.com/os-gomod/qrcode/v2/payload"
)

// Item represents a single input for batch QR code generation.
type Item struct {
	ID      string
	Data    string
	Payload payload.Payload
}

// Result holds the outcome of processing a single batch item.
type Result struct {
	ID     string
	QRCode *encoding.QRCode
	Data   []byte
	Err    error
	Path   string
}

// BatchStats holds aggregate statistics for a batch run.
type BatchStats struct {
	Total     int
	Succeeded int
	Failed    int
	TotalTime time.Duration
	AvgTime   time.Duration
	MinTime   time.Duration
	MaxTime   time.Duration
}

const defaultConcurrency = 4

// Processor orchestrates batch QR code generation with bounded concurrency.
// It uses the generic workerpool.WorkerPool interface internally, decoupled
// from the concrete pool implementation for testability and flexibility.
type Processor struct {
	gen          qrcode.Client
	concurrency  int
	format       qrcode.Format
	outputDir    string
	renderFormat bool
	store        storage.Storage
	poolFactory  func(int) workerpool.WorkerPool[Item, processResult]
}

// ProcessorOption configures a Processor.
type ProcessorOption func(*Processor)

// WithBatchConcurrency sets the maximum number of concurrent workers.
func WithBatchConcurrency(n int) ProcessorOption {
	return func(p *Processor) {
		if n > 0 {
			p.concurrency = n
		}
	}
}

// WithBatchFormat sets the output format and enables rendering.
func WithBatchFormat(f qrcode.Format) ProcessorOption {
	return func(p *Processor) {
		p.format = f
		p.renderFormat = true
	}
}

// WithBatchOutputDir sets the directory where rendered results are saved.
func WithBatchOutputDir(dir string) ProcessorOption {
	return func(p *Processor) {
		p.outputDir = dir
	}
}

// WithBatchStorage sets a custom storage backend for file output.
// If not set, the default is storage.NewFileSystem().
func WithBatchStorage(s storage.Storage) ProcessorOption {
	return func(p *Processor) {
		if s != nil {
			p.store = s
		}
	}
}

// WithBatchPoolFactory sets a custom worker pool factory for testing.
// This allows injecting mock pools that implement WorkerPool[Item, processResult].
func WithBatchPoolFactory(factory func(int) workerpool.WorkerPool[Item, processResult]) ProcessorOption {
	return func(p *Processor) {
		if factory != nil {
			p.poolFactory = factory
		}
	}
}

// NewProcessor creates a new batch Processor bound to the given Client.
func NewProcessor(gen qrcode.Client, opts ...ProcessorOption) *Processor {
	p := &Processor{
		gen:         gen,
		concurrency: defaultConcurrency,
		format:      qrcode.FormatPNG,
		store:       storage.NewFileSystem(),
	}
	// Default pool factory uses the production worker pool.
	p.poolFactory = func(workers int) workerpool.WorkerPool[Item, processResult] {
		return workerpool.New[Item, processResult](workers, workerpool.WithDurationTracking())
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

// processResult is the internal unit of work produced by the worker pool.
type processResult struct {
	ID     string
	QRCode *encoding.QRCode
	Data   []byte
	Path   string
}

// process processes all items using the bounded worker pool and returns
// ordered workerpool.Results. This is the single shared pipeline that all
// public methods (Process, ProcessWithStats, SaveToDir) delegate to.
func (p *Processor) process(ctx context.Context, items []Item, saveToDir string) ([]workerpool.Result[Item, processResult], error) {
	wp := p.poolFactory(p.concurrency)

	poolResults, err := wp.Process(ctx, items, func(ctx context.Context, item Item) (processResult, error) {
		pl := p.resolvePayload(item)
		qr, genErr := p.gen.Generate(ctx, pl)
		if genErr != nil {
			return processResult{ID: item.ID}, genErr
		}
		res := processResult{
			ID:     item.ID,
			QRCode: qr,
		}
		// Render to bytes if format is enabled.
		if p.renderFormat || saveToDir != "" {
			var buf bytes.Buffer
			if wErr := p.gen.GenerateToWriter(ctx, pl, &buf, p.format); wErr != nil {
				return processResult{ID: item.ID}, wErr
			}
			res.Data = buf.Bytes()
		}
		// Save to directory if requested.
		if saveToDir != "" && len(res.Data) > 0 {
			name := item.ID
			if name == "" {
				return processResult{ID: item.ID}, nil // caller fills name from index
			}
			ext := formatExtension(p.format)
			fp := filepath.Join(saveToDir, name+"."+ext)
			if wErr := p.store.Save(ctx, fp, res.Data, 0o644); wErr != nil {
				return processResult{ID: item.ID}, qrerrors.Wrap(qrerrors.ErrCodeFileWrite, "failed to write file", wErr)
			}
			res.Path = fp
		}
		return res, nil
	})

	return poolResults, err
}

// convertResults maps workerpool.Results to the public []Result type.
func convertResults(poolResults []workerpool.Result[Item, processResult], saveToDir string) []Result {
	results := make([]Result, len(poolResults))
	for i, pr := range poolResults {
		r := Result{
			ID:     pr.Value.ID,
			QRCode: pr.Value.QRCode,
			Data:   pr.Value.Data,
			Path:   pr.Value.Path,
		}
		if pr.Err != nil {
			r.Err = pr.Err
			// For processToDir, fill in the file path from index if ID was empty.
			if saveToDir != "" && r.ID == "" {
				ext := formatExtension(qrcode.FormatPNG)
				name := strconv.Itoa(i)
				fp := filepath.Join(saveToDir, name+"."+ext)
				r.Path = fp
			}
		}
		results[i] = r
	}
	return results
}

// Process generates QR codes for all items concurrently.
// Results are returned in the same order as the input items.
func (p *Processor) Process(ctx context.Context, items []Item) ([]Result, error) {
	if len(items) == 0 {
		return nil, nil
	}
	poolResults, _ := p.process(ctx, items, "")
	results := convertResults(poolResults, "")
	// Optionally save to configured output directory.
	if p.outputDir != "" {
		p.saveResults(ctx, results)
	}
	return results, buildBatchError(results)
}

// ProcessWithStats generates QR codes and returns per-item timing statistics.
func (p *Processor) ProcessWithStats(ctx context.Context, items []Item) ([]Result, *BatchStats, error) {
	batchStart := time.Now()
	if len(items) == 0 {
		return nil, &BatchStats{}, nil
	}
	poolResults, _ := p.process(ctx, items, "")
	results := convertResults(poolResults, "")
	if p.outputDir != "" {
		p.saveResults(ctx, results)
	}
	// Extract durations from pool results.
	durations := make([]time.Duration, len(poolResults))
	for i, pr := range poolResults {
		durations[i] = pr.Elapsed
	}
	stats := computeBatchStats(durations, results, time.Since(batchStart))
	return results, stats, buildBatchError(results)
}

// SaveToDir generates QR codes and saves each to a file in outputDir.
// If no format was configured, PNG is used by default.
func (p *Processor) SaveToDir(ctx context.Context, items []Item, outputDir string) ([]Result, error) {
	if len(items) == 0 {
		return nil, nil
	}
	origFormat := p.format
	origRenderFormat := p.renderFormat
	if !p.renderFormat {
		p.format = qrcode.FormatPNG
		p.renderFormat = true
	}
	poolResults, _ := p.process(ctx, items, outputDir)
	results := convertResults(poolResults, outputDir)
	p.format = origFormat
	p.renderFormat = origRenderFormat
	return results, buildBatchError(results)
}

// ---------------------------------------------------------------------------
// Input parsers
// ---------------------------------------------------------------------------

// FromJSON reads batch items from a JSON array.
func (*Processor) FromJSON(_ context.Context, reader io.Reader) ([]Item, error) {
	var raw []jsonItem
	if err := json.NewDecoder(reader).Decode(&raw); err != nil {
		return nil, qrerrors.Wrap(qrerrors.ErrCodeValidation, "failed to parse JSON input", err)
	}
	items := make([]Item, 0, len(raw))
	for i, r := range raw {
		if r.Data == "" {
			return nil, qrerrors.New(qrerrors.ErrCodeValidation,
				fmt.Sprintf("item %d: field \"data\" is required", i))
		}
		items = append(items, Item{
			ID:   r.ID,
			Data: r.Data,
		})
	}
	return items, nil
}

// FromCSV reads batch items from a CSV source. The CSV must have a "data"
// column; an optional "id" column sets the item ID.
//
//nolint:gocyclo,cyclop // CSV parsing requires sequential header/record processing
func (*Processor) FromCSV(_ context.Context, reader io.Reader) ([]Item, error) {
	cr := csv.NewReader(reader)
	cr.FieldsPerRecord = -1
	cr.TrimLeadingSpace = true
	header, err := cr.Read()
	if err != nil {
		if err == io.EOF {
			return nil, nil
		}
		return nil, qrerrors.Wrap(qrerrors.ErrCodeValidation, "failed to read CSV header", err)
	}
	normalised := make([]string, len(header))
	for i, h := range header {
		normalised[i] = strings.ToLower(strings.TrimSpace(h))
	}
	dataIdx, idIdx := -1, -1
	for i, col := range normalised {
		switch col {
		case "data":
			dataIdx = i
		case "id":
			idIdx = i
		}
	}
	if dataIdx == -1 {
		return nil, qrerrors.New(qrerrors.ErrCodeValidation,
			"CSV header must contain a \"data\" column")
	}
	var items []Item
	for {
		record, readErr := cr.Read()
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return items, qrerrors.Wrap(qrerrors.ErrCodeValidation, "failed to read CSV row", readErr)
		}
		if len(record) == 0 || (len(record) == 1 && strings.TrimSpace(record[0]) == "") {
			continue
		}
		item := Item{Data: strings.TrimSpace(record[dataIdx])}
		if idIdx >= 0 && idIdx < len(record) {
			item.ID = strings.TrimSpace(record[idIdx])
		}
		items = append(items, item)
	}
	return items, nil
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

// resolvePayload returns the item's Payload if set, otherwise wraps Data as TextPayload.
func (*Processor) resolvePayload(item Item) payload.Payload {
	if item.Payload != nil {
		return item.Payload
	}
	return &payload.TextPayload{Text: item.Data}
}

type jsonItem struct {
	ID     string `json:"id"`
	Data   string `json:"data"`
	Format string `json:"format,omitempty"`
}

// saveResults writes successful results to the configured output directory.
func (p *Processor) saveResults(ctx context.Context, results []Result) {
	if p.outputDir == "" {
		return
	}
	ext := formatExtension(p.format)
	for i, r := range results {
		if r.Err != nil || len(r.Data) == 0 {
			continue
		}
		name := r.ID
		if name == "" {
			name = strconv.Itoa(i)
		}
		fp := filepath.Join(p.outputDir, name+"."+ext)
		if wErr := p.store.Save(ctx, fp, r.Data, 0o644); wErr != nil {
			results[i].Err = qrerrors.Wrap(qrerrors.ErrCodeFileWrite, "failed to write file", wErr)
			continue
		}
		results[i].Path = fp
	}
}

// buildBatchError aggregates individual errors into a single BatchError.
func buildBatchError(results []Result) error {
	be := qrerrors.NewBatchError(len(results))
	for i, r := range results {
		if r.Err != nil {
			be.Errors[i] = r.Err
		}
	}
	if len(be.Errors) == 0 {
		return nil
	}
	return qrerrors.Wrap(qrerrors.ErrCodeBatch,
		fmt.Sprintf("batch processing completed with %d error(s)", len(be.Errors)), be)
}

// computeBatchStats derives aggregate timing statistics from per-item durations.
func computeBatchStats(durations []time.Duration, results []Result, totalTime time.Duration) *BatchStats {
	stats := &BatchStats{
		Total:     len(results),
		TotalTime: totalTime,
	}
	if len(results) == 0 {
		return stats
	}
	var sum time.Duration
	minDur := time.Duration(1<<63 - 1)
	maxDur := time.Duration(0)
	succeeded := 0
	for i, r := range results {
		if r.Err != nil {
			stats.Failed++
			continue
		}
		stats.Succeeded++
		d := durations[i]
		sum += d
		if d < minDur {
			minDur = d
		}
		if d > maxDur {
			maxDur = d
		}
	}
	if succeeded > 0 {
		stats.AvgTime = sum / time.Duration(succeeded)
		stats.MinTime = minDur
		stats.MaxTime = maxDur
	}
	return stats
}

// formatExtension returns the file extension for a qrcode.Format.
func formatExtension(f qrcode.Format) string {
	ext := f.Extension()
	if ext != "" && ext[0] == '.' {
		return ext[1:]
	}
	return ext
}

// ---------------------------------------------------------------------------
// Package-level convenience functions
// ---------------------------------------------------------------------------

// QuickBatch generates PNG QR codes for a list of data strings.
func QuickBatch(ctx context.Context, dataList []string, size ...int) ([][]byte, error) {
	s := 256
	if len(size) > 0 && size[0] > 0 {
		s = size[0]
	}
	gen, err := qrcode.New(qrcode.WithDefaultSize(s))
	if err != nil {
		return nil, fmt.Errorf("create qrcode client: %w", err)
	}
	defer func() { _ = gen.Close() }()
	items := make([]Item, len(dataList))
	for i, d := range dataList {
		items[i] = Item{Data: d}
	}
	proc := NewProcessor(gen, WithBatchFormat(qrcode.FormatPNG))
	results, procErr := proc.Process(ctx, items)
	output := make([][]byte, len(results))
	for i, r := range results {
		if r.Err != nil {
			output[i] = nil
			continue
		}
		if len(r.Data) > 0 {
			output[i] = r.Data
			continue
		}
		var buf bytes.Buffer
		if wErr := gen.GenerateToWriter(ctx, &payload.TextPayload{Text: dataList[i]}, &buf, qrcode.FormatPNG); wErr != nil {
			output[i] = nil
			continue
		}
		output[i] = buf.Bytes()
	}
	return output, procErr
}

// BatchGenerateWithStats is a convenience wrapper for ProcessWithStats.
func BatchGenerateWithStats(ctx context.Context, gen qrcode.Client, items []Item, opts ...ProcessorOption) ([]Result, *BatchStats, error) {
	proc := NewProcessor(gen, opts...)
	return proc.ProcessWithStats(ctx, items)
}

// BatchSaveToDir is a convenience wrapper for SaveToDir.
func BatchSaveToDir(ctx context.Context, gen qrcode.Client, items []Item, outputDir string, opts ...ProcessorOption) ([]Result, error) {
	proc := NewProcessor(gen, opts...)
	return proc.SaveToDir(ctx, items, outputDir)
}
