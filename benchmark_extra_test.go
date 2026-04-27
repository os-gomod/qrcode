package qrcode

import (
	"bytes"
	"context"
	"testing"

	"github.com/os-gomod/qrcode/v2/internal/encoding"
	"github.com/os-gomod/qrcode/v2/payload"
)

// ---------------------------------------------------------------------------
// Extended benchmarks — encoding engine
// ---------------------------------------------------------------------------

func BenchmarkEncode_Short(b *testing.B) {
	data := []byte("hi")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = encoding.Encode(data, 1, 0)
	}
}

func BenchmarkEncode_Medium(b *testing.B) {
	data := []byte("https://github.com/os-gomod/qrcode")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = encoding.Encode(data, 1, 0)
	}
}

func BenchmarkEncode_Long(b *testing.B) {
	data := bytes.Repeat([]byte("abcdefghij"), 30) // 300 bytes
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = encoding.Encode(data, 1, 0)
	}
}

// ---------------------------------------------------------------------------
// Render benchmarks — all formats
// ---------------------------------------------------------------------------

func BenchmarkRender_PDF(b *testing.B) {
	client := MustNew()
	defer func() { _ = client.Close() }()
	ctx := context.Background()
	p := &payload.TextPayload{Text: "benchmark-pdf"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = client.Render(ctx, p, FormatPDF)
	}
}

func BenchmarkRender_Base64(b *testing.B) {
	client := MustNew()
	defer func() { _ = client.Close() }()
	ctx := context.Background()
	p := &payload.TextPayload{Text: "benchmark-base64"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = client.Render(ctx, p, FormatBase64)
	}
}

func BenchmarkGenerateToWriter_PNG(b *testing.B) {
	client := MustNew()
	defer func() { _ = client.Close() }()
	ctx := context.Background()
	p := &payload.TextPayload{Text: "benchmark-writer-png"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		_ = client.GenerateToWriter(ctx, p, &buf, FormatPNG)
	}
}

// ---------------------------------------------------------------------------
// Payload encode benchmarks
// ---------------------------------------------------------------------------

func BenchmarkPayloadEncode_Text(b *testing.B) {
	p := &payload.TextPayload{Text: "Hello, World!"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = p.Encode()
	}
}

func BenchmarkPayloadEncode_URL(b *testing.B) {
	p := &payload.URLPayload{URL: "https://example.com/page?q=test"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = p.Encode()
	}
}

func BenchmarkPayloadEncode_WiFi(b *testing.B) {
	p := &payload.WiFiPayload{SSID: "MyNetwork", Password: "securepassword", Encryption: "WPA2"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = p.Encode()
	}
}

func BenchmarkPayloadEncode_VCard(b *testing.B) {
	p := &payload.VCardPayload{FirstName: "John", LastName: "Doe", Phone: "+1234", Email: "j@doe.com"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = p.Encode()
	}
}

func BenchmarkPayloadEncode_Calendar(b *testing.B) {
	p := &payload.CalendarPayload{Title: "Meeting"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = p.Encode()
	}
}

// ---------------------------------------------------------------------------
// Batch benchmarks
// ---------------------------------------------------------------------------

func BenchmarkBatch_10Items(b *testing.B) {
	payloads := make([]payload.Payload, 10)
	for i := range payloads {
		payloads[i] = &payload.TextPayload{Text: "batch-10"}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client := MustNew(WithWorkerCount(4))
		_, _ = client.Batch(context.Background(), payloads)
		_ = client.Close()
	}
}

func BenchmarkBatch_100Items(b *testing.B) {
	payloads := make([]payload.Payload, 100)
	for i := range payloads {
		payloads[i] = &payload.TextPayload{Text: "batch-100"}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client := MustNew(WithWorkerCount(8))
		_, _ = client.Batch(context.Background(), payloads)
		_ = client.Close()
	}
}

// ---------------------------------------------------------------------------
// Parallel generate benchmark
// ---------------------------------------------------------------------------

func BenchmarkParallelGenerate(b *testing.B) {
	client := MustNew()
	defer func() { _ = client.Close() }()
	ctx := context.Background()
	p := &payload.TextPayload{Text: "parallel-benchmark"}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = client.Generate(ctx, p)
		}
	})
}
