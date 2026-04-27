package qrcode

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/os-gomod/qrcode/v2/payload"
)

// ---------------------------------------------------------------------------
// Config tests
// ---------------------------------------------------------------------------

func TestDefaultConfig(t *testing.T) {
	cfg := defaultConfig()
	if cfg.DefaultECLevel != "M" {
		t.Errorf("DefaultECLevel = %q, want 'M'", cfg.DefaultECLevel)
	}
	if cfg.WorkerCount != 4 {
		t.Errorf("WorkerCount = %d, want 4", cfg.WorkerCount)
	}
	if cfg.DefaultSize != 300 {
		t.Errorf("DefaultSize = %d, want 300", cfg.DefaultSize)
	}
	if cfg.QuietZone != 4 {
		t.Errorf("QuietZone = %d, want 4", cfg.QuietZone)
	}
	if cfg.MaskPattern != -1 {
		t.Errorf("MaskPattern = %d, want -1", cfg.MaskPattern)
	}
	if cfg.AutoSize != true {
		t.Error("AutoSize should be true")
	}
}

func TestConfig_Clone(t *testing.T) {
	orig := defaultConfig()
	clone := orig.Clone()
	if clone == orig {
		t.Error("Clone should return a new pointer")
	}
	clone.DefaultSize = 999
	if orig.DefaultSize == 999 {
		t.Error("modifying clone should not affect original")
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		modify  func(*Config)
		wantErr bool
	}{
		{"valid default", func(c *Config) {}, false},
		{"min > max version", func(c *Config) { c.MinVersion = 10; c.MaxVersion = 5 }, true},
		{"default version out of range", func(c *Config) { c.DefaultVersion = 50 }, true},
		{"worker count too low", func(c *Config) { c.WorkerCount = 0 }, true},
		{"worker count too high", func(c *Config) { c.WorkerCount = 100 }, true},
		{"queue size too low", func(c *Config) { c.QueueSize = 0 }, true},
		{"size too small", func(c *Config) { c.DefaultSize = 50 }, true},
		{"size too large", func(c *Config) { c.DefaultSize = 5000 }, true},
		{"quiet zone negative", func(c *Config) { c.QuietZone = -1 }, true},
		{"quiet zone too large", func(c *Config) { c.QuietZone = 25 }, true},
		{"logo overlay without source", func(c *Config) { c.LogoOverlay = true; c.LogoSource = "" }, true},
		{"logo size ratio too small", func(c *Config) { c.LogoSizeRatio = 0.01 }, true},
		{"logo size ratio too large", func(c *Config) { c.LogoSizeRatio = 0.5 }, true},
		{"mask pattern too high", func(c *Config) { c.MaskPattern = 8 }, true},
		{"mask pattern valid auto", func(c *Config) { c.MaskPattern = -1 }, false},
		{"mask pattern valid", func(c *Config) { c.MaskPattern = 3 }, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := defaultConfig()
			tt.modify(cfg)
			err := cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseECLevel(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"L", 0},
		{"M", 1},
		{"Q", 2},
		{"H", 3},
		{"", -1},
		{"X", -1},
		{"m", -1},
	}
	for _, tt := range tests {
		got := parseECLevel(tt.input)
		if got != tt.want {
			t.Errorf("parseECLevel(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

// ---------------------------------------------------------------------------
// Format / ECLevel tests
// ---------------------------------------------------------------------------

func TestECLevel_String(t *testing.T) {
	tests := []struct {
		l    ECLevel
		want string
	}{
		{LevelL, "L"},
		{LevelM, "M"},
		{LevelQ, "Q"},
		{LevelH, "H"},
		{ECLevel(99), "M"}, // default fallback
	}
	for _, tt := range tests {
		if got := tt.l.String(); got != tt.want {
			t.Errorf("ECLevel(%d).String() = %q, want %q", tt.l, got, tt.want)
		}
	}
}

func TestFormat_String(t *testing.T) {
	tests := []struct {
		f    Format
		want string
	}{
		{FormatPNG, "png"},
		{FormatSVG, "svg"},
		{FormatTerminal, "terminal"},
		{FormatPDF, "pdf"},
		{FormatBase64, "base64"},
		{Format(99), "unknown"},
	}
	for _, tt := range tests {
		if got := tt.f.String(); got != tt.want {
			t.Errorf("Format(%d).String() = %q, want %q", tt.f, got, tt.want)
		}
	}
}

func TestFormat_Extension(t *testing.T) {
	tests := []struct {
		f    Format
		want string
	}{
		{FormatPNG, ".png"},
		{FormatSVG, ".svg"},
		{FormatTerminal, ".txt"},
		{FormatPDF, ".pdf"},
		{FormatBase64, ".b64"},
		{Format(99), ".png"}, // default
	}
	for _, tt := range tests {
		if got := tt.f.Extension(); got != tt.want {
			t.Errorf("Format(%d).Extension() = %q, want %q", tt.f, got, tt.want)
		}
	}
}

// ---------------------------------------------------------------------------
// Helpers tests
// ---------------------------------------------------------------------------

func TestQuickSize(t *testing.T) {
	if quickSize() != 256 {
		t.Errorf("quickSize() = %d, want 256", quickSize())
	}
	if quickSize(0) != 256 {
		t.Errorf("quickSize(0) = %d, want 256", quickSize(0))
	}
	if quickSize(512) != 512 {
		t.Errorf("quickSize(512) = %d, want 512", quickSize(512))
	}
}

// ---------------------------------------------------------------------------
// New / MustNew / Close
// ---------------------------------------------------------------------------

func TestNew(t *testing.T) {
	client, err := New()
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	if client == nil {
		t.Fatal("New() returned nil")
	}
	if client.Closed() {
		t.Error("new client should not be closed")
	}
}

func TestNew_WithOptions(t *testing.T) {
	client, err := New(WithDefaultSize(512), WithErrorCorrection(LevelH))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	defer func() { _ = client.Close() }()
}

func TestNew_InvalidConfig(t *testing.T) {
	_, err := New(WithDefaultSize(50)) // Too small.
	if err == nil {
		t.Error("expected error for invalid config")
	}
}

func TestMustNew(t *testing.T) {
	client := MustNew()
	defer func() { _ = client.Close() }()
	if client == nil {
		t.Fatal("MustNew() returned nil")
	}
}

func TestMustNew_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustNew should panic on invalid config")
		}
	}()
	MustNew(WithDefaultSize(50))
}

func TestNewClient_Alias(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Fatalf("NewClient() error: %v", err)
	}
	defer func() { _ = client.Close() }()
}

func TestMustNewClient_Alias(t *testing.T) {
	client := MustNewClient()
	defer func() { _ = client.Close() }()
	if client == nil {
		t.Fatal("MustNewClient() returned nil")
	}
}

// ---------------------------------------------------------------------------
// Client — the actual client implementation
// ---------------------------------------------------------------------------

func TestGenerate_TextPayload(t *testing.T) {
	client := MustNew()
	defer func() { _ = client.Close() }()
	ctx := context.Background()

	qr, err := client.Generate(ctx, &payload.TextPayload{Text: "Hello"})
	if err != nil {
		t.Fatalf("Generate() error: %v", err)
	}
	if qr == nil {
		t.Fatal("Generate() returned nil QRCode")
	}
	if qr.Version < 1 || qr.Version > 40 {
		t.Errorf("Version = %d, out of range", qr.Version)
	}
	if qr.Size < 21 {
		t.Errorf("Size = %d, too small", qr.Size)
	}
}

func TestGenerate_WithOptions(t *testing.T) {
	client := MustNew()
	defer func() { _ = client.Close() }()
	ctx := context.Background()

	qr, err := client.GenerateWithOptions(ctx, &payload.TextPayload{Text: "test"}, WithErrorCorrection(LevelH))
	if err != nil {
		t.Fatalf("GenerateWithOptions() error: %v", err)
	}
	if qr.ECLevel != 3 {
		t.Errorf("ECLevel = %d, want 3 (H)", qr.ECLevel)
	}
}

func TestGenerate_ClosedClient(t *testing.T) {
	client := MustNew()
	_ = client.Close()
	ctx := context.Background()
	_, err := client.Generate(ctx, &payload.TextPayload{Text: "test"})
	if err == nil {
		t.Error("expected error from closed client")
	}
}

func TestGenerate_CancelledContext(t *testing.T) {
	client := MustNew()
	defer func() { _ = client.Close() }()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := client.Generate(ctx, &payload.TextPayload{Text: "test"})
	if err == nil {
		t.Error("expected error with cancelled context")
	}
}

func TestGenerate_EmptyPayload(t *testing.T) {
	client := MustNew()
	defer func() { _ = client.Close() }()
	ctx := context.Background()
	_, err := client.Generate(ctx, &payload.TextPayload{Text: ""})
	if err == nil {
		t.Error("expected error for empty payload")
	}
}

func TestRender_AllFormats(t *testing.T) {
	client := MustNew()
	defer func() { _ = client.Close() }()
	ctx := context.Background()
	p := &payload.TextPayload{Text: "render-test"}

	formats := []Format{FormatPNG, FormatSVG, FormatTerminal, FormatPDF, FormatBase64}
	for _, f := range formats {
		t.Run(f.String(), func(t *testing.T) {
			data, err := client.Render(ctx, p, f)
			if err != nil {
				t.Fatalf("Render(%s) error: %v", f, err)
			}
			if len(data) == 0 {
				t.Errorf("Render(%s) returned empty data", f)
			}
		})
	}
}

func TestGenerateToWriter(t *testing.T) {
	client := MustNew()
	defer func() { _ = client.Close() }()
	ctx := context.Background()
	p := &payload.TextPayload{Text: "writer-test"}

	var buf strings.Builder
	err := client.GenerateToWriter(ctx, p, &buf, FormatSVG)
	if err != nil {
		t.Fatalf("GenerateToWriter() error: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("GenerateToWriter() should write data")
	}
}

func TestSave(t *testing.T) {
	client := MustNew()
	defer func() { _ = client.Close() }()
	ctx := context.Background()

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.png")
	p := &payload.TextPayload{Text: "save-test"}

	err := client.Save(ctx, p, path)
	if err != nil {
		t.Fatalf("Save() error: %v", err)
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("Save() should create the file")
	}
}

func TestSave_Subdirectory(t *testing.T) {
	client := MustNew()
	defer func() { _ = client.Close() }()
	ctx := context.Background()

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "sub", "dir", "test.svg")
	p := &payload.TextPayload{Text: "save-subdir"}

	err := client.Save(ctx, p, path)
	if err != nil {
		t.Fatalf("Save() with subdirs error: %v", err)
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("Save() should create subdirectories")
	}
}

func TestBatch(t *testing.T) {
	client := MustNew(WithWorkerCount(2))
	defer func() { _ = client.Close() }()
	ctx := context.Background()

	payloads := []payload.Payload{
		&payload.TextPayload{Text: "item1"},
		&payload.TextPayload{Text: "item2"},
		&payload.TextPayload{Text: "item3"},
	}
	results, err := client.Batch(ctx, payloads)
	if err != nil {
		t.Fatalf("Batch() error: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	for i, qr := range results {
		if qr == nil {
			t.Errorf("result[%d] is nil", i)
		}
	}
}

func TestBatch_Empty(t *testing.T) {
	client := MustNew()
	defer func() { _ = client.Close() }()
	results, err := client.Batch(context.Background(), nil)
	if err != nil {
		t.Error("Batch(nil) should not error")
	}
	if results != nil {
		t.Errorf("Batch(nil) should return nil, got %v", results)
	}
}

func TestSetOptions(t *testing.T) {
	client := MustNew()
	defer func() { _ = client.Close() }()

	err := client.SetOptions(WithDefaultSize(512))
	if err != nil {
		t.Fatalf("SetOptions() error: %v", err)
	}

	// Invalid options should fail.
	err = client.SetOptions(WithDefaultSize(50))
	if err == nil {
		t.Error("SetOptions() with invalid config should error")
	}
}

func TestSetOptions_AfterClose(t *testing.T) {
	client := MustNew()
	_ = client.Close()
	err := client.SetOptions(WithDefaultSize(512))
	if err == nil {
		t.Error("SetOptions() after Close() should error")
	}
}

// ---------------------------------------------------------------------------
// Quick helpers
// ---------------------------------------------------------------------------

func TestQuick(t *testing.T) {
	data, err := Quick("hello world")
	if err != nil {
		t.Fatalf("Quick() error: %v", err)
	}
	if len(data) == 0 {
		t.Error("Quick() returned empty data")
	}
}

func TestQuick_CustomSize(t *testing.T) {
	data, err := Quick("test", 512)
	if err != nil {
		t.Fatalf("Quick() error: %v", err)
	}
	if len(data) == 0 {
		t.Error("Quick(512) returned empty data")
	}
}

func TestQuickSVG(t *testing.T) {
	svg, err := QuickSVG("test")
	if err != nil {
		t.Fatalf("QuickSVG() error: %v", err)
	}
	if !strings.Contains(svg, "<svg") {
		t.Error("QuickSVG() should return SVG content")
	}
}

func TestQuickFile(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "quick.png")
	err := QuickFile("file test", path, 256)
	if err != nil {
		t.Fatalf("QuickFile() error: %v", err)
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("QuickFile() should create file")
	}
}

func TestQuickURL(t *testing.T) {
	data, err := QuickURL("https://example.com")
	if err != nil {
		t.Fatalf("QuickURL() error: %v", err)
	}
	if len(data) == 0 {
		t.Error("QuickURL() returned empty data")
	}
}

func TestQuickWiFi(t *testing.T) {
	data, err := QuickWiFi("MyNet", "password", "WPA2")
	if err != nil {
		t.Fatalf("QuickWiFi() error: %v", err)
	}
	if len(data) == 0 {
		t.Error("QuickWiFi() returned empty data")
	}
}

func TestQuickContact(t *testing.T) {
	data, err := QuickContact("John", "Doe", "+1234", "john@doe.com")
	if err != nil {
		t.Fatalf("QuickContact() error: %v", err)
	}
	if len(data) == 0 {
		t.Error("QuickContact() returned empty data")
	}
}

func TestQuickSMS(t *testing.T) {
	data, err := QuickSMS("+1234", "Hello")
	if err != nil {
		t.Fatalf("QuickSMS() error: %v", err)
	}
	if len(data) == 0 {
		t.Error("QuickSMS() returned empty data")
	}
}

func TestQuickEmail(t *testing.T) {
	data, err := QuickEmail("to@example.com", "Subject", "Body")
	if err != nil {
		t.Fatalf("QuickEmail() error: %v", err)
	}
	if len(data) == 0 {
		t.Error("QuickEmail() returned empty data")
	}
}

func TestQuickGeo(t *testing.T) {
	data, err := QuickGeo(37.7749, -122.4194)
	if err != nil {
		t.Fatalf("QuickGeo() error: %v", err)
	}
	if len(data) == 0 {
		t.Error("QuickGeo() returned empty data")
	}
}

func TestQuickEvent(t *testing.T) {
	start := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)
	end := time.Date(2026, 1, 15, 11, 0, 0, 0, time.UTC)
	data, err := QuickEvent("Meeting", "Room 1", start, end)
	if err != nil {
		t.Fatalf("QuickEvent() error: %v", err)
	}
	if len(data) == 0 {
		t.Error("QuickEvent() returned empty data")
	}
}

func TestQuickPayment(t *testing.T) {
	data, err := QuickPayment("user@example.com", "25.00")
	if err != nil {
		t.Fatalf("QuickPayment() error: %v", err)
	}
	if len(data) == 0 {
		t.Error("QuickPayment() returned empty data")
	}
}

// ---------------------------------------------------------------------------
// Builder tests
// ---------------------------------------------------------------------------

func TestBuilder_Default(t *testing.T) {
	b := NewBuilder()
	client, err := b.Build()
	if err != nil {
		t.Fatalf("Builder.Build() error: %v", err)
	}
	defer func() { _ = client.Close() }()
}

func TestBuilder_WithOptions(t *testing.T) {
	b := NewBuilder().Size(512).ErrorCorrection(LevelH).Margin(8)
	client, err := b.Build()
	if err != nil {
		t.Fatalf("Builder.Build() error: %v", err)
	}
	defer func() { _ = client.Close() }()
}

func TestBuilder_MustBuild(t *testing.T) {
	b := NewBuilder()
	client := b.MustBuild()
	defer func() { _ = client.Close() }()
	if client == nil {
		t.Fatal("MustBuild() returned nil")
	}
}

func TestBuilder_MustBuild_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustBuild should panic on invalid config")
		}
	}()
	NewBuilder().Size(50).MustBuild()
}

func TestBuilder_Clone(t *testing.T) {
	b := NewBuilder().Size(512)
	clone := b.Clone()
	if clone == b {
		t.Error("Clone should return a new pointer")
	}
	clone.Size(1024)
	// Original should not be affected (builders accumulate options).
	_ = clone
}

func TestBuilder_Quick(t *testing.T) {
	b := NewBuilder()
	data, err := b.Quick("test")
	if err != nil {
		t.Fatalf("Builder.Quick() error: %v", err)
	}
	if len(data) == 0 {
		t.Error("Builder.Quick() returned empty data")
	}
}

func TestBuilder_QuickSVG(t *testing.T) {
	b := NewBuilder()
	svg, err := b.QuickSVG("test")
	if err != nil {
		t.Fatalf("Builder.QuickSVG() error: %v", err)
	}
	if !strings.Contains(svg, "<svg") {
		t.Error("Builder.QuickSVG() should return SVG content")
	}
}

func TestBuilder_FluentChain(t *testing.T) {
	client, err := NewBuilder().
		Size(512).
		ErrorCorrection(LevelQ).
		Margin(8).
		ForegroundColor("#FF0000").
		BackgroundColor("#FFFFFF").
		WorkerCount(8).
		Build()
	if err != nil {
		t.Fatalf("fluent build error: %v", err)
	}
	defer func() { _ = client.Close() }()
}

// ---------------------------------------------------------------------------
// Context helpers
// ---------------------------------------------------------------------------

func TestContextWithQR(t *testing.T) {
	client := MustNew()
	defer func() { _ = client.Close() }()
	ctx := ContextWithQR(context.Background(), client)

	retrieved, ok := QRFromContext(ctx)
	if !ok {
		t.Fatal("QRFromContext should find the client")
	}
	if retrieved != client {
		t.Error("retrieved client should match the stored client")
	}
}

func TestQRFromContext_Missing(t *testing.T) {
	_, ok := QRFromContext(context.Background())
	if ok {
		t.Error("QRFromContext should return false for empty context")
	}
}

// ---------------------------------------------------------------------------
// Options tests
// ---------------------------------------------------------------------------

func TestAllOptions(t *testing.T) {
	cfg := defaultConfig()
	opts := []Option{
		WithVersion(5),
		WithMinVersion(2),
		WithMaxVersion(20),
		WithErrorCorrection(LevelH),
		WithAutoSize(false),
		WithWorkerCount(16),
		WithQueueSize(2048),
		WithDefaultFormat(FormatSVG),
		WithDefaultSize(1024),
		WithQuietZone(10),
		WithForegroundColor("#FF0000"),
		WithBackgroundColor("#00FF00"),
		WithMaskPattern(3),
		WithLogo("logo.png", 0.25),
		WithLogoOverlay(true),
		WithLogoTint("#0000FF"),
		WithPrefix("qr_"),
		WithSlowOperation(200 * time.Millisecond),
	}
	for _, opt := range opts {
		opt(cfg)
	}
	if cfg.DefaultVersion != 5 {
		t.Errorf("Version = %d", cfg.DefaultVersion)
	}
	if cfg.MaxVersion != 20 {
		t.Errorf("MaxVersion = %d", cfg.MaxVersion)
	}
	if cfg.WorkerCount != 16 {
		t.Errorf("WorkerCount = %d", cfg.WorkerCount)
	}
	if cfg.DefaultSize != 1024 {
		t.Errorf("DefaultSize = %d", cfg.DefaultSize)
	}
	if cfg.LogoOverlay != true {
		t.Error("LogoOverlay should be true")
	}
}
