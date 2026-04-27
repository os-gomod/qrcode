package qrcode

import (
	"bytes"
	"context"
	"testing"

	"github.com/os-gomod/qrcode/v2/internal/encoding"
	"github.com/os-gomod/qrcode/v2/payload"
)

func BenchmarkQuick(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = Quick("benchmark-test-data")
	}
}

func BenchmarkQuickSVG(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = QuickSVG("benchmark-test-data")
	}
}

func BenchmarkGenerate(b *testing.B) {
	client := MustNew()
	defer func() { _ = client.Close() }()
	ctx := context.Background()
	p := &payload.TextPayload{Text: "benchmark-test-data"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = client.Generate(ctx, p)
	}
}

func BenchmarkGenerateWithOptions(b *testing.B) {
	client := MustNew()
	defer func() { _ = client.Close() }()
	ctx := context.Background()
	p := &payload.TextPayload{Text: "benchmark-test-data"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = client.GenerateWithOptions(ctx, p, WithErrorCorrection(LevelH))
	}
}

func BenchmarkRender_PNG(b *testing.B) {
	client := MustNew()
	defer func() { _ = client.Close() }()
	ctx := context.Background()
	p := &payload.TextPayload{Text: "benchmark-render-png"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = client.Render(ctx, p, FormatPNG)
	}
}

func BenchmarkRender_SVG(b *testing.B) {
	client := MustNew()
	defer func() { _ = client.Close() }()
	ctx := context.Background()
	p := &payload.TextPayload{Text: "benchmark-render-svg"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = client.Render(ctx, p, FormatSVG)
	}
}

func BenchmarkRender_Terminal(b *testing.B) {
	client := MustNew()
	defer func() { _ = client.Close() }()
	ctx := context.Background()
	p := &payload.TextPayload{Text: "benchmark-render-terminal"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = client.Render(ctx, p, FormatTerminal)
	}
}

func BenchmarkGenerateToWriter_SVG(b *testing.B) {
	client := MustNew()
	defer func() { _ = client.Close() }()
	ctx := context.Background()
	p := &payload.TextPayload{Text: "benchmark-writer-svg"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		_ = client.GenerateToWriter(ctx, p, &buf, FormatSVG)
	}
}

func BenchmarkBatch(b *testing.B) {
	payloads := make([]payload.Payload, 50)
	for i := range payloads {
		payloads[i] = &payload.TextPayload{Text: "batch-item"}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client := MustNew(WithWorkerCount(4))
		_, _ = client.Batch(context.Background(), payloads)
		_ = client.Close()
	}
}

func BenchmarkNew(b *testing.B) {
	for i := 0; i < b.N; i++ {
		client, _ := New(WithDefaultSize(256))
		if client != nil {
			_ = client.Close()
		}
	}
}

func BenchmarkBuilder(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client, _ := NewBuilder().Size(512).ErrorCorrection(LevelQ).Build()
		if client != nil {
			_ = client.Close()
		}
	}
}

func BenchmarkEncode(b *testing.B) {
	data := []byte("benchmark-encode-data-for-qrcode-generation")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = encoding.Encode(data, 1, 0)
	}
}
