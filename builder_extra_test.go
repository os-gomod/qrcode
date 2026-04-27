package qrcode

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/os-gomod/qrcode/v2/payload"
)

// ---------------------------------------------------------------------------
// Builder fluent method coverage
// ---------------------------------------------------------------------------

func TestBuilder_AllFluentMethods(t *testing.T) {
	// Chain every single builder method to verify they return *Builder and don't panic.
	b := NewBuilder().
		Version(5).
		MinVersion(2).
		MaxVersion(10).
		MaskPattern(3).
		Format(FormatSVG).
		Logo("logo.png", 0.2).
		LogoOverlay(true).
		LogoTint("#FF0000").
		QueueSize(512).
		Prefix("qr_").
		AutoSize(false).
		Options(WithQuietZone(6))

	// Build should succeed.
	client, err := b.Build()
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}
	defer func() { _ = client.Close() }()

	// Verify the config was applied.
	if client.Closed() {
		t.Error("client should not be closed")
	}
}

func TestBuilder_Version(t *testing.T) {
	b := NewBuilder().Version(10)
	if len(b.opts) == 0 {
		t.Error("Version should add an option")
	}
	client, err := b.Build()
	if err != nil {
		t.Fatalf("Build with Version failed: %v", err)
	}
	_ = client.Close()
}

func TestBuilder_MinVersion(t *testing.T) {
	b := NewBuilder().MinVersion(3)
	if len(b.opts) == 0 {
		t.Error("MinVersion should add an option")
	}
	client, err := b.Build()
	if err != nil {
		t.Fatalf("Build with MinVersion failed: %v", err)
	}
	_ = client.Close()
}

func TestBuilder_MaxVersion(t *testing.T) {
	b := NewBuilder().MaxVersion(20)
	if len(b.opts) == 0 {
		t.Error("MaxVersion should add an option")
	}
	client, err := b.Build()
	if err != nil {
		t.Fatalf("Build with MaxVersion failed: %v", err)
	}
	_ = client.Close()
}

func TestBuilder_MaskPattern(t *testing.T) {
	b := NewBuilder().MaskPattern(5)
	if len(b.opts) == 0 {
		t.Error("MaskPattern should add an option")
	}
	client, err := b.Build()
	if err != nil {
		t.Fatalf("Build with MaskPattern failed: %v", err)
	}
	_ = client.Close()
}

func TestBuilder_Format(t *testing.T) {
	b := NewBuilder().Format(FormatPDF)
	if len(b.opts) == 0 {
		t.Error("Format should add an option")
	}
	client, err := b.Build()
	if err != nil {
		t.Fatalf("Build with Format failed: %v", err)
	}
	_ = client.Close()
}

func TestBuilder_Logo(t *testing.T) {
	b := NewBuilder().Logo("test.png", 0.15)
	if len(b.opts) == 0 {
		t.Error("Logo should add an option")
	}
}

func TestBuilder_LogoOverlay(t *testing.T) {
	b := NewBuilder().LogoOverlay(true)
	if len(b.opts) == 0 {
		t.Error("LogoOverlay should add an option")
	}
}

func TestBuilder_LogoTint(t *testing.T) {
	b := NewBuilder().LogoTint("#00FF00")
	if len(b.opts) == 0 {
		t.Error("LogoTint should add an option")
	}
}

func TestBuilder_QueueSize(t *testing.T) {
	b := NewBuilder().QueueSize(2048)
	if len(b.opts) == 0 {
		t.Error("QueueSize should add an option")
	}
	client, err := b.Build()
	if err != nil {
		t.Fatalf("Build with QueueSize failed: %v", err)
	}
	_ = client.Close()
}

func TestBuilder_Prefix(t *testing.T) {
	b := NewBuilder().Prefix("test_")
	if len(b.opts) == 0 {
		t.Error("Prefix should add an option")
	}
}

func TestBuilder_AutoSize(t *testing.T) {
	b := NewBuilder().AutoSize(false)
	if len(b.opts) == 0 {
		t.Error("AutoSize should add an option")
	}
}

func TestBuilder_Options(t *testing.T) {
	b := NewBuilder().Options(WithQuietZone(10), WithDefaultSize(400))
	if len(b.opts) != 2 {
		t.Errorf("Options should add 2 options, got %d", len(b.opts))
	}
}

// ---------------------------------------------------------------------------
// Builder Quick methods coverage
// ---------------------------------------------------------------------------

func TestBuilder_QuickFile(t *testing.T) {
	b := NewBuilder()
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "builder_quick.png")

	err := b.QuickFile("hello builder", path, 256)
	if err != nil {
		t.Fatalf("QuickFile failed: %v", err)
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("file should exist")
	}
}

func TestBuilder_QuickURL(t *testing.T) {
	b := NewBuilder()
	data, err := b.QuickURL("https://example.com", 256)
	if err != nil {
		t.Fatalf("QuickURL failed: %v", err)
	}
	if len(data) == 0 {
		t.Error("output should not be empty")
	}
}

func TestBuilder_QuickWiFi(t *testing.T) {
	b := NewBuilder()
	data, err := b.QuickWiFi("TestNet", "password123", "WPA2", 256)
	if err != nil {
		t.Fatalf("QuickWiFi failed: %v", err)
	}
	if len(data) == 0 {
		t.Error("output should not be empty")
	}
}

func TestBuilder_QuickContact(t *testing.T) {
	b := NewBuilder()
	data, err := b.QuickContact("John", "Doe", "+1234", "j@doe.com", 256)
	if err != nil {
		t.Fatalf("QuickContact failed: %v", err)
	}
	if len(data) == 0 {
		t.Error("output should not be empty")
	}
}

func TestBuilder_QuickSMS(t *testing.T) {
	b := NewBuilder()
	data, err := b.QuickSMS("+1234567890", "Hello", 256)
	if err != nil {
		t.Fatalf("QuickSMS failed: %v", err)
	}
	if len(data) == 0 {
		t.Error("output should not be empty")
	}
}

func TestBuilder_QuickEmail(t *testing.T) {
	b := NewBuilder()
	data, err := b.QuickEmail("test@example.com", "Hi", "Body", 256)
	if err != nil {
		t.Fatalf("QuickEmail failed: %v", err)
	}
	if len(data) == 0 {
		t.Error("output should not be empty")
	}
}

func TestBuilder_QuickGeo(t *testing.T) {
	b := NewBuilder()
	data, err := b.QuickGeo(37.7749, -122.4194, 256)
	if err != nil {
		t.Fatalf("QuickGeo failed: %v", err)
	}
	if len(data) == 0 {
		t.Error("output should not be empty")
	}
}

func TestBuilder_QuickEvent(t *testing.T) {
	b := NewBuilder()
	start := time.Date(2026, 6, 15, 10, 0, 0, 0, time.UTC)
	end := time.Date(2026, 6, 15, 11, 0, 0, 0, time.UTC)
	data, err := b.QuickEvent("Meeting", "Room 1", start, end, 256)
	if err != nil {
		t.Fatalf("QuickEvent failed: %v", err)
	}
	if len(data) == 0 {
		t.Error("output should not be empty")
	}
}

// ---------------------------------------------------------------------------
// Client edge cases
// ---------------------------------------------------------------------------

func TestGenerator_GenerateWithOptions_EmbedLogo(t *testing.T) {
	client, err := New(WithLogo("nonexistent.png", 0.2))
	if err != nil {
		t.Fatalf("New with logo option failed: %v", err)
	}
	defer func() { _ = client.Close() }()

	ctx := context.Background()
	qr, err := client.GenerateWithOptions(ctx, &payload.TextPayload{Text: "logo test"})
	if err != nil {
		t.Fatalf("GenerateWithOptions with logo failed: %v", err)
	}
	if qr == nil {
		t.Error("QR should not be nil")
	}
}

func TestGenerator_GenerateToWriter_Terminal(t *testing.T) {
	client := MustNew()
	defer func() { _ = client.Close() }()

	ctx := context.Background()
	qr, err := client.Generate(ctx, &payload.TextPayload{Text: "terminal"})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	_ = qr
}

func TestGenerator_Save_Subdirectory(t *testing.T) {
	client := MustNew()
	defer func() { _ = client.Close() }()

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "sub", "dir", "test.png")

	ctx := context.Background()
	err := client.Save(ctx, &payload.TextPayload{Text: "save test"}, path)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("file should exist after Save")
	}
}

func TestGenerator_Batch_EmptyItems(t *testing.T) {
	client := MustNew()
	defer func() { _ = client.Close() }()

	ctx := context.Background()
	results, err := client.Batch(ctx, []payload.Payload{})
	if err != nil {
		t.Fatalf("Batch with empty items failed: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

// ---------------------------------------------------------------------------
// Race condition: concurrent generator usage
// ---------------------------------------------------------------------------

func TestGenerator_ConcurrentRender(t *testing.T) {
	client := MustNew()
	defer func() { _ = client.Close() }()

	ctx := context.Background()
	p := &payload.TextPayload{Text: "race test"}

	const goroutines = 10
	errCh := make(chan error, goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			_, err := client.Render(ctx, p, FormatPNG)
			errCh <- err
		}()
	}

	for i := 0; i < goroutines; i++ {
		if err := <-errCh; err != nil {
			t.Errorf("concurrent Render failed: %v", err)
		}
	}
}

func TestGenerator_ConcurrentGenerate(t *testing.T) {
	client := MustNew()
	defer func() { _ = client.Close() }()

	ctx := context.Background()

	const goroutines = 10
	errCh := make(chan error, goroutines)

	for i := 0; i < goroutines; i++ {
		go func(idx int) {
			p := &payload.TextPayload{Text: "concurrent-" + string(rune('A'+idx))}
			_, err := client.Generate(ctx, p)
			errCh <- err
		}(i)
	}

	for i := 0; i < goroutines; i++ {
		if err := <-errCh; err != nil {
			t.Errorf("concurrent Generate failed: %v", err)
		}
	}
}
