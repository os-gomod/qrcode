// Package batch provides concurrent batch processing for QR code generation with
// support for file output, JSON/CSV input parsing, and per-item statistics.
//
// The package is designed around the [Processor] type, which orchestrates
// concurrent QR code generation using a configurable worker pool. Input can
// be supplied programmatically via [Item] slices or parsed from JSON and CSV
// readers.
//
// # Quick Start
//
// For simple use cases where you just need PNG bytes for a list of strings,
// use [QuickBatch]:
//
//	pngs, err := batch.QuickBatch(ctx, []string{"hello", "world"})
//
// # Processor
//
// For more control over output format, concurrency, and file saving:
//
//	gen, _ := qrcode.New(qrcode.WithDefaultSize(256))
//	proc := batch.NewProcessor(gen,
//	    batch.WithBatchConcurrency(8),
//	    batch.WithBatchFormat(qrcode.FormatPNG),
//	)
//	results, err := proc.Process(ctx, items)
package batch

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	qrcode "github.com/os-gomod/qrcode"
	"github.com/os-gomod/qrcode/encoding"
	qrerrors "github.com/os-gomod/qrcode/errors"
	"github.com/os-gomod/qrcode/payload"
)

// Item represents a single QR code generation task in a batch.
//
// Either Data (plain text) or Payload (a pre-built payload) should be set.
// If both are provided, Payload takes precedence. The optional ID field is
// used for file naming when saving results and for correlating errors.
type Item struct {
	// ID is an optional identifier used for file naming and error correlation.
	ID string
	// Data is the raw text to encode when no Payload is provided.
	Data string
	// Payload is a pre-built payload for encoding. Takes precedence over Data.
	Payload payload.Payload
}

// Result holds the output of a single batch item.
//
// On success, QRCode contains the encoded matrix and Data contains the
// rendered image bytes (if a format was configured). Path is populated
// when results are saved to disk. On failure, Err is non-nil.
type Result struct {
	// ID is the identifier from the corresponding Item.
	ID string
	// QRCode is the encoded QR code matrix (nil on error).
	QRCode *encoding.QRCode
	// Data is the rendered image bytes in the configured format (nil if not rendering).
	Data []byte
	// Err holds any error encountered during generation.
	Err error
	// Path is the filesystem path of the saved file (empty if not saved to disk).
	Path string
}

// BatchStats holds performance statistics for a batch run.
//
// Timings reflect per-item wall-clock durations and are only meaningful
// when concurrency is factored in (TotalTime is the overall elapsed time).
//
//nolint:revive // stutter: BatchStats is the canonical name for this struct
type BatchStats struct {
	// Total is the total number of items in the batch.
	Total int
	// Succeeded is the number of items that completed without error.
	Succeeded int
	// Failed is the number of items that encountered an error.
	Failed int
	// TotalTime is the wall-clock duration of the entire batch run.
	TotalTime time.Duration
	// AvgTime is the average per-item generation duration (successful items only).
	AvgTime time.Duration
	// MinTime is the fastest single-item generation duration.
	MinTime time.Duration
	// MaxTime is the slowest single-item generation duration.
	MaxTime time.Duration
}

const defaultConcurrency = 4

// Processor handles batch QR code generation with configurable concurrency,
// output format, and optional file saving.
//
// Use [NewProcessor] to create a Processor with optional configuration
// via [ProcessorOption] functions.
type Processor struct {
	gen          qrcode.Generator
	concurrency  int
	format       qrcode.Format
	outputDir    string
	renderFormat bool
}

// ProcessorOption configures a Processor during construction.
//
// Options are applied in order and may be composed.
type ProcessorOption func(*Processor)

// WithBatchConcurrency sets the number of concurrent workers used during
// batch processing. Defaults to 4 if not specified. Values less than 1 are
// ignored.
func WithBatchConcurrency(n int) ProcessorOption {
	return func(p *Processor) {
		if n > 0 {
			p.concurrency = n
		}
	}
}

// WithBatchFormat sets the output format for batch results and enables
// format rendering in the Data field of each [Result]. Supported formats
// include PNG, SVG, PDF, Terminal, and Base64.
func WithBatchFormat(f qrcode.Format) ProcessorOption {
	return func(p *Processor) {
		p.format = f
		p.renderFormat = true
	}
}

// WithBatchOutputDir sets the directory where generated QR code files
// are written. The directory is created automatically if it does not exist.
func WithBatchOutputDir(dir string) ProcessorOption {
	return func(p *Processor) {
		p.outputDir = dir
	}
}

// NewProcessor creates a new batch processor with the given QR code
// generator and optional configuration. The processor uses 4 concurrent
// workers and PNG output format by default.
func NewProcessor(gen qrcode.Generator, opts ...ProcessorOption) *Processor {
	p := &Processor{
		gen:         gen,
		concurrency: defaultConcurrency,
		format:      qrcode.FormatPNG,
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

// Process generates QR codes for all items concurrently, returning a
// slice of results aligned by index with the input. If an output directory
// is configured, results are also saved to disk. Returns nil if the input
// is empty. The returned error aggregates any per-item failures.
func (p *Processor) Process(ctx context.Context, items []Item) ([]Result, error) {
	if len(items) == 0 {
		return nil, nil
	}
	n := len(items)
	results := make([]Result, n)
	type workUnit struct {
		idx  int
		item Item
	}
	work := make(chan workUnit, n)
	for i, item := range items {
		work <- workUnit{idx: i, item: item}
	}
	close(work)
	var wg sync.WaitGroup
	for w := 0; w < p.concurrency && w < n; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for unit := range work {
				select {
				case <-ctx.Done():
					results[unit.idx].ID = unit.item.ID
					results[unit.idx].Err = ctx.Err()
					continue
				default:
				}
				pl := p.resolvePayload(unit.item)
				qr, err := p.gen.Generate(ctx, pl)
				results[unit.idx].ID = unit.item.ID
				if err != nil {
					results[unit.idx].Err = err
					continue
				}
				results[unit.idx].QRCode = qr
				if p.renderFormat {
					var buf bytes.Buffer
					if wErr := p.gen.GenerateToWriter(ctx, pl, &buf, p.format); wErr != nil {
						results[unit.idx].Err = wErr
						continue
					}
					results[unit.idx].Data = buf.Bytes()
				}
			}
		}()
	}
	wg.Wait()
	if p.outputDir != "" {
		p.saveResults(results)
	}
	return results, buildBatchError(results)
}

// ProcessWithStats generates QR codes for all items and returns detailed
// per-item timing statistics in addition to the results. Statistics include
// total/average/min/max generation durations and success/failure counts.
func (p *Processor) ProcessWithStats(ctx context.Context, items []Item) ([]Result, *BatchStats, error) {
	batchStart := time.Now()
	if len(items) == 0 {
		return nil, &BatchStats{}, nil
	}
	n := len(items)
	results := make([]Result, n)
	durations := make([]time.Duration, n)
	type workUnit struct {
		idx  int
		item Item
	}
	work := make(chan workUnit, n)
	for i, item := range items {
		work <- workUnit{idx: i, item: item}
	}
	close(work)
	var wg sync.WaitGroup
	for w := 0; w < p.concurrency && w < n; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for unit := range work {
				select {
				case <-ctx.Done():
					results[unit.idx].ID = unit.item.ID
					durations[unit.idx] = 0
					results[unit.idx].Err = ctx.Err()
					continue
				default:
				}
				start := time.Now()
				pl := p.resolvePayload(unit.item)
				qr, err := p.gen.Generate(ctx, pl)
				elapsed := time.Since(start)
				results[unit.idx].ID = unit.item.ID
				durations[unit.idx] = elapsed
				if err != nil {
					results[unit.idx].Err = err
					continue
				}
				results[unit.idx].QRCode = qr
				if p.renderFormat {
					var buf bytes.Buffer
					if wErr := p.gen.GenerateToWriter(ctx, pl, &buf, p.format); wErr != nil {
						results[unit.idx].Err = wErr
						continue
					}
					results[unit.idx].Data = buf.Bytes()
				}
			}
		}()
	}
	wg.Wait()
	if p.outputDir != "" {
		p.saveResults(results)
	}
	stats := computeBatchStats(durations, results, time.Since(batchStart))
	return results, stats, buildBatchError(results)
}

// SaveToDir generates QR codes for all items and saves them as files in
// the specified output directory. The directory is created if needed. Each
// file is named after the item ID (or its index) with the appropriate
// extension for the configured format. PNG is used as the default format.
func (p *Processor) SaveToDir(ctx context.Context, items []Item, outputDir string) ([]Result, error) {
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return nil, qrerrors.Wrap(qrerrors.ErrCodeFileWrite, "failed to create output directory", err)
	}
	origFormat := p.format
	origRenderFormat := p.renderFormat
	if !p.renderFormat {
		p.format = qrcode.FormatPNG
		p.renderFormat = true
	}
	results, err := p.processToDir(ctx, items, outputDir)
	p.format = origFormat
	p.renderFormat = origRenderFormat
	return results, err
}

func (p *Processor) processToDir(ctx context.Context, items []Item, outputDir string) ([]Result, error) {
	n := len(items)
	if n == 0 {
		return nil, nil
	}
	results := make([]Result, n)
	durations := make([]time.Duration, n)
	ext := formatExtension(p.format)
	type workUnit struct {
		idx  int
		item Item
	}
	work := make(chan workUnit, n)
	for i, item := range items {
		work <- workUnit{idx: i, item: item}
	}
	close(work)
	var wg sync.WaitGroup
	for w := 0; w < p.concurrency && w < n; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for unit := range work {
				select {
				case <-ctx.Done():
					results[unit.idx].ID = unit.item.ID
					durations[unit.idx] = 0
					results[unit.idx].Err = ctx.Err()
					continue
				default:
				}
				start := time.Now()
				pl := p.resolvePayload(unit.item)
				qr, err := p.gen.Generate(ctx, pl)
				elapsed := time.Since(start)
				results[unit.idx].ID = unit.item.ID
				durations[unit.idx] = elapsed
				if err != nil {
					results[unit.idx].Err = err
					continue
				}
				results[unit.idx].QRCode = qr
				var buf bytes.Buffer
				if wErr := p.gen.GenerateToWriter(ctx, pl, &buf, p.format); wErr != nil {
					results[unit.idx].Err = wErr
					continue
				}
				results[unit.idx].Data = buf.Bytes()
				name := unit.item.ID
				if name == "" {
					name = fmt.Sprintf("%d", unit.idx)
				}
				fp := filepath.Join(outputDir, name+"."+ext)
				if wErr := os.WriteFile(fp, buf.Bytes(), 0o644); wErr != nil { //nolint:gosec // G306: output files are intentionally world-readable
					results[unit.idx].Err = qrerrors.Wrap(qrerrors.ErrCodeFileWrite, "failed to write file", wErr)
					continue
				}
				results[unit.idx].Path = fp
			}
		}()
	}
	wg.Wait()
	return results, buildBatchError(results)
}

// FromJSON parses a JSON array of objects from reader into a slice of
// [Item] values. Each object must have a "data" field; an optional "id"
// field is used for file naming. Example input:
//
//	[{"id": "item1", "data": "https://example.com"}]
func (p *Processor) FromJSON(_ context.Context, reader io.Reader) ([]Item, error) {
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

// FromCSV parses CSV data from reader into a slice of [Item] values.
// The first row is treated as a header. A required "data" column provides
// the text to encode; an optional "id" column is used for file naming.
// Headers are case-insensitive and leading whitespace is trimmed.
func (p *Processor) FromCSV(_ context.Context, reader io.Reader) ([]Item, error) {
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

func (p *Processor) resolvePayload(item Item) payload.Payload {
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

func (p *Processor) saveResults(results []Result) {
	if p.outputDir == "" {
		return
	}
	_ = os.MkdirAll(p.outputDir, 0o755)
	ext := formatExtension(p.format)
	for i, r := range results {
		if r.Err != nil || len(r.Data) == 0 {
			continue
		}
		name := r.ID
		if name == "" {
			name = fmt.Sprintf("%d", i)
		}
		fp := filepath.Join(p.outputDir, name+"."+ext)
		if wErr := os.WriteFile(fp, r.Data, 0o644); wErr != nil { //nolint:gosec // G306: output files are intentionally world-readable
			results[i].Err = qrerrors.Wrap(qrerrors.ErrCodeFileWrite, "failed to write file", wErr)
			continue
		}
		results[i].Path = fp
	}
}

func buildBatchError(results []Result) error {
	be := qrerrors.NewBatchError()
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

func formatExtension(f qrcode.Format) string {
	switch f {
	case qrcode.FormatPNG:
		return "png"
	case qrcode.FormatSVG:
		return "svg"
	case qrcode.FormatTerminal:
		return "txt"
	case qrcode.FormatPDF:
		return "pdf"
	case qrcode.FormatBase64:
		return "b64"
	default:
		return "png"
	}
}

// QuickBatch generates PNG QR codes for a list of text strings.
// It creates a temporary generator with the given image size (default 256)
// and returns a slice of PNG byte slices aligned with the input. Individual
// items that fail are represented by nil entries; the returned error
// aggregates any failures.
func QuickBatch(ctx context.Context, dataList []string, size ...int) ([][]byte, error) {
	s := 256
	if len(size) > 0 && size[0] > 0 {
		s = size[0]
	}
	gen, err := qrcode.New(qrcode.WithDefaultSize(s))
	if err != nil {
		return nil, err
	}
	defer gen.Close(ctx) //nolint:errcheck // Close error intentionally ignored; generator lifecycle managed by caller
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

// BatchGenerateWithStats is a convenience function that creates a [Processor]
// with the given options and calls [Processor.ProcessWithStats], returning
// results along with per-item timing statistics.
//
//nolint:revive // stutter: BatchGenerateWithStats is the canonical name for this function
func BatchGenerateWithStats(ctx context.Context, gen qrcode.Generator, items []Item, opts ...ProcessorOption) ([]Result, *BatchStats, error) {
	proc := NewProcessor(gen, opts...)
	return proc.ProcessWithStats(ctx, items)
}

// BatchSaveToDir is a convenience function that creates a [Processor]
// with the given options and calls [Processor.SaveToDir], generating QR
// codes and saving them as files in the specified output directory.
//
//nolint:revive // stutter: BatchSaveToDir is the canonical name for this function
func BatchSaveToDir(ctx context.Context, gen qrcode.Generator, items []Item, outputDir string, opts ...ProcessorOption) ([]Result, error) {
	proc := NewProcessor(gen, opts...)
	return proc.SaveToDir(ctx, items, outputDir)
}
