package qrcode

import (
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// ConfigPatch tests (Phase 6 additions)
// ---------------------------------------------------------------------------

func TestApplyPatch_NilPatch(t *testing.T) {
	base := defaultConfig()
	result := ApplyPatch(base, &ConfigPatch{})
	if result.DefaultECLevel != base.DefaultECLevel {
		t.Error("empty patch should not modify base")
	}
	// Verify it's a new copy.
	if result == base {
		t.Error("ApplyPatch should return a new Config, not the same pointer")
	}
}

func TestApplyPatch_PartialOverride(t *testing.T) {
	base := defaultConfig()
	patch := ConfigPatch{
		WorkerCount: IntP(16),
		AutoSize:    BoolP(false),
	}
	result := ApplyPatch(base, &patch)
	if result.WorkerCount != 16 {
		t.Errorf("WorkerCount = %d, want 16", result.WorkerCount)
	}
	if result.AutoSize {
		t.Error("AutoSize should be false")
	}
	// Other fields should be unchanged.
	if result.DefaultECLevel != base.DefaultECLevel {
		t.Error("DefaultECLevel should be unchanged")
	}
	if result.DefaultSize != base.DefaultSize {
		t.Error("DefaultSize should be unchanged")
	}
}

func TestApplyPatch_ZeroValueOverride(t *testing.T) {
	// The key advantage of ConfigPatch: we CAN set fields to zero values.
	base := defaultConfig() // WorkerCount=4, QueueSize=1024
	patch := ConfigPatch{
		WorkerCount: IntP(0),
		QueueSize:   IntP(0),
	}
	result := ApplyPatch(base, &patch)
	if result.WorkerCount != 0 {
		t.Errorf("WorkerCount = %d, want 0 (explicit zero override)", result.WorkerCount)
	}
	if result.QueueSize != 0 {
		t.Errorf("QueueSize = %d, want 0 (explicit zero override)", result.QueueSize)
	}
}

func TestApplyPatch_AllFields(t *testing.T) {
	base := defaultConfig()
	d := 5 * time.Second
	patch := ConfigPatch{
		DefaultVersion:  IntP(5),
		DefaultECLevel:  StringP("H"),
		MinVersion:      IntP(2),
		MaxVersion:      IntP(20),
		AutoSize:        BoolP(false),
		WorkerCount:     IntP(8),
		QueueSize:       IntP(2048),
		DefaultFormat:   StringP("svg"),
		DefaultSize:     IntP(512),
		QuietZone:       IntP(2),
		ForegroundColor: StringP("#FF0000"),
		BackgroundColor: StringP("#00FF00"),
		MaskPattern:     IntP(3),
		LogoSource:      StringP("/path/to/logo.png"),
		LogoSizeRatio:   Float64P(0.3),
		LogoOverlay:     BoolP(true),
		LogoTint:        StringP("#0000FF"),
		Prefix:          StringP("qr-"),
		SlowOperation:   DurationP(d),
	}
	result := ApplyPatch(base, &patch)
	if result.DefaultVersion != 5 {
		t.Errorf("DefaultVersion = %d, want 5", result.DefaultVersion)
	}
	if result.DefaultECLevel != "H" {
		t.Errorf("DefaultECLevel = %q, want %q", result.DefaultECLevel, "H")
	}
	if result.WorkerCount != 8 {
		t.Errorf("WorkerCount = %d, want 8", result.WorkerCount)
	}
	if result.DefaultFormat != "svg" {
		t.Errorf("DefaultFormat = %q, want %q", result.DefaultFormat, "svg")
	}
	if result.SlowOperation != d {
		t.Errorf("SlowOperation = %v, want %v", result.SlowOperation, d)
	}
}

func TestApplyPatch_DoesNotMutateBase(t *testing.T) {
	base := defaultConfig()
	origWorkerCount := base.WorkerCount
	patch := ConfigPatch{WorkerCount: IntP(99)}
	_ = ApplyPatch(base, &patch)
	if base.WorkerCount != origWorkerCount {
		t.Error("ApplyPatch should not mutate the base config")
	}
}

func TestValidatePatch_NilPatch(t *testing.T) {
	err := ValidatePatch(&ConfigPatch{})
	if err != nil {
		t.Errorf("empty patch should be valid, got: %v", err)
	}
}

func TestValidatePatch_InvalidWorkerCount(t *testing.T) {
	err := ValidatePatch(&ConfigPatch{WorkerCount: IntP(0)})
	if err == nil {
		t.Error("should reject worker_count=0")
	}
	err = ValidatePatch(&ConfigPatch{WorkerCount: IntP(100)})
	if err == nil {
		t.Error("should reject worker_count=100")
	}
}

func TestValidatePatch_InvalidQueueSize(t *testing.T) {
	err := ValidatePatch(&ConfigPatch{QueueSize: IntP(0)})
	if err == nil {
		t.Error("should reject queue_size=0")
	}
}

func TestValidatePatch_InvalidDefaultSize(t *testing.T) {
	err := ValidatePatch(&ConfigPatch{DefaultSize: IntP(50)})
	if err == nil {
		t.Error("should reject default_size=50")
	}
}

func TestValidatePatch_InvalidQuietZone(t *testing.T) {
	err := ValidatePatch(&ConfigPatch{QuietZone: IntP(-1)})
	if err == nil {
		t.Error("should reject quiet_zone=-1")
	}
	err = ValidatePatch(&ConfigPatch{QuietZone: IntP(25)})
	if err == nil {
		t.Error("should reject quiet_zone=25")
	}
}

func TestValidatePatch_MinMaxVersion(t *testing.T) {
	err := ValidatePatch(&ConfigPatch{
		MinVersion: IntP(10),
		MaxVersion: IntP(5),
	})
	if err == nil {
		t.Error("should reject min > max")
	}
	// Valid case.
	err = ValidatePatch(&ConfigPatch{
		MinVersion: IntP(1),
		MaxVersion: IntP(10),
	})
	if err != nil {
		t.Errorf("valid min/max should pass, got: %v", err)
	}
}

func TestValidatePatch_LogoOverlay(t *testing.T) {
	err := ValidatePatch(&ConfigPatch{
		LogoOverlay: BoolP(true),
		LogoSource:  StringP(""),
	})
	if err == nil {
		t.Error("should reject logo_overlay=true with empty logo_source")
	}
}

func TestValidatePatch_LogoSizeRatio(t *testing.T) {
	err := ValidatePatch(&ConfigPatch{LogoSizeRatio: Float64P(0.01)})
	if err == nil {
		t.Error("should reject logo_size_ratio=0.01")
	}
}

func TestValidatePatch_MaskPattern(t *testing.T) {
	err := ValidatePatch(&ConfigPatch{MaskPattern: IntP(8)})
	if err == nil {
		t.Error("should reject mask_pattern=8")
	}
	err = ValidatePatch(&ConfigPatch{MaskPattern: IntP(-2)})
	if err == nil {
		t.Error("should reject mask_pattern=-2")
	}
}

func TestValidatePatch_Valid(t *testing.T) {
	patch := ConfigPatch{
		WorkerCount:   IntP(8),
		QueueSize:     IntP(512),
		DefaultSize:   IntP(256),
		QuietZone:     IntP(4),
		MaskPattern:   IntP(3),
		LogoSizeRatio: Float64P(0.2),
	}
	err := ValidatePatch(&patch)
	if err != nil {
		t.Errorf("valid patch should pass, got: %v", err)
	}
}

func TestConfigToPatch(t *testing.T) {
	cfg := defaultConfig()
	patch := ConfigToPatch(cfg)
	if patch.WorkerCount == nil || *patch.WorkerCount != cfg.WorkerCount {
		t.Errorf("WorkerCount not preserved in patch")
	}
	if patch.AutoSize == nil || *patch.AutoSize != cfg.AutoSize {
		t.Error("AutoSize not preserved in patch")
	}
	if patch.DefaultSize == nil || *patch.DefaultSize != cfg.DefaultSize {
		t.Error("DefaultSize not preserved in patch")
	}
	// Round-trip: ApplyPatch should produce identical config.
	restored := ApplyPatch(defaultConfig(), &patch)
	if restored.WorkerCount != cfg.WorkerCount {
		t.Error("round-trip WorkerCount mismatch")
	}
}

// ---------------------------------------------------------------------------
// Pointer helper tests
// ---------------------------------------------------------------------------

func TestPointerHelpers(t *testing.T) {
	if *IntP(42) != 42 {
		t.Error("IntP failed")
	}
	if *StringP("hello") != "hello" {
		t.Error("StringP failed")
	}
	if *BoolP(true) != true {
		t.Error("BoolP failed")
	}
	if *Float64P(3.14) != 3.14 {
		t.Error("Float64P failed")
	}
	d := 5 * time.Second
	if *DurationP(d) != d {
		t.Error("DurationP failed")
	}
}
