package batch

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	qrcode "github.com/os-gomod/qrcode/v2"
	"github.com/os-gomod/qrcode/v2/payload"
)

func newTestProcessor(t *testing.T, opts ...ProcessorOption) *Processor {
	t.Helper()
	gen, err := qrcode.New(qrcode.WithDefaultSize(256))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	t.Cleanup(func() { _ = gen.Close() })
	p := NewProcessor(gen, opts...)
	return p
}

func TestProcess_Basic(t *testing.T) {
	p := newTestProcessor(t, WithBatchFormat(qrcode.FormatPNG))
	items := []Item{
		{ID: "a", Data: "hello"},
		{ID: "b", Data: "world"},
	}
	results, err := p.Process(context.Background(), items)
	if err != nil {
		t.Fatalf("Process() error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	for i, r := range results {
		if r.Err != nil {
			t.Errorf("result[%d] error: %v", i, r.Err)
		}
		if r.QRCode == nil {
			t.Errorf("result[%d] QRCode is nil", i)
		}
	}
}

func TestProcess_Empty(t *testing.T) {
	p := newTestProcessor(t)
	results, err := p.Process(context.Background(), nil)
	if err != nil {
		t.Error("Process(nil) should not error")
	}
	if results != nil {
		t.Errorf("Process(nil) should return nil, got %v", results)
	}
}

func TestProcess_WithPayload(t *testing.T) {
	p := newTestProcessor(t)
	items := []Item{
		{Payload: &payload.URLPayload{URL: "https://example.com"}},
		{Payload: &payload.WiFiPayload{SSID: "Test", Password: "pw", Encryption: "WPA2"}},
	}
	results, err := p.Process(context.Background(), items)
	if err != nil {
		t.Fatalf("Process() error: %v", err)
	}
	for i, r := range results {
		if r.Err != nil {
			t.Errorf("result[%d] error: %v", i, r.Err)
		}
	}
}

func TestProcess_CancelledContext(t *testing.T) {
	p := newTestProcessor(t, WithBatchConcurrency(1))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	items := []Item{{ID: "a", Data: strings.Repeat("x", 1000)}}
	_, _ = p.Process(ctx, items)
	// Should return without panicking.
}

func TestProcessWithStats(t *testing.T) {
	p := newTestProcessor(t, WithBatchFormat(qrcode.FormatPNG))
	items := []Item{
		{ID: "a", Data: "stat1"},
		{ID: "b", Data: "stat2"},
	}
	results, stats, err := p.ProcessWithStats(context.Background(), items)
	if err != nil {
		t.Fatalf("ProcessWithStats() error: %v", err)
	}
	if stats == nil {
		t.Fatal("stats should not be nil")
	}
	if stats.Total != 2 {
		t.Errorf("stats.Total = %d, want 2", stats.Total)
	}
	if stats.Succeeded != 2 {
		t.Errorf("stats.Succeeded = %d, want 2", stats.Succeeded)
	}
	if stats.Failed != 0 {
		t.Errorf("stats.Failed = %d, want 0", stats.Failed)
	}
	if stats.TotalTime <= 0 {
		t.Error("stats.TotalTime should be positive")
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestProcessWithStats_Empty(t *testing.T) {
	p := newTestProcessor(t)
	results, stats, err := p.ProcessWithStats(context.Background(), nil)
	if err != nil {
		t.Error("should not error")
	}
	if results != nil {
		t.Error("results should be nil")
	}
	if stats.Total != 0 {
		t.Errorf("stats.Total = %d, want 0", stats.Total)
	}
}

func TestSaveToDir(t *testing.T) {
	tmpDir := t.TempDir()
	p := newTestProcessor(t)
	items := []Item{
		{ID: "qr1", Data: "save-dir-test"},
	}
	results, err := p.SaveToDir(context.Background(), items, tmpDir)
	if err != nil {
		t.Fatalf("SaveToDir() error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Err != nil {
		t.Errorf("result error: %v", results[0].Err)
	}
	expectedPath := filepath.Join(tmpDir, "qr1.png")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("expected file at %s", expectedPath)
	}
}

func TestFromJSON(t *testing.T) {
	p := newTestProcessor(t)
	input := `[{"id":"a","data":"hello"},{"id":"b","data":"world"}]`
	items, err := p.FromJSON(context.Background(), strings.NewReader(input))
	if err != nil {
		t.Fatalf("FromJSON() error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].ID != "a" || items[0].Data != "hello" {
		t.Errorf("item[0] = %+v", items[0])
	}
}

func TestFromJSON_InvalidJSON(t *testing.T) {
	p := newTestProcessor(t)
	_, err := p.FromJSON(context.Background(), strings.NewReader("not json"))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestFromJSON_MissingData(t *testing.T) {
	p := newTestProcessor(t)
	input := `[{"id":"a"}]`
	_, err := p.FromJSON(context.Background(), strings.NewReader(input))
	if err == nil {
		t.Error("expected error for missing 'data' field")
	}
}

func TestFromCSV(t *testing.T) {
	p := newTestProcessor(t)
	input := "id,data\ntest1,hello\ntest2,world\n"
	items, err := p.FromCSV(context.Background(), strings.NewReader(input))
	if err != nil {
		t.Fatalf("FromCSV() error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].ID != "test1" || items[0].Data != "hello" {
		t.Errorf("item[0] = %+v", items[0])
	}
}

func TestFromCSV_NoHeader(t *testing.T) {
	p := newTestProcessor(t)
	input := "id,value\na,b\n"
	_, err := p.FromCSV(context.Background(), strings.NewReader(input))
	if err == nil {
		t.Error("expected error when 'data' column is missing")
	}
}

func TestFromCSV_Empty(t *testing.T) {
	p := newTestProcessor(t)
	items, err := p.FromCSV(context.Background(), strings.NewReader(""))
	if err != nil {
		t.Error("empty CSV should return nil, nil")
	}
	if items != nil {
		t.Errorf("expected nil items, got %v", items)
	}
}

func TestFormatExtension(t *testing.T) {
	tests := []struct {
		f    qrcode.Format
		want string
	}{
		{qrcode.FormatPNG, "png"},
		{qrcode.FormatSVG, "svg"},
		{qrcode.FormatTerminal, "txt"},
		{qrcode.FormatPDF, "pdf"},
		{qrcode.FormatBase64, "b64"},
		{qrcode.Format(99), "png"},
	}
	for _, tt := range tests {
		got := formatExtension(tt.f)
		if got != tt.want {
			t.Errorf("formatExtension(%d) = %q, want %q", tt.f, got, tt.want)
		}
	}
}

func TestBuildBatchError(t *testing.T) {
	results := []Result{
		{ID: "a"},
		{ID: "b", Err: nil},
		{ID: "c", Err: fmt.Errorf("some error")},
	}
	err := buildBatchError(results)
	if err == nil {
		t.Error("should return error when there are failures")
	}
}

func TestResolvePayload(t *testing.T) {
	p := newTestProcessor(t)
	item := Item{Data: "hello"}
	pl := p.resolvePayload(item)
	if _, ok := pl.(*payload.TextPayload); !ok {
		t.Errorf("expected TextPayload, got %T", pl)
	}
	item2 := Item{Payload: &payload.URLPayload{URL: "https://example.com"}}
	pl2 := p.resolvePayload(item2)
	if _, ok := pl2.(*payload.URLPayload); !ok {
		t.Errorf("expected URLPayload, got %T", pl2)
	}
}

func TestQuickBatch(t *testing.T) {
	data := []string{"a", "b", "c"}
	results, err := QuickBatch(context.Background(), data, 128)
	if err != nil {
		t.Fatalf("QuickBatch() error: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	for i, r := range results {
		if len(r) == 0 {
			t.Errorf("result[%d] is empty", i)
		}
	}
}

func TestQuickBatch_Empty(t *testing.T) {
	results, err := QuickBatch(context.Background(), nil)
	if err != nil {
		t.Error("QuickBatch(nil) should not error")
	}
	// QuickBatch returns empty slice (not nil) for empty input.
	if len(results) != 0 {
		t.Errorf("QuickBatch(nil) should return empty, got %d items", len(results))
	}
}

func TestNewProcessor(t *testing.T) {
	gen, _ := qrcode.New()
	defer func() { _ = gen.Close() }()
	p := NewProcessor(gen)
	if p == nil {
		t.Fatal("NewProcessor() returned nil")
	}
	if p.concurrency != defaultConcurrency {
		t.Errorf("default concurrency = %d, want %d", p.concurrency, defaultConcurrency)
	}
}

func TestWithBatchConcurrency(t *testing.T) {
	gen, _ := qrcode.New()
	defer func() { _ = gen.Close() }()
	p := NewProcessor(gen, WithBatchConcurrency(8))
	if p.concurrency != 8 {
		t.Errorf("concurrency = %d, want 8", p.concurrency)
	}
}

func TestWithBatchConcurrency_ZeroIgnored(t *testing.T) {
	gen, _ := qrcode.New()
	defer func() { _ = gen.Close() }()
	p := NewProcessor(gen, WithBatchConcurrency(0))
	if p.concurrency != defaultConcurrency {
		t.Errorf("zero concurrency should be ignored, got %d", p.concurrency)
	}
}

func TestWithBatchFormat(t *testing.T) {
	gen, _ := qrcode.New()
	defer func() { _ = gen.Close() }()
	p := NewProcessor(gen, WithBatchFormat(qrcode.FormatSVG))
	if p.format != qrcode.FormatSVG {
		t.Errorf("format = %d, want SVG", p.format)
	}
	if !p.renderFormat {
		t.Error("renderFormat should be true")
	}
}

func TestWithBatchStorage(t *testing.T) {
	gen, _ := qrcode.New()
	defer func() { _ = gen.Close() }()
	memStore := &testStorage{}
	p := NewProcessor(gen, WithBatchStorage(memStore))
	if p.store != memStore {
		t.Error("store should be set to custom storage")
	}
}

func TestProcessor_SaveToDir_WithFormat(t *testing.T) {
	tmpDir := t.TempDir()
	p := newTestProcessor(t, WithBatchFormat(qrcode.FormatSVG))
	items := []Item{{ID: "svg-test", Data: "svg save test"}}
	results, err := p.SaveToDir(context.Background(), items, tmpDir)
	if err != nil {
		t.Fatalf("SaveToDir() SVG error: %v", err)
	}
	expectedPath := filepath.Join(tmpDir, "svg-test.svg")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("expected SVG file at %s", expectedPath)
	}
	_ = results
}

// testStorage is a mock Storage for testing.
type testStorage struct {
	saved map[string][]byte
}

func (s *testStorage) Save(ctx context.Context, path string, data []byte, perm os.FileMode) error {
	if s.saved == nil {
		s.saved = make(map[string][]byte)
	}
	s.saved[path] = data
	return nil
}
