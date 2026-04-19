package qrcode

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/os-gomod/qrcode/payload"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		opts    []Option
		wantErr bool
	}{
		{name: "no options"},
		{name: "with size", opts: []Option{WithDefaultSize(200)}},
		{name: "with EC level", opts: []Option{WithErrorCorrection(LevelH)}},
		{name: "with cache", opts: []Option{}},
		{name: "invalid worker count", opts: []Option{WithWorkerCount(0)}, wantErr: true},
		{name: "invalid size", opts: []Option{WithDefaultSize(50)}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g, err := New(tt.opts...)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if g == nil {
				t.Fatal("expected non-nil Generator")
			}
			defer g.Close(context.Background())
		})
	}
}

func TestMustNew(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustNew should panic on error")
		}
	}()
	MustNew(WithWorkerCount(0))
}

func TestMustNewValid(t *testing.T) {
	g := MustNew()
	if g == nil {
		t.Fatal("MustNew returned nil")
	}
	defer g.Close(context.Background())
}

func TestErrorCorrectionLevelString(t *testing.T) {
	tests := []struct {
		level ErrorCorrectionLevel
		want  string
	}{
		{LevelL, "L"},
		{LevelM, "M"},
		{LevelQ, "Q"},
		{LevelH, "H"},
		{ErrorCorrectionLevel(99), "M"},
	}

	for _, tt := range tests {
		got := tt.level.String()
		if got != tt.want {
			t.Errorf("Level(%d).String() = %q, want %q", tt.level, got, tt.want)
		}
	}
}

func TestFormatString(t *testing.T) {
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
		got := tt.f.String()
		if got != tt.want {
			t.Errorf("Format(%d).String() = %q, want %q", tt.f, got, tt.want)
		}
	}
}

func TestNewBuilder(t *testing.T) {
	b := NewBuilder()
	if b == nil {
		t.Fatal("NewBuilder returned nil")
	}
}

func TestBuilderBuild(t *testing.T) {
	b := NewBuilder().Size(200)
	g, err := b.Build()
	if err != nil {
		t.Fatalf("Build error: %v", err)
	}
	if g == nil {
		t.Fatal("Build returned nil")
	}
	defer g.Close(context.Background())
}

func TestBuilderClone(t *testing.T) {
	b := NewBuilder().Size(200).Margin(2)
	c := b.Clone()
	if c == nil {
		t.Fatal("Clone returned nil")
	}

	g1, err := b.Build()
	if err != nil {
		t.Fatalf("Build error: %v", err)
	}
	defer g1.Close(context.Background())

	g2, err := c.Build()
	if err != nil {
		t.Fatalf("Cloned Build error: %v", err)
	}
	defer g2.Close(context.Background())
}

func TestQuick(t *testing.T) {
	data, err := Quick("Hello World")
	if err != nil {
		t.Fatalf("Quick error: %v", err)
	}
	if len(data) == 0 {
		t.Error("Quick returned empty data")
	}
}

func TestQuickWithSize(t *testing.T) {
	data, err := Quick("Hello", 512)
	if err != nil {
		t.Fatalf("Quick error: %v", err)
	}
	if len(data) == 0 {
		t.Error("Quick returned empty data")
	}
}

func TestQuickSVG(t *testing.T) {
	svg, err := QuickSVG("Hello World")
	if err != nil {
		t.Fatalf("QuickSVG error: %v", err)
	}
	if !strings.Contains(svg, "<svg") {
		t.Error("QuickSVG should return SVG content")
	}
}

func TestQuickWiFi(t *testing.T) {
	data, err := QuickWiFi("MyNetwork", "password123", "WPA2")
	if err != nil {
		t.Fatalf("QuickWiFi error: %v", err)
	}
	if len(data) == 0 {
		t.Error("QuickWiFi returned empty data")
	}
}

func TestQuickURL(t *testing.T) {
	data, err := QuickURL("https://example.com")
	if err != nil {
		t.Fatalf("QuickURL error: %v", err)
	}
	if len(data) == 0 {
		t.Error("QuickURL returned empty data")
	}
}

func TestGeneratePNG(t *testing.T) {
	g, err := New(WithDefaultSize(200))
	if err != nil {
		t.Fatalf("New error: %v", err)
	}
	defer g.Close(context.Background())

	p, _ := payload.Text("Test PNG")
	data, err := GeneratePNG(context.Background(), g, p)
	if err != nil {
		t.Fatalf("GeneratePNG error: %v", err)
	}
	if len(data) == 0 {
		t.Error("GeneratePNG returned empty data")
	}
}

func TestGenerateSVG(t *testing.T) {
	g, err := New(WithDefaultSize(200))
	if err != nil {
		t.Fatalf("New error: %v", err)
	}
	defer g.Close(context.Background())

	p, _ := payload.Text("Test SVG")
	svg, err := GenerateSVG(context.Background(), g, p)
	if err != nil {
		t.Fatalf("GenerateSVG error: %v", err)
	}
	if !strings.Contains(svg, "<svg") {
		t.Error("GenerateSVG should return SVG content")
	}
}

func TestSavePNG(t *testing.T) {
	g, err := New(WithDefaultSize(200))
	if err != nil {
		t.Fatalf("New error: %v", err)
	}
	defer g.Close(context.Background())

	p, _ := payload.Text("Save Test")
	tmpFile := filepath(t.TempDir(), "test.png")
	err = SavePNG(context.Background(), g, p, tmpFile)
	if err != nil {
		t.Fatalf("SavePNG error: %v", err)
	}
	if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
		t.Error("SavePNG should create file")
	}
}

func filepath(dir, name string) string {
	return dir + string(os.PathSeparator) + name
}

func TestContextWithQR(t *testing.T) {
	g, err := New()
	if err != nil {
		t.Fatalf("New error: %v", err)
	}
	defer g.Close(context.Background())

	ctx := ContextWithQR(context.Background(), g)
	retrieved, ok := QRFromContext(ctx)
	if !ok {
		t.Error("QRFromContext should return true")
	}
	if retrieved == nil {
		t.Error("QRFromContext should return non-nil")
	}

	// Missing context
	_, ok = QRFromContext(context.Background())
	if ok {
		t.Error("QRFromContext should return false for missing context")
	}
}

func TestBuilderQuickMethods(t *testing.T) {
	// Test Builder.Quick
	b := NewBuilder()
	data, err := b.Quick("Hello Builder")
	if err != nil {
		t.Fatalf("Builder.Quick error: %v", err)
	}
	if len(data) == 0 {
		t.Error("Builder.Quick returned empty data")
	}

	// Test Builder.QuickURL
	data, err = b.QuickURL("https://example.com")
	if err != nil {
		t.Fatalf("Builder.QuickURL error: %v", err)
	}
	if len(data) == 0 {
		t.Error("Builder.QuickURL returned empty data")
	}

	// Test Builder.QuickWiFi
	data, err = b.QuickWiFi("SSID", "pass", "WPA2")
	if err != nil {
		t.Fatalf("Builder.QuickWiFi error: %v", err)
	}
	if len(data) == 0 {
		t.Error("Builder.QuickWiFi returned empty data")
	}

	// Test Builder.QuickSVG
	svg, err := b.QuickSVG("Hello SVG")
	if err != nil {
		t.Fatalf("Builder.QuickSVG error: %v", err)
	}
	if !strings.Contains(svg, "<svg") {
		t.Error("Builder.QuickSVG should return SVG")
	}

	// Test Builder.QuickContact
	data, err = b.QuickContact("John", "Doe", "555-1234", "john@example.com")
	if err != nil {
		t.Fatalf("Builder.QuickContact error: %v", err)
	}
	if len(data) == 0 {
		t.Error("Builder.QuickContact returned empty data")
	}

	// Test Builder.QuickSMS
	data, err = b.QuickSMS("+1234567890", "Hello")
	if err != nil {
		t.Fatalf("Builder.QuickSMS error: %v", err)
	}
	if len(data) == 0 {
		t.Error("Builder.QuickSMS returned empty data")
	}

	// Test Builder.QuickEmail
	data, err = b.QuickEmail("user@example.com", "Subject", "Body")
	if err != nil {
		t.Fatalf("Builder.QuickEmail error: %v", err)
	}
	if len(data) == 0 {
		t.Error("Builder.QuickEmail returned empty data")
	}

	// Test Builder.QuickGeo
	data, err = b.QuickGeo(37.77, -122.42)
	if err != nil {
		t.Fatalf("Builder.QuickGeo error: %v", err)
	}
	if len(data) == 0 {
		t.Error("Builder.QuickGeo returned empty data")
	}

	// Test Builder.QuickEvent
	data, err = b.QuickEvent("Meeting", "Office", time.Now(), time.Now().Add(2*time.Hour))
	if err != nil {
		t.Fatalf("Builder.QuickEvent error: %v", err)
	}
	if len(data) == 0 {
		t.Error("Builder.QuickEvent returned empty data")
	}
}

func TestGenerateASCII(t *testing.T) {
	g, err := New(WithDefaultSize(200))
	if err != nil {
		t.Fatalf("New error: %v", err)
	}
	defer g.Close(context.Background())

	p, _ := payload.Text("ASCII")
	ascii, err := GenerateASCII(context.Background(), g, p)
	if err != nil {
		t.Fatalf("GenerateASCII error: %v", err)
	}
	if len(ascii) == 0 {
		t.Error("GenerateASCII returned empty string")
	}
}

func TestGenerateBase64(t *testing.T) {
	g, err := New(WithDefaultSize(200))
	if err != nil {
		t.Fatalf("New error: %v", err)
	}
	defer g.Close(context.Background())

	p, _ := payload.Text("Base64")
	b64, err := GenerateBase64(context.Background(), g, p)
	if err != nil {
		t.Fatalf("GenerateBase64 error: %v", err)
	}
	if !strings.Contains(b64, "data:image/png;base64,") {
		t.Error("GenerateBase64 should return data URL")
	}
}

func TestGeneratorInterface(t *testing.T) {
	g, err := New()
	if err != nil {
		t.Fatalf("New error: %v", err)
	}
	defer g.Close(context.Background())

	ctx := context.Background()
	p, _ := payload.Text("Interface test")

	// Test Generate
	qr, err := g.Generate(ctx, p)
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}
	if qr == nil {
		t.Error("Generate returned nil QRCode")
	}

	// Test Closed
	if g.Closed() {
		t.Error("generator should not be closed")
	}
}

func TestExtensionFromPath(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"/path/to/file.png", ".png"},
		{"/path/to/file.svg", ".svg"},
		{"/path/to/file.PDF", ".pdf"},
		{"/path/to/file", ""},
	}

	for _, tt := range tests {
		got := extensionFromPath(tt.path)
		if got != tt.want {
			t.Errorf("extensionFromPath(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

func TestQuickContact(t *testing.T) {
	data, err := QuickContact("John", "Doe", "555-1234", "john@example.com")
	if err != nil {
		t.Fatalf("QuickContact error: %v", err)
	}
	if len(data) == 0 {
		t.Error("QuickContact returned empty data")
	}
}

func TestQuickSMS(t *testing.T) {
	data, err := QuickSMS("+1234567890", "Hello SMS")
	if err != nil {
		t.Fatalf("QuickSMS error: %v", err)
	}
	if len(data) == 0 {
		t.Error("QuickSMS returned empty data")
	}
}

func TestQuickPayment(t *testing.T) {
	data, err := QuickPayment("user", "10.00")
	if err != nil {
		t.Fatalf("QuickPayment error: %v", err)
	}
	if len(data) == 0 {
		t.Error("QuickPayment returned empty data")
	}
}

func TestBuilderBuildAndGeneratePNG(t *testing.T) {
	b := NewBuilder().Size(200)
	p, _ := payload.Text("BuildAndGenerate")
	data, err := b.BuildAndGeneratePNG(context.Background(), p)
	if err != nil {
		t.Fatalf("BuildAndGeneratePNG error: %v", err)
	}
	if len(data) == 0 {
		t.Error("BuildAndGeneratePNG returned empty data")
	}
}

func TestBuilderBuildAndGenerateSVG(t *testing.T) {
	b := NewBuilder().Size(200)
	p, _ := payload.Text("BuildAndGenerate")
	svg, err := b.BuildAndGenerateSVG(context.Background(), p)
	if err != nil {
		t.Fatalf("BuildAndGenerateSVG error: %v", err)
	}
	if !strings.Contains(svg, "<svg") {
		t.Error("BuildAndGenerateSVG should return SVG")
	}
}

func TestBuilderQuickFile(t *testing.T) {
	b := NewBuilder().Size(200)
	tmpDir := t.TempDir()
	tmpFile := tmpDir + "/test_quick.png"
	err := b.QuickFile("Quick File Test", tmpFile)
	if err != nil {
		t.Fatalf("QuickFile error: %v", err)
	}
	if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
		t.Error("QuickFile should create file")
	}
}

func TestQuickFile(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := tmpDir + "/test_quick2.png"
	err := QuickFile("Quick File Test", tmpFile)
	if err != nil {
		t.Fatalf("QuickFile error: %v", err)
	}
	if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
		t.Error("QuickFile should create file")
	}
}

func TestSave(t *testing.T) {
	g, err := New(WithDefaultSize(200))
	if err != nil {
		t.Fatalf("New error: %v", err)
	}
	defer g.Close(context.Background())

	p, _ := payload.Text("Save Test")

	// Save as PNG
	tmpDir := t.TempDir()
	tmpFile := tmpDir + "/save_test.png"
	err = Save(context.Background(), g, p, tmpFile)
	if err != nil {
		t.Fatalf("Save PNG error: %v", err)
	}

	// Save as SVG
	tmpFile2 := tmpDir + "/save_test.svg"
	err = Save(context.Background(), g, p, tmpFile2)
	if err != nil {
		t.Fatalf("Save SVG error: %v", err)
	}

	// Read and check SVG content
	svgData, _ := os.ReadFile(tmpFile2)
	if !strings.Contains(string(svgData), "<svg") {
		t.Error("saved SVG should contain <svg>")
	}
}

func TestGeneratorBatchEmpty(t *testing.T) {
	g, err := New()
	if err != nil {
		t.Fatalf("New error: %v", err)
	}
	defer g.Close(context.Background())

	results, err := g.Batch(context.Background(), nil)
	if err != nil {
		t.Fatalf("Batch error: %v", err)
	}
	if len(results) != 0 {
		t.Error("empty batch should return nil results")
	}
}

func TestGenerateToWriter(t *testing.T) {
	g, err := New(WithDefaultSize(200))
	if err != nil {
		t.Fatalf("New error: %v", err)
	}
	defer g.Close(context.Background())

	p, _ := payload.Text("Writer test")

	var buf bytes.Buffer
	err = g.GenerateToWriter(context.Background(), p, &buf, FormatPNG)
	if err != nil {
		t.Fatalf("GenerateToWriter error: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("GenerateToWriter should write data")
	}
}

func TestGenerateToWriterMultipleFormats(t *testing.T) {
	g, err := New(WithDefaultSize(200))
	if err != nil {
		t.Fatalf("New error: %v", err)
	}
	defer g.Close(context.Background())

	p, _ := payload.Text("Format test")

	formats := []Format{FormatPNG, FormatSVG, FormatTerminal, FormatPDF, FormatBase64}
	for _, f := range formats {
		var buf bytes.Buffer
		err := g.GenerateToWriter(context.Background(), p, &buf, f)
		if err != nil {
			t.Errorf("GenerateToWriter format %d error: %v", f, err)
		}
		if buf.Len() == 0 {
			t.Errorf("GenerateToWriter format %d should write data", f)
		}
	}
}

func TestGenerateWithOptions(t *testing.T) {
	g, err := New(WithDefaultSize(200), WithErrorCorrection(LevelM))
	if err != nil {
		t.Fatalf("New error: %v", err)
	}
	defer g.Close(context.Background())

	p, _ := payload.Text("WithOptions test")

	// Override to LevelH
	qr, err := g.GenerateWithOptions(context.Background(), p, WithErrorCorrection(LevelH))
	if err != nil {
		t.Fatalf("GenerateWithOptions error: %v", err)
	}
	if qr == nil {
		t.Error("GenerateWithOptions returned nil")
	}
}

func TestGenerateToWriterSVG(t *testing.T) {
	g, err := New(WithDefaultSize(200))
	if err != nil {
		t.Fatalf("New error: %v", err)
	}
	defer g.Close(context.Background())

	p, _ := payload.Text("SVG writer test")
	var buf bytes.Buffer
	err = g.GenerateToWriter(context.Background(), p, &buf, FormatSVG)
	if err != nil {
		t.Fatalf("GenerateToWriter SVG error: %v", err)
	}
	if !strings.Contains(buf.String(), "<svg") {
		t.Error("SVG writer should output SVG content")
	}
}

func TestGenerateToWriterBase64(t *testing.T) {
	g, err := New(WithDefaultSize(200))
	if err != nil {
		t.Fatalf("New error: %v", err)
	}
	defer g.Close(context.Background())

	p, _ := payload.Text("Base64 writer test")
	var buf bytes.Buffer
	err = g.GenerateToWriter(context.Background(), p, &buf, FormatBase64)
	if err != nil {
		t.Fatalf("GenerateToWriter Base64 error: %v", err)
	}
	if !strings.Contains(buf.String(), "data:image/png;base64,") {
		t.Error("Base64 writer should output data URL")
	}
}

func TestGeneratorSetOptions(t *testing.T) {
	g, err := New(WithDefaultSize(200))
	if err != nil {
		t.Fatalf("New error: %v", err)
	}
	defer g.Close(context.Background())

	// Change size
	err = g.SetOptions(WithDefaultSize(400))
	if err != nil {
		t.Fatalf("SetOptions error: %v", err)
	}

	// Invalid option should fail
	if err := g.SetOptions(WithDefaultSize(50)); err == nil {
		t.Error("SetOptions should fail for invalid size")
	}
}

func TestGeneratorSetOptionsAfterClose(t *testing.T) {
	g, err := New(WithDefaultSize(200))
	if err != nil {
		t.Fatalf("New error: %v", err)
	}
	g.Close(context.Background())

	if err := g.SetOptions(WithDefaultSize(300)); err == nil {
		t.Error("SetOptions should fail after Close")
	}
}

func TestGeneratorBatchNonEmpty(t *testing.T) {
	g, err := New(WithDefaultSize(200))
	if err != nil {
		t.Fatalf("New error: %v", err)
	}
	defer g.Close(context.Background())

	p1, _ := payload.Text("Batch1")
	p2, _ := payload.Text("Batch2")
	p3, _ := payload.URL("https://example.com")

	results, err := g.Batch(context.Background(), []payload.Payload{p1, p2, p3})
	if err != nil {
		t.Fatalf("Batch error: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("Batch should return 3 results, got %d", len(results))
	}
	for i, qr := range results {
		if qr == nil {
			t.Errorf("results[%d] is nil", i)
		}
	}
}

func TestGeneratorBatchWithOptions(t *testing.T) {
	g, err := New(WithDefaultSize(200))
	if err != nil {
		t.Fatalf("New error: %v", err)
	}
	defer g.Close(context.Background())

	p1, _ := payload.Text("A")
	p2, _ := payload.Text("B")
	results, err := g.Batch(context.Background(), []payload.Payload{p1, p2},
		WithErrorCorrection(LevelQ),
	)
	if err != nil {
		t.Fatalf("Batch with options error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
}

func TestGeneratorGenerateAfterClose(t *testing.T) {
	g, err := New(WithDefaultSize(200))
	if err != nil {
		t.Fatalf("New error: %v", err)
	}
	g.Close(context.Background())

	if !g.Closed() {
		t.Error("Closed() should return true after Close()")
	}

	p, _ := payload.Text("After close")
	_, err = g.Generate(context.Background(), p)
	if err == nil {
		t.Error("Generate should fail after Close")
	}
}

func TestGeneratorGenerateWithOptionsInvalid(t *testing.T) {
	g, err := New(WithDefaultSize(200))
	if err != nil {
		t.Fatalf("New error: %v", err)
	}
	defer g.Close(context.Background())

	p, _ := payload.Text("Test")
	_, err = g.GenerateWithOptions(context.Background(), p, WithDefaultSize(50))
	if err == nil {
		t.Error("GenerateWithOptions should fail for invalid per-call size")
	}
}

func TestGeneratorSetOptionsMultiple(t *testing.T) {
	g, err := New(WithDefaultSize(200))
	if err != nil {
		t.Fatalf("New error: %v", err)
	}
	defer g.Close(context.Background())

	// Set multiple valid options
	err = g.SetOptions(
		WithDefaultSize(300),
		WithErrorCorrection(LevelQ),
		WithQuietZone(4),
		WithForegroundColor("#FF0000"),
		WithBackgroundColor("#FFFFFF"),
	)
	if err != nil {
		t.Fatalf("SetOptions error: %v", err)
	}

	// Verify generation still works
	p, _ := payload.Text("WithOptions")
	qr, err := g.Generate(context.Background(), p)
	if err != nil {
		t.Fatalf("Generate after SetOptions error: %v", err)
	}
	if qr == nil {
		t.Error("Generate should return non-nil after SetOptions")
	}
}

func TestGeneratorSetOptionsInvalidWorkerCount(t *testing.T) {
	g, err := New(WithDefaultSize(200))
	if err != nil {
		t.Fatalf("New error: %v", err)
	}
	defer g.Close(context.Background())

	err = g.SetOptions(WithWorkerCount(0))
	if err == nil {
		t.Error("SetOptions should fail for worker count 0")
	}
}

func TestGenerateWithDifferentECLevels(t *testing.T) {
	levels := []ErrorCorrectionLevel{LevelL, LevelM, LevelQ, LevelH}
	for _, level := range levels {
		t.Run(level.String(), func(t *testing.T) {
			g, err := New(WithDefaultSize(200), WithErrorCorrection(level))
			if err != nil {
				t.Fatalf("New error: %v", err)
			}
			defer g.Close(context.Background())

			p, _ := payload.Text("EC Level " + level.String())
			qr, err := g.Generate(context.Background(), p)
			if err != nil {
				t.Fatalf("Generate error: %v", err)
			}
			if qr == nil {
				t.Error("Generate returned nil")
			}
		})
	}
}

func TestNewWithVariousOptions(t *testing.T) {
	tests := []struct {
		name string
		opts []Option
	}{
		{name: "with version", opts: []Option{WithVersion(5)}},
		{name: "with auto size", opts: []Option{WithAutoSize(true)}},
		{name: "with colors", opts: []Option{WithForegroundColor("#000"), WithBackgroundColor("#FFF")}},
		{name: "with quiet zone", opts: []Option{WithQuietZone(5)}},
		{name: "with mask pattern", opts: []Option{WithMaskPattern(3)}},
		{name: "with format svg", opts: []Option{WithDefaultFormat(FormatSVG)}},
		{name: "with queue size", opts: []Option{WithQueueSize(100)}},
		{name: "with version range", opts: []Option{WithMinVersion(1), WithMaxVersion(10)}},
		{name: "with concurrency", opts: []Option{WithConcurrency(4)}},
		{name: "with prefix", opts: []Option{WithPrefix("https://example.com/")}},
		{name: "with slow operation", opts: []Option{WithSlowOperation(5 * time.Second)}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g, err := New(tt.opts...)
			if err != nil {
				t.Fatalf("New(%v) error: %v", tt.name, err)
			}
			defer g.Close(context.Background())
			if g == nil {
				t.Fatal("expected non-nil generator")
			}
		})
	}
}
