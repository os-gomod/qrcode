package qrcode

import (
	"time"
)

type Option func(*Config)

func WithVersion(v int) Option {
	return func(c *Config) {
		c.DefaultVersion = v
	}
}

func WithMinVersion(v int) Option {
	return func(c *Config) {
		c.MinVersion = v
	}
}

func WithMaxVersion(v int) Option {
	return func(c *Config) {
		c.MaxVersion = v
	}
}

func WithErrorCorrection(level ECLevel) Option {
	return func(c *Config) {
		c.DefaultECLevel = level.String()
	}
}

func WithAutoSize(auto bool) Option {
	return func(c *Config) {
		c.AutoSize = auto
	}
}

func WithWorkerCount(n int) Option {
	return func(c *Config) {
		c.WorkerCount = n
	}
}

func WithQueueSize(n int) Option {
	return func(c *Config) {
		c.QueueSize = n
	}
}

func WithDefaultFormat(f Format) Option {
	return func(c *Config) {
		c.DefaultFormat = f.String()
	}
}

func WithDefaultSize(size int) Option {
	return func(c *Config) {
		c.DefaultSize = size
	}
}

func WithQuietZone(zone int) Option {
	return func(c *Config) {
		c.QuietZone = zone
	}
}

func WithForegroundColor(color string) Option {
	return func(c *Config) {
		c.ForegroundColor = color
	}
}

func WithBackgroundColor(color string) Option {
	return func(c *Config) {
		c.BackgroundColor = color
	}
}

func WithMaskPattern(pattern int) Option {
	return func(c *Config) {
		c.MaskPattern = pattern
	}
}

func WithLogo(logoSource string, sizeRatio float64) Option {
	return func(c *Config) {
		c.LogoSource = logoSource
		c.LogoSizeRatio = sizeRatio
		c.LogoOverlay = true
	}
}

func WithLogoOverlay(enabled bool) Option {
	return func(c *Config) {
		c.LogoOverlay = enabled
	}
}

func WithLogoTint(color string) Option {
	return func(c *Config) {
		c.LogoTint = color
	}
}

func WithPrefix(prefix string) Option {
	return func(c *Config) {
		c.Prefix = prefix
	}
}

func WithSlowOperation(d time.Duration) Option {
	return func(c *Config) {
		c.SlowOperation = d
	}
}
