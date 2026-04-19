package renderer

import (
	"context"
	"fmt"
	"io"

	"github.com/os-gomod/qrcode/encoding"
)

// TerminalRenderer renders QR codes as Unicode block characters suitable for
// terminal output.
//
// The renderer uses a pair of Unicode block characters per column to represent
// two rows of modules at once, halving the vertical line count:
//
//	██  – both top and bottom modules are dark
//	▀▀  – top module dark, bottom module light
//	▄▄  – top module light, bottom module dark
//	    – both modules light (two spaces)
//
// When the foreground or background color differs from the defaults (black/white),
// ANSI 24-bit true-color escape sequences are automatically applied:
//
//	\033[38;2;R;G;Bm  – set foreground color
//	\033[48;2;R;G;Bm  – set background color
//	\033[0m           – reset all attributes
//
// Example:
//
//	r := renderer.NewTerminalRenderer()
//	err := r.Render(ctx, qr, os.Stdout,
//	    renderer.WithForegroundColor("#00FF00"),
//	    renderer.WithBackgroundColor("#000000"),
//	)
type TerminalRenderer struct{}

// NewTerminalRenderer creates a new TerminalRenderer. The returned renderer is
// stateless and safe for concurrent use.
func NewTerminalRenderer() *TerminalRenderer {
	return &TerminalRenderer{}
}

// Type returns the format identifier "terminal".
func (r *TerminalRenderer) Type() string { return "terminal" }

// ContentType returns the MIME type "text/plain".
func (r *TerminalRenderer) ContentType() string { return "text/plain" }

// Render writes the QR code as Unicode block characters to w using the given
// render options.
//
// The output uses two rows of Unicode block characters per two rows of QR
// modules. Each column emits two characters (to maintain a roughly square
// aspect ratio in most terminal fonts). ANSI color codes are prepended
// when non-default colors are specified.
func (r *TerminalRenderer) Render(_ context.Context, qr *encoding.QRCode, w io.Writer, opts ...RenderOption) error {
	cfg := ApplyOptions(opts...)
	_, _, _, err := ParseHexColor(cfg.ForegroundColor)
	if err != nil {
		return fmt.Errorf("invalid foreground color: %w", err)
	}
	_, _, _, err = ParseHexColor(cfg.BackgroundColor)
	if err != nil {
		return fmt.Errorf("invalid background color: %w", err)
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
	fgR, fgG, fgB := colorComponents(cfg.ForegroundColor)
	bgR, bgG, bgB := colorComponents(cfg.BackgroundColor)
	ansiFg := fmt.Sprintf("\033[38;2;%d;%d;%dm", fgR, fgG, fgB)
	ansiBg := fmt.Sprintf("\033[48;2;%d;%d;%dm", bgR, bgG, bgB)
	useANSI := cfg.ForegroundColor != "#000000" || cfg.BackgroundColor != "#FFFFFF"
	var buf []rune
	for row := 0; row < totalSize; row += 2 {
		if useANSI {
			buf = append(buf, []rune(ansiBg)...)
		}
		for col := 0; col < totalSize; col++ {
			top := isDark(row, col)
			bot := isDark(row+1, col)
			if useANSI {
				buf = append(buf, []rune(ansiFg)...)
			}
			switch {
			case top && bot:
				buf = append(buf, '█', '█')
			case top && !bot:
				buf = append(buf, '▀', '▀')
			case !top && bot:
				buf = append(buf, '▄', '▄')
			default:
				buf = append(buf, ' ', ' ')
			}
		}
		if useANSI {
			buf = append(buf, []rune(ansiReset)...)
		}
		buf = append(buf, '\n')
	}
	_, err = w.Write([]byte(string(buf)))
	return err
}

func colorComponents(hex string) (int, int, int) {
	r, g, b, err := ParseHexColor(hex)
	if err != nil {
		return 0, 0, 0
	}
	return int(r), int(g), int(b)
}
