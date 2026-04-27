package renderer

import (
	"bytes"
	"context"
	"fmt"

	"github.com/os-gomod/qrcode/v2/internal/encoding"
)

// TerminalRenderer renders QR codes as terminal/ANSI block characters.
// It is safe for concurrent use.
type TerminalRenderer struct{}

// NewTerminalRenderer creates a new TerminalRenderer.
func NewTerminalRenderer() *TerminalRenderer {
	return &TerminalRenderer{}
}

// Render encodes the QR matrix as terminal output (UTF-8 block characters with optional ANSI colors).
//
//nolint:gocyclo,cyclop // terminal rendering requires per-cell ANSI logic
func (*TerminalRenderer) Render(_ context.Context, qr *encoding.QRCode, opts ...RenderOption) ([]byte, error) {
	cfg := ApplyOptions(opts...)

	if _, _, _, err := ParseHexColor(cfg.ForegroundColor); err != nil {
		return nil, fmt.Errorf("invalid foreground color: %w", err)
	}
	if _, _, _, err := ParseHexColor(cfg.BackgroundColor); err != nil {
		return nil, fmt.Errorf("invalid background color: %w", err)
	}

	totalSize := qr.Size + 2*cfg.QuietZone
	isDark := func(row, col int) bool {
		if row < cfg.QuietZone || row >= cfg.QuietZone+qr.Size {
			return false
		}
		if col < cfg.QuietZone || col >= cfg.QuietZone+qr.Size {
			return false
		}
		return qr.Modules[row-cfg.QuietZone][col-cfg.QuietZone]
	}

	ansiReset := "\033[0m"
	fgR, fgG, fgB := colorToInt(cfg.ForegroundColor)
	bgR, bgG, bgB := colorToInt(cfg.BackgroundColor)
	ansiFg := fmt.Sprintf("\033[38;2;%d;%d;%dm", fgR, fgG, fgB)
	ansiBg := fmt.Sprintf("\033[48;2;%d;%d;%dm", bgR, bgG, bgB)
	useANSI := cfg.ForegroundColor != "#000000" || cfg.BackgroundColor != "#FFFFFF"

	var buf bytes.Buffer
	for row := 0; row < totalSize; row += 2 {
		if useANSI {
			buf.WriteString(ansiBg)
		}
		for col := range totalSize {
			top := isDark(row, col)
			bot := isDark(row+1, col)
			if useANSI {
				buf.WriteString(ansiFg)
			}
			switch {
			case top && bot:
				buf.WriteString("\xe2\x96\x88\xe2\x96\x88")
			case top && !bot:
				buf.WriteString("\xe2\x96\x80\xe2\x96\x80")
			case !top && bot:
				buf.WriteString("\xe2\x96\x84\xe2\x96\x84")
			default:
				buf.WriteString("  ")
			}
		}
		if useANSI {
			buf.WriteString(ansiReset)
		}
		buf.WriteByte('\n')
	}

	return buf.Bytes(), nil
}
