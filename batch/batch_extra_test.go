package batch

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	qrcode "github.com/os-gomod/qrcode/v2"
	"github.com/os-gomod/qrcode/v2/payload"
)

// ---------------------------------------------------------------------------
// Processor option and construction tests
// ---------------------------------------------------------------------------

func TestProcess_DefaultOptions(t *testing.T) {
	// Process without any options — uses defaults (no format, default concurrency).
	gen, err := qrcode.New(qrcode.WithDefaultSize(256))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	defer func() { _ = gen.Close() }()

	p := NewProcessor(gen) // no options
	items := []Item{
		{ID: "x", Data: "default-opt-test"},
	}
	results, err := p.Process(context.Background(), items)
	if err != nil {
		t.Fatalf("Process() error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Err != nil {
		t.Errorf("result error: %v", results[0].Err)
	}
	if results[0].QRCode == nil {
		t.Error("expected non-nil QRCode")
	}
	// Without format option, renderFormat is false, so Data should be empty.
	if len(results[0].Data) != 0 {
		t.Errorf("expected empty Data without format option, got %d bytes", len(results[0].Data))
	}
}

func TestNewProcessor_MultipleOptions(t *testing.T) {
	gen, _ := qrcode.New()
	defer func() { _ = gen.Close() }()

	mockStore := &mockStorage{saved: make(map[string][]byte)}
	p := NewProcessor(gen,
		WithBatchConcurrency(2),
		WithBatchFormat(qrcode.FormatSVG),
		WithBatchOutputDir("/tmp/out"),
		WithBatchStorage(mockStore),
	)
	if p.concurrency != 2 {
		t.Errorf("concurrency = %d, want 2", p.concurrency)
	}
	if p.format != qrcode.FormatSVG {
		t.Errorf("format = %v, want SVG", p.format)
	}
	if !p.renderFormat {
		t.Error("renderFormat should be true")
	}
	if p.outputDir != "/tmp/out" {
		t.Errorf("outputDir = %q, want /tmp/out", p.outputDir)
	}
	if p.store != mockStore {
		t.Error("store should be mockStorage")
	}
}

func TestWithBatchOutputDir(t *testing.T) {
	gen, _ := qrcode.New()
	defer func() { _ = gen.Close() }()

	p := NewProcessor(gen, WithBatchOutputDir("/custom/dir"))
	if p.outputDir != "/custom/dir" {
		t.Errorf("outputDir = %q, want /custom/dir", p.outputDir)
	}
}

func TestProcess_WithOutputDirOption(t *testing.T) {
	// When outputDir is set AND format is enabled, Process should auto-save results.
	tmpDir := t.TempDir()
	gen, err := qrcode.New(qrcode.WithDefaultSize(256))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	defer func() { _ = gen.Close() }()

	p := NewProcessor(gen,
		WithBatchFormat(qrcode.FormatPNG),
		WithBatchOutputDir(tmpDir),
	)
	items := []Item{
		{ID: "auto-save-1", Data: "auto-save-test"},
	}
	results, err := p.Process(context.Background(), items)
	if err != nil {
		t.Fatalf("Process() error: %v", err)
	}
	if results[0].Err != nil {
		t.Fatalf("result error: %v", results[0].Err)
	}
	// File should exist in outputDir
	expectedPath := filepath.Join(tmpDir, "auto-save-1.png")
	if _, statErr := os.Stat(expectedPath); os.IsNotExist(statErr) {
		t.Errorf("expected file at %s", expectedPath)
	}
}

// ---------------------------------------------------------------------------
// ProcessWithStats detailed tests
// ---------------------------------------------------------------------------

func TestProcessWithStats_DetailedTimings(t *testing.T) {
	gen, err := qrcode.New(qrcode.WithDefaultSize(256))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	defer func() { _ = gen.Close() }()

	p := NewProcessor(gen, WithBatchFormat(qrcode.FormatPNG))
	items := []Item{
		{ID: "a", Data: "timing-test-1"},
		{ID: "b", Data: "timing-test-2"},
		{ID: "c", Data: "timing-test-3"},
	}
	results, stats, err := p.ProcessWithStats(context.Background(), items)
	if err != nil {
		t.Fatalf("ProcessWithStats() error: %v", err)
	}
	if stats == nil {
		t.Fatal("stats should not be nil")
	}
	if stats.Total != 3 {
		t.Errorf("stats.Total = %d, want 3", stats.Total)
	}
	if stats.Succeeded != 3 {
		t.Errorf("stats.Succeeded = %d, want 3", stats.Succeeded)
	}
	if stats.Failed != 0 {
		t.Errorf("stats.Failed = %d, want 0", stats.Failed)
	}
	if stats.TotalTime <= 0 {
		t.Error("stats.TotalTime should be > 0")
	}
	if stats.AvgTime < 0 {
		t.Errorf("stats.AvgTime should be >= 0, got %v", stats.AvgTime)
	}
	if stats.MinTime < 0 {
		t.Errorf("stats.MinTime should be >= 0, got %v", stats.MinTime)
	}
	if stats.MaxTime < 0 {
		t.Errorf("stats.MaxTime should be >= 0, got %v", stats.MaxTime)
	}
	if stats.MinTime > stats.MaxTime {
		t.Errorf("MinTime(%v) should be <= MaxTime(%v)", stats.MinTime, stats.MaxTime)
	}
	// AvgTime should be between MinTime and MaxTime when all items succeed
	if stats.MinTime > 0 && stats.MaxTime > 0 {
		if stats.AvgTime < stats.MinTime || stats.AvgTime > stats.MaxTime {
			t.Errorf("AvgTime(%v) should be between MinTime(%v) and MaxTime(%v)",
				stats.AvgTime, stats.MinTime, stats.MaxTime)
		}
	}
	// All results should be non-nil
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	for i, r := range results {
		if r.Err != nil {
			t.Errorf("results[%d] error: %v", i, r.Err)
		}
	}
}

func TestProcessWithStats_WithFailures(t *testing.T) {
	gen, err := qrcode.New(qrcode.WithDefaultSize(256))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	defer func() { _ = gen.Close() }()

	p := NewProcessor(gen, WithBatchFormat(qrcode.FormatPNG))
	items := []Item{
		{ID: "ok", Data: "valid-data"},
		{ID: "bad", Data: ""}, // empty data → TextPayload validates empty → fails
		{ID: "ok2", Data: "also-valid"},
	}
	results, stats, err := p.ProcessWithStats(context.Background(), items)
	// ProcessWithStats returns a batch error since there's at least one failure.
	if err == nil {
		t.Error("expected batch error when some items fail")
	}
	if stats == nil {
		t.Fatal("stats should not be nil")
	}
	if stats.Total != 3 {
		t.Errorf("stats.Total = %d, want 3", stats.Total)
	}
	if stats.Succeeded != 2 {
		t.Errorf("stats.Succeeded = %d, want 2", stats.Succeeded)
	}
	if stats.Failed != 1 {
		t.Errorf("stats.Failed = %d, want 1", stats.Failed)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	// Item at index 1 should have failed
	if results[1].Err == nil {
		t.Error("results[1] (empty data) should have an error")
	}
	// Items 0 and 2 should succeed
	if results[0].Err != nil {
		t.Errorf("results[0] error: %v", results[0].Err)
	}
	if results[2].Err != nil {
		t.Errorf("results[2] error: %v", results[2].Err)
	}
}

// ---------------------------------------------------------------------------
// Process with mixed valid/invalid items
// ---------------------------------------------------------------------------

func TestProcess_MixedValidInvalid(t *testing.T) {
	gen, err := qrcode.New(qrcode.WithDefaultSize(256))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	defer func() { _ = gen.Close() }()

	p := NewProcessor(gen, WithBatchFormat(qrcode.FormatPNG))
	items := []Item{
		{ID: "good1", Data: "hello"},
		{ID: "empty", Data: ""},
		{ID: "good2", Data: "world"},
		{ID: "empty2", Data: ""},
		{ID: "good3", Data: "foo"},
	}
	results, err := p.Process(context.Background(), items)
	if err == nil {
		t.Error("expected error when some items fail")
	}
	if len(results) != 5 {
		t.Fatalf("expected 5 results, got %d", len(results))
	}
	// Check successes
	for _, idx := range []int{0, 2, 4} {
		if results[idx].Err != nil {
			t.Errorf("results[%d] should succeed, got error: %v", idx, results[idx].Err)
		}
	}
	// Check failures
	for _, idx := range []int{1, 3} {
		if results[idx].Err == nil {
			t.Errorf("results[%d] (empty data) should have failed", idx)
		}
	}
}

// ---------------------------------------------------------------------------
// FromJSON additional tests
// ---------------------------------------------------------------------------

func TestFromJSON_EmptyArray(t *testing.T) {
	p := newTestProcessor(t)
	items, err := p.FromJSON(context.Background(), strings.NewReader(`[]`))
	if err != nil {
		t.Fatalf("FromJSON([]) error: %v", err)
	}
	if items == nil {
		t.Fatal("expected non-nil slice for empty array")
	}
	if len(items) != 0 {
		t.Errorf("expected 0 items, got %d", len(items))
	}
}

func TestFromJSON_WithFormatField(t *testing.T) {
	// The jsonItem struct has a "format" field which is parsed but not used.
	p := newTestProcessor(t)
	input := `[{"id":"fmt1","data":"hello","format":"text"},{"id":"fmt2","data":"world","format":"url"}]`
	items, err := p.FromJSON(context.Background(), strings.NewReader(input))
	if err != nil {
		t.Fatalf("FromJSON() error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].ID != "fmt1" || items[0].Data != "hello" {
		t.Errorf("item[0] = %+v, want ID=fmt1 Data=hello", items[0])
	}
	if items[1].ID != "fmt2" || items[1].Data != "world" {
		t.Errorf("item[1] = %+v, want ID=fmt2 Data=world", items[1])
	}
	// Payload should be nil since FromJSON doesn't set it
	if items[0].Payload != nil {
		t.Error("FromJSON should not set Payload")
	}
}

func TestFromJSON_SingleItem(t *testing.T) {
	p := newTestProcessor(t)
	input := `[{"id":"solo","data":"single"}]`
	items, err := p.FromJSON(context.Background(), strings.NewReader(input))
	if err != nil {
		t.Fatalf("FromJSON() error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].ID != "solo" || items[0].Data != "single" {
		t.Errorf("item[0] = %+v", items[0])
	}
}

func TestFromJSON_WithoutID(t *testing.T) {
	p := newTestProcessor(t)
	input := `[{"data":"no-id"}]`
	items, err := p.FromJSON(context.Background(), strings.NewReader(input))
	if err != nil {
		t.Fatalf("FromJSON() error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].ID != "" {
		t.Errorf("expected empty ID, got %q", items[0].ID)
	}
	if items[0].Data != "no-id" {
		t.Errorf("expected Data='no-id', got %q", items[0].Data)
	}
}

// ---------------------------------------------------------------------------
// FromCSV additional tests
// ---------------------------------------------------------------------------

func TestFromCSV_OnlyDataColumn(t *testing.T) {
	p := newTestProcessor(t)
	input := "data\nhello\nworld\n"
	items, err := p.FromCSV(context.Background(), strings.NewReader(input))
	if err != nil {
		t.Fatalf("FromCSV() error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].Data != "hello" {
		t.Errorf("item[0].Data = %q, want hello", items[0].Data)
	}
	if items[0].ID != "" {
		t.Errorf("item[0].ID should be empty, got %q", items[0].ID)
	}
}

func TestFromCSV_HeaderOnly(t *testing.T) {
	p := newTestProcessor(t)
	input := "id,data\n"
	items, err := p.FromCSV(context.Background(), strings.NewReader(input))
	if err != nil {
		t.Fatalf("FromCSV() error: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected 0 items for header-only CSV, got %d", len(items))
	}
}

func TestFromCSV_BlankRows(t *testing.T) {
	// Blank/whitespace-only rows should be skipped.
	p := newTestProcessor(t)
	input := "id,data\nrow1,val1\n\nrow2,val2\n  \n"
	items, err := p.FromCSV(context.Background(), strings.NewReader(input))
	if err != nil {
		t.Fatalf("FromCSV() error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items (blank rows skipped), got %d", len(items))
	}
	if items[0].ID != "row1" || items[0].Data != "val1" {
		t.Errorf("item[0] = %+v", items[0])
	}
	if items[1].ID != "row2" || items[1].Data != "val2" {
		t.Errorf("item[1] = %+v", items[1])
	}
}

func TestFromCSV_CaseInsensitiveHeaders(t *testing.T) {
	p := newTestProcessor(t)
	input := "ID,DATA\na1,alpha\n"
	items, err := p.FromCSV(context.Background(), strings.NewReader(input))
	if err != nil {
		t.Fatalf("FromCSV() error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].ID != "a1" || items[0].Data != "alpha" {
		t.Errorf("item[0] = %+v", items[0])
	}
}

func TestFromCSV_ExtraColumns(t *testing.T) {
	// Extra columns beyond id and data should be ignored.
	p := newTestProcessor(t)
	input := "id,data,extra,ignored\nr1,val1,x,y\n"
	items, err := p.FromCSV(context.Background(), strings.NewReader(input))
	if err != nil {
		t.Fatalf("FromCSV() error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].ID != "r1" || items[0].Data != "val1" {
		t.Errorf("item[0] = %+v", items[0])
	}
}

// ---------------------------------------------------------------------------
// SaveToDir with custom storage
// ---------------------------------------------------------------------------

func TestSaveToDir_CustomStorage(t *testing.T) {
	mockStore := &mockStorage{saved: make(map[string][]byte)}
	gen, err := qrcode.New(qrcode.WithDefaultSize(256))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	defer func() { _ = gen.Close() }()

	p := NewProcessor(gen, WithBatchStorage(mockStore))
	items := []Item{
		{ID: "mock-qr", Data: "mock-storage-test"},
	}
	results, procErr := p.SaveToDir(context.Background(), items, "/mock/dir")
	if procErr != nil {
		t.Fatalf("SaveToDir() error: %v", procErr)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Err != nil {
		t.Fatalf("result error: %v", results[0].Err)
	}
	// Verify the mock storage was called
	if len(mockStore.saved) == 0 {
		t.Fatal("mock storage should have saved a file")
	}
	expectedKey := "/mock/dir/mock-qr.png"
	if _, ok := mockStore.saved[expectedKey]; !ok {
		// Print what was actually saved
		for k := range mockStore.saved {
			t.Errorf("expected key %q, found key %q", expectedKey, k)
		}
	}
	if len(mockStore.saved[expectedKey]) == 0 {
		t.Error("saved data should not be empty")
	}
}

func TestSaveToDir_MultipleItems(t *testing.T) {
	tmpDir := t.TempDir()
	p := newTestProcessor(t)
	items := []Item{
		{ID: "multi1", Data: "first"},
		{ID: "multi2", Data: "second"},
		{ID: "multi3", Data: "third"},
	}
	results, err := p.SaveToDir(context.Background(), items, tmpDir)
	if err != nil {
		t.Fatalf("SaveToDir() error: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	for i, r := range results {
		if r.Err != nil {
			t.Errorf("results[%d] error: %v", i, r.Err)
		}
		expectedPath := filepath.Join(tmpDir, items[i].ID+".png")
		if _, statErr := os.Stat(expectedPath); os.IsNotExist(statErr) {
			t.Errorf("expected file at %s for item %d", expectedPath, i)
		}
	}
}

func TestSaveToDir_Empty(t *testing.T) {
	p := newTestProcessor(t)
	results, err := p.SaveToDir(context.Background(), nil, "/tmp/empty")
	if err != nil {
		t.Error("SaveToDir(nil) should not error")
	}
	if results != nil {
		t.Errorf("expected nil results, got %v", results)
	}
}

// ---------------------------------------------------------------------------
// QuickBatch additional tests
// ---------------------------------------------------------------------------

func TestQuickBatch_DefaultSize(t *testing.T) {
	// QuickBatch without size parameter should use default 256.
	results, err := QuickBatch(context.Background(), []string{"default-size-test"})
	if err != nil {
		t.Fatalf("QuickBatch() error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if len(results[0]) == 0 {
		t.Error("result should contain PNG data")
	}
}

func TestQuickBatch_WithEmptyString(t *testing.T) {
	// Empty string should produce nil in output and a batch error.
	results, err := QuickBatch(context.Background(), []string{"valid", ""})
	// Should return a batch error since one item fails
	if err == nil {
		t.Error("expected error when one item has empty data")
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if len(results[0]) == 0 {
		t.Error("results[0] (valid) should have data")
	}
	if results[1] != nil {
		t.Errorf("results[1] (empty) should be nil, got %d bytes", len(results[1]))
	}
}

// ---------------------------------------------------------------------------
// Package-level convenience functions
// ---------------------------------------------------------------------------

func TestBatchGenerateWithStats(t *testing.T) {
	gen, err := qrcode.New(qrcode.WithDefaultSize(256))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	defer func() { _ = gen.Close() }()

	items := []Item{
		{ID: "bg1", Data: "batch-gen-test"},
	}
	results, stats, err := BatchGenerateWithStats(context.Background(), gen, items, WithBatchFormat(qrcode.FormatPNG))
	if err != nil {
		t.Fatalf("BatchGenerateWithStats() error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Err != nil {
		t.Errorf("result error: %v", results[0].Err)
	}
	if stats == nil {
		t.Fatal("stats should not be nil")
	}
	if stats.Total != 1 || stats.Succeeded != 1 || stats.Failed != 0 {
		t.Errorf("unexpected stats: %+v", stats)
	}
}

func TestBatchSaveToDir(t *testing.T) {
	tmpDir := t.TempDir()
	gen, err := qrcode.New(qrcode.WithDefaultSize(256))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	defer func() { _ = gen.Close() }()

	items := []Item{
		{ID: "conv-qr", Data: "convenience-fn"},
	}
	results, procErr := BatchSaveToDir(context.Background(), gen, items, tmpDir)
	if procErr != nil {
		t.Fatalf("BatchSaveToDir() error: %v", procErr)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	expectedPath := filepath.Join(tmpDir, "conv-qr.png")
	if _, statErr := os.Stat(expectedPath); os.IsNotExist(statErr) {
		t.Errorf("expected file at %s", expectedPath)
	}
}

// ---------------------------------------------------------------------------
// Concurrent / race condition test
// ---------------------------------------------------------------------------

func TestConcurrentProcess(t *testing.T) {
	gen, err := qrcode.New(qrcode.WithDefaultSize(256))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	defer func() { _ = gen.Close() }()

	p := NewProcessor(gen, WithBatchFormat(qrcode.FormatPNG))
	items := []Item{
		{ID: "race-a", Data: "data-a"},
		{ID: "race-b", Data: "data-b"},
	}

	const goroutines = 10
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			results, procErr := p.Process(context.Background(), items)
			if procErr != nil {
				t.Errorf("concurrent Process error: %v", procErr)
				return
			}
			if len(results) != 2 {
				t.Errorf("expected 2 results, got %d", len(results))
			}
			for j, r := range results {
				if r.Err != nil {
					t.Errorf("results[%d] error: %v", j, r.Err)
				}
				if r.QRCode == nil {
					t.Errorf("results[%d] QRCode is nil", j)
				}
			}
		}()
	}

	wg.Wait()
}

func TestConcurrentProcessWithStats(t *testing.T) {
	gen, err := qrcode.New(qrcode.WithDefaultSize(256))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	defer func() { _ = gen.Close() }()

	p := NewProcessor(gen, WithBatchFormat(qrcode.FormatPNG))
	items := []Item{
		{ID: "stats-race-1", Data: "concurrent-stats"},
		{ID: "stats-race-2", Data: "concurrent-stats-2"},
	}

	const goroutines = 5
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			results, stats, procErr := p.ProcessWithStats(context.Background(), items)
			if procErr != nil {
				t.Errorf("concurrent ProcessWithStats error: %v", procErr)
				return
			}
			if len(results) != 2 {
				t.Errorf("expected 2 results, got %d", len(results))
			}
			if stats == nil {
				t.Error("stats should not be nil")
				return
			}
			if stats.Total != 2 {
				t.Errorf("stats.Total = %d, want 2", stats.Total)
			}
		}()
	}

	wg.Wait()
}

// ---------------------------------------------------------------------------
// Processor with custom payload types
// ---------------------------------------------------------------------------

func TestProcess_VariousPayloadTypes(t *testing.T) {
	gen, err := qrcode.New(qrcode.WithDefaultSize(256))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	defer func() { _ = gen.Close() }()

	p := NewProcessor(gen)
	items := []Item{
		{Payload: &payload.EmailPayload{To: "test@example.com", Subject: "Hello", Body: "World"}},
		{Payload: &payload.PhonePayload{Number: "+1234567890"}},
		{Payload: &payload.SMSPayload{Phone: "+1234567890", Message: "Hi"}},
		{Payload: &payload.GeoPayload{Latitude: 40.7128, Longitude: -74.0060}},
		{Payload: &payload.TextPayload{Text: "plain text payload"}},
	}
	results, procErr := p.Process(context.Background(), items)
	if procErr != nil {
		t.Fatalf("Process() error: %v", procErr)
	}
	if len(results) != 5 {
		t.Fatalf("expected 5 results, got %d", len(results))
	}
	for i, r := range results {
		if r.Err != nil {
			t.Errorf("results[%d] error: %v", i, r.Err)
		}
		if r.QRCode == nil {
			t.Errorf("results[%d] QRCode is nil", i)
		}
	}
}

// ---------------------------------------------------------------------------
// Time-based tests (ensure timing stats are correct)
// ---------------------------------------------------------------------------

func TestProcessWithStats_TotalTimeIsRealistic(t *testing.T) {
	gen, err := qrcode.New(qrcode.WithDefaultSize(256))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	defer func() { _ = gen.Close() }()

	p := NewProcessor(gen, WithBatchFormat(qrcode.FormatPNG), WithBatchConcurrency(1))
	items := []Item{
		{ID: "time-1", Data: "timing-realistic"},
	}

	start := time.Now()
	_, stats, err := p.ProcessWithStats(context.Background(), items)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("ProcessWithStats() error: %v", err)
	}
	if stats == nil {
		t.Fatal("stats should not be nil")
	}
	// TotalTime should be less than the wall-clock time (we only measure processing)
	if stats.TotalTime > elapsed+time.Millisecond {
		t.Errorf("stats.TotalTime=%v should be <= wall-clock=%v", stats.TotalTime, elapsed)
	}
	// But TotalTime should be > 0
	if stats.TotalTime <= 0 {
		t.Error("stats.TotalTime should be positive")
	}
}

// ---------------------------------------------------------------------------
// Mock storage implementation
// ---------------------------------------------------------------------------

type mockStorage struct {
	mu      sync.Mutex
	saved   map[string][]byte
	saveErr error
}

func (s *mockStorage) Save(_ context.Context, path string, data []byte, _ os.FileMode) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.saveErr != nil {
		return s.saveErr
	}
	if s.saved == nil {
		s.saved = make(map[string][]byte)
	}
	s.saved[path] = data
	return nil
}

// ---------------------------------------------------------------------------
// Mock storage that returns errors
// ---------------------------------------------------------------------------

func TestSaveToDir_StorageError(t *testing.T) {
	failStore := &mockStorage{saveErr: os.ErrPermission}
	gen, err := qrcode.New(qrcode.WithDefaultSize(256))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	defer func() { _ = gen.Close() }()

	p := NewProcessor(gen, WithBatchStorage(failStore))
	items := []Item{
		{ID: "fail-qr", Data: "storage-fail-test"},
	}
	results, procErr := p.SaveToDir(context.Background(), items, "/fail/dir")
	// Should return a batch error since the file write fails
	if procErr == nil {
		t.Error("expected error when storage.Save fails")
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Err == nil {
		t.Error("result should have error from storage failure")
	}
}
