package batch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	qrcode "github.com/os-gomod/qrcode"
	"github.com/os-gomod/qrcode/encoding"
	"github.com/os-gomod/qrcode/payload"
)

// newTestGen creates a lightweight generator for batch tests.
func newTestGen(t *testing.T) qrcode.Generator {
	t.Helper()
	gen, err := qrcode.New(qrcode.WithDefaultSize(100))
	if err != nil {
		t.Fatalf("failed to create test generator: %v", err)
	}
	t.Cleanup(func() { gen.Close(context.Background()) })
	return gen
}

// --- NewProcessor ---

func TestNewProcessor_Defaults(t *testing.T) {
	t.Parallel()
	gen := newTestGen(t)
	p := NewProcessor(gen)
	if p == nil {
		t.Fatal("NewProcessor returned nil")
	}
	if p.concurrency != defaultConcurrency {
		t.Errorf("default concurrency = %d, want %d", p.concurrency, defaultConcurrency)
	}
	if p.format != qrcode.FormatPNG {
		t.Errorf("default format = %v, want FormatPNG", p.format)
	}
	if p.outputDir != "" {
		t.Errorf("default outputDir = %q, want empty", p.outputDir)
	}
}

func TestNewProcessor_WithConcurrency(t *testing.T) {
	t.Parallel()
	gen := newTestGen(t)
	p := NewProcessor(gen, WithBatchConcurrency(8))
	if p.concurrency != 8 {
		t.Errorf("concurrency = %d, want 8", p.concurrency)
	}
}

func TestNewProcessor_WithConcurrencyZero(t *testing.T) {
	t.Parallel()
	gen := newTestGen(t)
	p := NewProcessor(gen, WithBatchConcurrency(0))
	// Zero should be ignored; default is used
	if p.concurrency != defaultConcurrency {
		t.Errorf("concurrency = %d, want %d (default)", p.concurrency, defaultConcurrency)
	}
}

func TestNewProcessor_WithFormat(t *testing.T) {
	t.Parallel()
	gen := newTestGen(t)
	p := NewProcessor(gen, WithBatchFormat(qrcode.FormatSVG))
	if p.format != qrcode.FormatSVG {
		t.Errorf("format = %v, want FormatSVG", p.format)
	}
	if !p.renderFormat {
		t.Error("renderFormat should be true after WithBatchFormat")
	}
}

func TestNewProcessor_WithOutputDir(t *testing.T) {
	t.Parallel()
	gen := newTestGen(t)
	p := NewProcessor(gen, WithBatchOutputDir("/tmp/qrcodes"))
	if p.outputDir != "/tmp/qrcodes" {
		t.Errorf("outputDir = %q, want /tmp/qrcodes", p.outputDir)
	}
}

func TestNewProcessor_MultipleOptions(t *testing.T) {
	t.Parallel()
	gen := newTestGen(t)
	p := NewProcessor(gen,
		WithBatchConcurrency(2),
		WithBatchFormat(qrcode.FormatPNG),
		WithBatchOutputDir("/out"),
	)
	if p.concurrency != 2 {
		t.Errorf("concurrency = %d, want 2", p.concurrency)
	}
	if !p.renderFormat {
		t.Error("renderFormat should be true")
	}
	if p.outputDir != "/out" {
		t.Errorf("outputDir = %q", p.outputDir)
	}
}

// --- Process ---

func TestProcess_EmptyItems(t *testing.T) {
	t.Parallel()
	gen := newTestGen(t)
	p := NewProcessor(gen)
	results, err := p.Process(context.Background(), nil)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if results != nil {
		t.Errorf("expected nil results for empty input, got %d results", len(results))
	}
}

func TestProcess_SingleItem(t *testing.T) {
	t.Parallel()
	gen := newTestGen(t)
	p := NewProcessor(gen)
	results, err := p.Process(context.Background(), []Item{
		{ID: "test1", Data: "hello world"},
	})
	if err != nil {
		t.Fatalf("Process error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].ID != "test1" {
		t.Errorf("ID = %q, want %q", results[0].ID, "test1")
	}
	if results[0].Err != nil {
		t.Errorf("unexpected error: %v", results[0].Err)
	}
	if results[0].QRCode == nil {
		t.Error("QRCode should not be nil")
	}
}

func TestProcess_MultipleItems(t *testing.T) {
	t.Parallel()
	gen := newTestGen(t)
	p := NewProcessor(gen, WithBatchConcurrency(2))
	items := []Item{
		{ID: "a", Data: "alpha"},
		{ID: "b", Data: "beta"},
		{ID: "c", Data: "gamma"},
	}
	results, err := p.Process(context.Background(), items)
	if err != nil {
		t.Fatalf("Process error: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	for _, r := range results {
		if r.QRCode == nil {
			t.Errorf("result %q: QRCode is nil", r.ID)
		}
		if r.Err != nil {
			t.Errorf("result %q: unexpected error %v", r.ID, r.Err)
		}
	}
}

func TestProcess_WithRenderedFormat(t *testing.T) {
	t.Parallel()
	gen := newTestGen(t)
	p := NewProcessor(gen, WithBatchFormat(qrcode.FormatPNG))
	results, err := p.Process(context.Background(), []Item{
		{ID: "img", Data: "test data"},
	})
	if err != nil {
		t.Fatalf("Process error: %v", err)
	}
	if len(results) != 1 || results[0].Err != nil {
		t.Fatalf("unexpected error in result")
	}
	if len(results[0].Data) == 0 {
		t.Error("expected rendered PNG data, got empty bytes")
	}
	// PNG magic bytes
	if len(results[0].Data) > 4 && results[0].Data[0] != 0x89 {
		t.Error("expected PNG magic bytes")
	}
}

func TestProcess_WithPayload(t *testing.T) {
	t.Parallel()
	gen := newTestGen(t)
	p := NewProcessor(gen)
	items := []Item{
		{ID: "custom", Payload: &payload.TextPayload{Text: "custom payload"}},
	}
	results, err := p.Process(context.Background(), items)
	if err != nil {
		t.Fatalf("Process error: %v", err)
	}
	if results[0].QRCode == nil {
		t.Error("expected QRCode from custom payload")
	}
}

func TestProcess_CancelledContext(t *testing.T) {
	t.Parallel()
	gen := newTestGen(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	p := NewProcessor(gen, WithBatchConcurrency(1))
	results, err := p.Process(ctx, []Item{
		{ID: "cancelled", Data: "should fail"},
	})
	// Results should still be returned with context error
	if results == nil {
		t.Fatal("expected results even with cancelled context")
	}
	if results[0].Err == nil {
		t.Error("expected context cancelled error")
	}
	// err may be non-nil due to batch error wrapper
	_ = err
}

func TestProcess_PreservesOrder(t *testing.T) {
	t.Parallel()
	gen := newTestGen(t)
	p := NewProcessor(gen, WithBatchConcurrency(4))
	items := make([]Item, 10)
	for i := range items {
		items[i] = Item{ID: fmt.Sprintf("item-%d", i), Data: fmt.Sprintf("data-%d", i)}
	}
	results, err := p.Process(context.Background(), items)
	if err != nil {
		t.Fatalf("Process error: %v", err)
	}
	if len(results) != 10 {
		t.Fatalf("expected 10 results, got %d", len(results))
	}
	for i, r := range results {
		expected := fmt.Sprintf("item-%d", i)
		if r.ID != expected {
			t.Errorf("result[%d].ID = %q, want %q", i, r.ID, expected)
		}
	}
}

// --- ProcessWithStats ---

func TestProcessWithStats_EmptyItems(t *testing.T) {
	t.Parallel()
	gen := newTestGen(t)
	p := NewProcessor(gen)
	results, stats, err := p.ProcessWithStats(context.Background(), nil)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if results != nil {
		t.Error("expected nil results")
	}
	if stats == nil {
		t.Fatal("expected non-nil stats")
	}
	if stats.Total != 0 {
		t.Errorf("stats.Total = %d, want 0", stats.Total)
	}
}

func TestProcessWithStats_SuccessfulBatch(t *testing.T) {
	t.Parallel()
	gen := newTestGen(t)
	p := NewProcessor(gen)
	items := []Item{
		{ID: "s1", Data: "stat-test-1"},
		{ID: "s2", Data: "stat-test-2"},
		{ID: "s3", Data: "stat-test-3"},
	}
	results, stats, err := p.ProcessWithStats(context.Background(), items)
	if err != nil {
		t.Fatalf("ProcessWithStats error: %v", err)
	}
	if stats.Total != 3 {
		t.Errorf("Total = %d, want 3", stats.Total)
	}
	if stats.Succeeded != 3 {
		t.Errorf("Succeeded = %d, want 3", stats.Succeeded)
	}
	if stats.Failed != 0 {
		t.Errorf("Failed = %d, want 0", stats.Failed)
	}
	if stats.TotalTime <= 0 {
		t.Error("TotalTime should be positive")
	}
	if stats.MinTime > stats.MaxTime {
		t.Error("MinTime should be <= MaxTime")
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
}

// --- SaveToDir ---

func TestSaveToDir_CreatesDirectoryAndFiles(t *testing.T) {
	t.Parallel()
	gen := newTestGen(t)
	p := NewProcessor(gen)

	dir := filepath.Join(os.TempDir(), fmt.Sprintf("qr_test_%d", time.Now().UnixNano()))
	t.Cleanup(func() { os.RemoveAll(dir) })

	items := []Item{
		{ID: "code1", Data: "save-test-1"},
		{ID: "code2", Data: "save-test-2"},
	}
	results, err := p.SaveToDir(context.Background(), items, dir)
	if err != nil {
		t.Fatalf("SaveToDir error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	for _, r := range results {
		if r.Err != nil {
			t.Errorf("unexpected error: %v", r.Err)
		}
		if r.Path == "" {
			t.Error("Path should be set after SaveToDir")
		}
		if _, statErr := os.Stat(r.Path); statErr != nil {
			t.Errorf("file not found at %s: %v", r.Path, statErr)
		}
	}
}

func TestSaveToDir_EmptyItems(t *testing.T) {
	t.Parallel()
	gen := newTestGen(t)
	p := NewProcessor(gen)

	dir := filepath.Join(os.TempDir(), fmt.Sprintf("qr_empty_%d", time.Now().UnixNano()))
	t.Cleanup(func() { os.RemoveAll(dir) })

	results, err := p.SaveToDir(context.Background(), nil, dir)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if results != nil {
		t.Error("expected nil results for empty items")
	}
}

func TestSaveToDir_WithExistingFormat(t *testing.T) {
	t.Parallel()
	gen := newTestGen(t)
	// Set SVG format; SaveToDir should still create PNG files since no format was rendered
	p := NewProcessor(gen, WithBatchFormat(qrcode.FormatSVG))

	dir := filepath.Join(os.TempDir(), fmt.Sprintf("qr_svg_%d", time.Now().UnixNano()))
	t.Cleanup(func() { os.RemoveAll(dir) })

	items := []Item{
		{ID: "svg-test", Data: "svg-data"},
	}
	results, err := p.SaveToDir(context.Background(), items, dir)
	if err != nil {
		t.Fatalf("SaveToDir error: %v", err)
	}
	if results[0].Err != nil {
		t.Errorf("unexpected error: %v", results[0].Err)
	}
	if results[0].Path == "" {
		t.Error("Path should be set")
	}
}

func TestSaveToDir_InvalidDir(t *testing.T) {
	t.Parallel()
	gen := newTestGen(t)
	p := NewProcessor(gen)

	// Use a path that can't be created
	results, err := p.SaveToDir(context.Background(), []Item{
		{ID: "x", Data: "data"},
	}, "/dev/null/impossible/subdir")
	if err == nil {
		t.Error("expected error for invalid directory")
	}
	if results != nil {
		t.Error("expected nil results for invalid dir")
	}
}

// --- FromJSON ---

func TestFromJSON_ValidInput(t *testing.T) {
	t.Parallel()
	gen := newTestGen(t)
	p := NewProcessor(gen)

	input := `[{"id":"j1","data":"hello"},{"id":"j2","data":"world"}]`
	items, err := p.FromJSON(context.Background(), strings.NewReader(input))
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].ID != "j1" || items[0].Data != "hello" {
		t.Errorf("item[0] = %+v, want {ID:j1, Data:hello}", items[0])
	}
	if items[1].ID != "j2" || items[1].Data != "world" {
		t.Errorf("item[1] = %+v, want {ID:j2, Data:world}", items[1])
	}
}

func TestFromJSON_InvalidJSON(t *testing.T) {
	t.Parallel()
	gen := newTestGen(t)
	p := NewProcessor(gen)

	_, err := p.FromJSON(context.Background(), strings.NewReader("not json"))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestFromJSON_MissingDataField(t *testing.T) {
	t.Parallel()
	gen := newTestGen(t)
	p := NewProcessor(gen)

	input := `[{"id":"no-data"}]`
	_, err := p.FromJSON(context.Background(), strings.NewReader(input))
	if err == nil {
		t.Error("expected error when 'data' field is missing")
	}
}

func TestFromJSON_EmptyArray(t *testing.T) {
	t.Parallel()
	gen := newTestGen(t)
	p := NewProcessor(gen)

	input := `[]`
	items, err := p.FromJSON(context.Background(), strings.NewReader(input))
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected 0 items, got %d", len(items))
	}
}

func TestFromJSON_WithFormatField(t *testing.T) {
	t.Parallel()
	gen := newTestGen(t)
	p := NewProcessor(gen)

	input := `[{"id":"fmt","data":"test","format":"svg"}]`
	items, err := p.FromJSON(context.Background(), strings.NewReader(input))
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	if len(items) != 1 || items[0].Data != "test" {
		t.Errorf("unexpected items: %+v", items)
	}
}

// --- FromCSV ---

func TestFromCSV_ValidInput(t *testing.T) {
	t.Parallel()
	gen := newTestGen(t)
	p := NewProcessor(gen)

	input := "id,data\nrow1,hello\nrow2,world\n"
	items, err := p.FromCSV(context.Background(), strings.NewReader(input))
	if err != nil {
		t.Fatalf("FromCSV error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].ID != "row1" || items[0].Data != "hello" {
		t.Errorf("item[0] = %+v", items[0])
	}
	if items[1].ID != "row2" || items[1].Data != "world" {
		t.Errorf("item[1] = %+v", items[1])
	}
}

func TestFromCSV_DataOnlyHeader(t *testing.T) {
	t.Parallel()
	gen := newTestGen(t)
	p := NewProcessor(gen)

	input := "data\nhello\nworld\n"
	items, err := p.FromCSV(context.Background(), strings.NewReader(input))
	if err != nil {
		t.Fatalf("FromCSV error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].Data != "hello" {
		t.Errorf("item[0].Data = %q, want %q", items[0].Data, "hello")
	}
	if items[1].Data != "world" {
		t.Errorf("item[1].Data = %q, want %q", items[1].Data, "world")
	}
}

func TestFromCSV_MissingDataColumn(t *testing.T) {
	t.Parallel()
	gen := newTestGen(t)
	p := NewProcessor(gen)

	input := "id,name\n1,foo\n"
	_, err := p.FromCSV(context.Background(), strings.NewReader(input))
	if err == nil {
		t.Error("expected error when 'data' column is missing")
	}
}

func TestFromCSV_EmptyInput(t *testing.T) {
	t.Parallel()
	gen := newTestGen(t)
	p := NewProcessor(gen)

	items, err := p.FromCSV(context.Background(), strings.NewReader(""))
	if err != nil {
		t.Fatalf("FromCSV error: %v", err)
	}
	if items != nil {
		t.Errorf("expected nil items for empty input, got %d", len(items))
	}
}

func TestFromCSV_HeaderOnly(t *testing.T) {
	t.Parallel()
	gen := newTestGen(t)
	p := NewProcessor(gen)

	input := "data\n"
	items, err := p.FromCSV(context.Background(), strings.NewReader(input))
	if err != nil {
		t.Fatalf("FromCSV error: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected 0 items for header-only CSV, got %d", len(items))
	}
}

func TestFromCSV_SkipsBlankRows(t *testing.T) {
	t.Parallel()
	gen := newTestGen(t)
	p := NewProcessor(gen)

	input := "data\nhello\n\n\nworld\n"
	items, err := p.FromCSV(context.Background(), strings.NewReader(input))
	if err != nil {
		t.Fatalf("FromCSV error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items (blank rows skipped), got %d", len(items))
	}
}

func TestFromCSV_CaseInsensitiveHeader(t *testing.T) {
	t.Parallel()
	gen := newTestGen(t)
	p := NewProcessor(gen)

	input := "ID,DATA\nr1,test1\nr2,test2\n"
	items, err := p.FromCSV(context.Background(), strings.NewReader(input))
	if err != nil {
		t.Fatalf("FromCSV error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].ID != "r1" || items[0].Data != "test1" {
		t.Errorf("item[0] = %+v", items[0])
	}
}

// --- QuickBatch ---

func TestQuickBatch_BasicUsage(t *testing.T) {
	t.Parallel()
	output, err := QuickBatch(context.Background(), []string{"hello", "world"}, 100)
	if err != nil {
		t.Fatalf("QuickBatch error: %v", err)
	}
	if len(output) != 2 {
		t.Fatalf("expected 2 results, got %d", len(output))
	}
	for i, data := range output {
		if len(data) == 0 {
			t.Errorf("output[%d] is empty", i)
		}
	}
}

func TestQuickBatch_EmptyInput(t *testing.T) {
	t.Parallel()
	output, err := QuickBatch(context.Background(), nil)
	if err != nil {
		t.Fatalf("QuickBatch error: %v", err)
	}
	if len(output) != 0 {
		t.Errorf("expected 0 results, got %d", len(output))
	}
}

func TestQuickBatch_DefaultSize(t *testing.T) {
	t.Parallel()
	// No size argument; should use default of 256
	output, err := QuickBatch(context.Background(), []string{"test"})
	if err != nil {
		t.Fatalf("QuickBatch error: %v", err)
	}
	if len(output) != 1 || len(output[0]) == 0 {
		t.Error("expected one non-empty result")
	}
}

func TestQuickBatch_ZeroSize(t *testing.T) {
	t.Parallel()
	// Zero size should fall back to default
	output, err := QuickBatch(context.Background(), []string{"test"}, 0)
	if err != nil {
		t.Fatalf("QuickBatch error: %v", err)
	}
	if len(output) != 1 || len(output[0]) == 0 {
		t.Error("expected one non-empty result")
	}
}

// --- BatchGenerateWithStats ---

func TestBatchGenerateWithStats_Basic(t *testing.T) {
	t.Parallel()
	gen := newTestGen(t)
	items := []Item{
		{ID: "bs1", Data: "batch-stat-1"},
		{ID: "bs2", Data: "batch-stat-2"},
	}
	results, stats, err := BatchGenerateWithStats(context.Background(), gen, items)
	if err != nil {
		t.Fatalf("BatchGenerateWithStats error: %v", err)
	}
	if stats == nil {
		t.Fatal("expected non-nil stats")
	}
	if stats.Total != 2 {
		t.Errorf("stats.Total = %d, want 2", stats.Total)
	}
	if stats.Succeeded != 2 {
		t.Errorf("stats.Succeeded = %d, want 2", stats.Succeeded)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
}

func TestBatchGenerateWithStats_WithOptions(t *testing.T) {
	t.Parallel()
	gen := newTestGen(t)
	items := []Item{{ID: "opt", Data: "with-options"}}
	results, stats, err := BatchGenerateWithStats(
		context.Background(), gen, items,
		WithBatchConcurrency(1),
		WithBatchFormat(qrcode.FormatPNG),
	)
	if err != nil {
		t.Fatalf("BatchGenerateWithStats error: %v", err)
	}
	if stats.Total != 1 {
		t.Errorf("stats.Total = %d, want 1", stats.Total)
	}
	if len(results[0].Data) == 0 {
		t.Error("expected rendered data with format option")
	}
}

// --- BatchSaveToDir ---

func TestBatchSaveToDir_Basic(t *testing.T) {
	t.Parallel()
	gen := newTestGen(t)
	dir := filepath.Join(os.TempDir(), fmt.Sprintf("qr_bsd_%d", time.Now().UnixNano()))
	t.Cleanup(func() { os.RemoveAll(dir) })

	items := []Item{
		{ID: "bsd1", Data: "batch-save-1"},
	}
	results, err := BatchSaveToDir(context.Background(), gen, items, dir)
	if err != nil {
		t.Fatalf("BatchSaveToDir error: %v", err)
	}
	if len(results) != 1 || results[0].Err != nil {
		t.Fatalf("unexpected error in results")
	}
	if results[0].Path == "" {
		t.Error("expected Path to be set")
	}
}

// --- formatExtension (tested indirectly via Process) ---

func TestFormatExtension_Indirect(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		fmt  qrcode.Format
		want string
	}{
		{"PNG", qrcode.FormatPNG, "png"},
		{"SVG", qrcode.FormatSVG, "svg"},
		{"Terminal", qrcode.FormatTerminal, "txt"},
		{"PDF", qrcode.FormatPDF, "pdf"},
		{"Base64", qrcode.FormatBase64, "b64"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			gen := newTestGen(t)
			dir := filepath.Join(os.TempDir(), fmt.Sprintf("qr_ext_%s_%d", tc.name, time.Now().UnixNano()))
			t.Cleanup(func() { os.RemoveAll(dir) })

			p := NewProcessor(gen, WithBatchFormat(tc.fmt))
			items := []Item{{ID: "ext-test", Data: "extension-test"}}
			results, err := p.SaveToDir(context.Background(), items, dir)
			if err != nil {
				t.Fatalf("SaveToDir error: %v", err)
			}
			if results[0].Err != nil {
				t.Fatalf("result error: %v", results[0].Err)
			}
			if !strings.HasSuffix(results[0].Path, "."+tc.want) {
				t.Errorf("path %q should end with .%s", results[0].Path, tc.want)
			}
		})
	}
}

// --- computeBatchStats (tested indirectly via ProcessWithStats) ---

func TestComputeBatchStats_Indirect_AllSuccess(t *testing.T) {
	t.Parallel()
	gen := newTestGen(t)
	p := NewProcessor(gen)
	items := []Item{
		{ID: "cs1", Data: "computed-1"},
		{ID: "cs2", Data: "computed-2"},
	}
	_, stats, err := p.ProcessWithStats(context.Background(), items)
	if err != nil {
		t.Fatalf("ProcessWithStats error: %v", err)
	}
	if stats.Failed != 0 {
		t.Errorf("expected 0 failures, got %d", stats.Failed)
	}
	if stats.Succeeded != 2 {
		t.Errorf("expected 2 successes, got %d", stats.Succeeded)
	}
	if stats.MinTime > stats.MaxTime {
		t.Error("MinTime should be <= MaxTime")
	}
}

// --- resolvePayload ---

func TestResolvePayload_TextPayload(t *testing.T) {
	t.Parallel()
	gen := newTestGen(t)
	p := NewProcessor(gen)

	item := Item{Data: "hello"}
	pl := p.resolvePayload(item)
	if pl == nil {
		t.Fatal("resolvePayload returned nil")
	}
	encoded, err := pl.Encode()
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}
	if encoded != "hello" {
		t.Errorf("encoded = %q, want %q", encoded, "hello")
	}
}

func TestResolvePayload_CustomPayload(t *testing.T) {
	t.Parallel()
	gen := newTestGen(t)
	p := NewProcessor(gen)

	custom := &payload.TextPayload{Text: "custom"}
	item := Item{Payload: custom}
	pl := p.resolvePayload(item)
	if pl != custom {
		t.Error("resolvePayload should return the custom Payload when set")
	}
}

// --- BatchStats struct ---

func TestBatchStats_ZeroValue(t *testing.T) {
	t.Parallel()
	s := &BatchStats{}
	if s.Total != 0 || s.Succeeded != 0 || s.Failed != 0 {
		t.Error("zero-value BatchStats should have zero counts")
	}
	if s.TotalTime != 0 || s.AvgTime != 0 {
		t.Error("zero-value BatchStats should have zero durations")
	}
}

// --- Integration: full pipeline ---

func TestIntegration_FromJSONToProcess(t *testing.T) {
	t.Parallel()
	gen := newTestGen(t)
	p := NewProcessor(gen, WithBatchFormat(qrcode.FormatPNG))

	jsonInput := `[
                {"id":"i1","data":"integration-test-1"},
                {"id":"i2","data":"integration-test-2"},
                {"id":"i3","data":"integration-test-3"}
        ]`
	items, err := p.FromJSON(context.Background(), strings.NewReader(jsonInput))
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}

	results, err := p.Process(context.Background(), items)
	if err != nil {
		t.Fatalf("Process error: %v", err)
	}

	for i, r := range results {
		if r.Err != nil {
			t.Errorf("result[%d] error: %v", i, r.Err)
		}
		if r.QRCode == nil {
			t.Errorf("result[%d]: QRCode is nil", i)
		}
		if len(r.Data) == 0 {
			t.Errorf("result[%d]: expected rendered data", i)
		}
	}
}

func TestIntegration_FromCSVToProcessWithStats(t *testing.T) {
	t.Parallel()
	gen := newTestGen(t)
	p := NewProcessor(gen)

	csvInput := "id,data\nrow1,csv-stat-1\nrow2,csv-stat-2\n"
	items, err := p.FromCSV(context.Background(), strings.NewReader(csvInput))
	if err != nil {
		t.Fatalf("FromCSV error: %v", err)
	}

	results, stats, err := p.ProcessWithStats(context.Background(), items)
	if err != nil {
		t.Fatalf("ProcessWithStats error: %v", err)
	}
	if stats.Total != 2 || stats.Succeeded != 2 {
		t.Errorf("stats = %+v", stats)
	}
	_ = results
}

// --- JSON round-trip ---

func TestFromJSON_RoundTrip(t *testing.T) {
	t.Parallel()
	gen := newTestGen(t)
	p := NewProcessor(gen)

	original := []Item{
		{ID: "rt1", Data: "round-trip-1"},
		{ID: "rt2", Data: "round-trip-2"},
	}
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("json.Marshal error: %v", err)
	}
	parsed, err := p.FromJSON(context.Background(), bytes.NewReader(data))
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	if len(parsed) != len(original) {
		t.Fatalf("expected %d items, got %d", len(original), len(parsed))
	}
	for i := range original {
		if parsed[i].ID != original[i].ID || parsed[i].Data != original[i].Data {
			t.Errorf("item[%d] mismatch: got %+v, want %+v", i, parsed[i], original[i])
		}
	}
}

// --- Process with output dir set (saveResults path) ---

func TestProcess_WithOutputDir(t *testing.T) {
	t.Parallel()
	gen := newTestGen(t)
	dir := filepath.Join(os.TempDir(), fmt.Sprintf("qr_proc_out_%d", time.Now().UnixNano()))
	t.Cleanup(func() { os.RemoveAll(dir) })

	p := NewProcessor(gen, WithBatchFormat(qrcode.FormatPNG), WithBatchOutputDir(dir))
	items := []Item{
		{ID: "out1", Data: "output-dir-test"},
	}
	results, err := p.Process(context.Background(), items)
	if err != nil {
		t.Fatalf("Process error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Path == "" {
		t.Error("Path should be set when outputDir is configured")
	}
}

// --- Misc edge cases ---

func TestProcess_NoID_GeneratesIndexedFilename(t *testing.T) {
	t.Parallel()
	gen := newTestGen(t)
	dir := filepath.Join(os.TempDir(), fmt.Sprintf("qr_noid_%d", time.Now().UnixNano()))
	t.Cleanup(func() { os.RemoveAll(dir) })

	p := NewProcessor(gen)
	items := []Item{
		{Data: "no-id-test"},
	}
	results, err := p.SaveToDir(context.Background(), items, dir)
	if err != nil {
		t.Fatalf("SaveToDir error: %v", err)
	}
	if results[0].Err != nil {
		t.Fatalf("result error: %v", results[0].Err)
	}
	// Should create file named "0.png" (index-based)
	if !strings.HasSuffix(results[0].Path, "0.png") {
		t.Errorf("expected filename '0.png', got %s", filepath.Base(results[0].Path))
	}
}

func TestProcess_LargeBatch(t *testing.T) {
	t.Parallel()
	gen := newTestGen(t)
	p := NewProcessor(gen, WithBatchConcurrency(2))

	items := make([]Item, 50)
	for i := range items {
		items[i] = Item{ID: fmt.Sprintf("large-%d", i), Data: fmt.Sprintf("large-batch-item-%d", i)}
	}
	results, err := p.Process(context.Background(), items)
	if err != nil {
		t.Fatalf("Process error: %v", err)
	}
	if len(results) != 50 {
		t.Fatalf("expected 50 results, got %d", len(results))
	}
	successCount := 0
	for _, r := range results {
		if r.Err == nil && r.QRCode != nil {
			successCount++
		}
	}
	if successCount != 50 {
		t.Errorf("expected 50 successful results, got %d", successCount)
	}
}

func TestProcess_WithRenderedSVG(t *testing.T) {
	t.Parallel()
	gen := newTestGen(t)
	p := NewProcessor(gen, WithBatchFormat(qrcode.FormatSVG))
	items := []Item{{ID: "svg", Data: "svg-render-test"}}
	results, err := p.Process(context.Background(), items)
	if err != nil {
		t.Fatalf("Process error: %v", err)
	}
	if len(results[0].Data) == 0 {
		t.Error("expected SVG data")
	}
	if !bytes.Contains(results[0].Data, []byte("<svg")) {
		t.Error("expected SVG content in output data")
	}
}

// --- BatchStats computation edge cases ---

func TestBatchStats_AllFailures(t *testing.T) {
	t.Parallel()
	// Create a cancelled context so all items fail
	gen := newTestGen(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	p := NewProcessor(gen)
	items := []Item{
		{ID: "f1", Data: "fail-1"},
		{ID: "f2", Data: "fail-2"},
	}
	_, stats, err := p.ProcessWithStats(ctx, items)
	_ = err
	if stats.Total != 2 {
		t.Errorf("Total = %d, want 2", stats.Total)
	}
	if stats.Failed != 2 {
		t.Errorf("Failed = %d, want 2", stats.Failed)
	}
	if stats.Succeeded != 0 {
		t.Errorf("Succeeded = %d, want 0", stats.Succeeded)
	}
}

// Ensure encoding.QRCode has expected fields
func TestQRCodeFieldsExist(t *testing.T) {
	t.Parallel()
	qr := &encoding.QRCode{
		Version:     1,
		Size:        21,
		ECLevel:     1,
		MaskPattern: 0,
		Modules:     nil,
	}
	if qr.Version != 1 {
		t.Errorf("Version = %d", qr.Version)
	}
	if qr.Size != 21 {
		t.Errorf("Size = %d", qr.Size)
	}
	if qr.ECLevel != 1 {
		t.Errorf("ECLevel = %d", qr.ECLevel)
	}
}
