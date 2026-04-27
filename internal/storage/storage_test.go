package storage

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestFileSystem_Save(t *testing.T) {
	ctx := context.Background()
	fs := NewFileSystem()
	tmpDir := t.TempDir()

	path := filepath.Join(tmpDir, "sub", "dir", "test.png")
	data := []byte("fake-png-data")

	if err := fs.Save(ctx, path, data, 0o644); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if string(got) != "fake-png-data" {
		t.Errorf("expected fake-png-data, got %s", string(got))
	}
}

func TestFileSystem_SaveOverwrite(t *testing.T) {
	ctx := context.Background()
	fs := NewFileSystem()
	tmpDir := t.TempDir()

	path := filepath.Join(tmpDir, "overwrite.txt")

	if err := fs.Save(ctx, path, []byte("v1"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := fs.Save(ctx, path, []byte("v2"), 0o644); err != nil {
		t.Fatal(err)
	}

	got, _ := os.ReadFile(path)
	if string(got) != "v2" {
		t.Errorf("expected v2, got %s", string(got))
	}
}

func TestFileSystem_SaveContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	fs := NewFileSystem()
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "cancelled.png")

	err := fs.Save(ctx, path, []byte("data"), 0o644)
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

func TestFileSystem_SaveFilePermissions(t *testing.T) {
	ctx := context.Background()
	fs := NewFileSystem()
	tmpDir := t.TempDir()

	path := filepath.Join(tmpDir, "perms.txt")
	if err := fs.Save(ctx, path, []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf("expected 0600, got %o", info.Mode().Perm())
	}
}

func TestFileSystem_SaveEmptyData(t *testing.T) {
	ctx := context.Background()
	fs := NewFileSystem()
	tmpDir := t.TempDir()

	path := filepath.Join(tmpDir, "empty.txt")
	if err := fs.Save(ctx, path, []byte{}, 0o644); err != nil {
		t.Fatal(err)
	}

	got, _ := os.ReadFile(path)
	if len(got) != 0 {
		t.Errorf("expected empty file, got %d bytes", len(got))
	}
}

func TestFileSystem_SaveCurrentDir(t *testing.T) {
	ctx := context.Background()
	fs := NewFileSystem()
	tmpDir := t.TempDir()

	path := filepath.Join(tmpDir, "direct.txt")
	if err := fs.Save(ctx, path, []byte("direct"), 0o644); err != nil {
		t.Fatal(err)
	}

	got, _ := os.ReadFile(path)
	if string(got) != "direct" {
		t.Errorf("expected 'direct', got %s", string(got))
	}
}

func TestNewFileSystem(t *testing.T) {
	fs := NewFileSystem()
	if fs == nil {
		t.Error("NewFileSystem should not return nil")
	}
	var _ Storage = fs
}

func TestFileSystem_InterfaceCompliance(t *testing.T) {
	var _ Storage = (*FileSystem)(nil)
}

// ---------------------------------------------------------------------------
// Phase 4 additional tests
// ---------------------------------------------------------------------------

func TestFileSystem_SaveInvalidPath(t *testing.T) {
	ctx := context.Background()
	fs := NewFileSystem()

	// Path with null byte should fail.
	err := fs.Save(ctx, "/tmp/\x00invalid.png", []byte("data"), 0o644)
	if err == nil {
		t.Error("expected error for path with null byte")
	}
}

func TestFileSystem_SavePermissionDenied(t *testing.T) {
	ctx := context.Background()
	fs := NewFileSystem()

	// Skip test if running as root (root can write anywhere).
	if os.Getuid() == 0 {
		t.Skip("skipping permission test: running as root")
	}

	// Create a read-only directory.
	readOnlyDir := filepath.Join(t.TempDir(), "readonly")
	if err := os.Mkdir(readOnlyDir, 0o555); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.Chmod(readOnlyDir, 0o755) // restore for cleanup
	}()

	path := filepath.Join(readOnlyDir, "noperm.txt")
	err := fs.Save(ctx, path, []byte("data"), 0o644)
	if err == nil {
		t.Error("expected error for permission denied")
	}
}

func TestFileSystem_SaveAtomic_NoTempFilesLeft(t *testing.T) {
	ctx := context.Background()
	fs := NewFileSystem()
	tmpDir := t.TempDir()

	path := filepath.Join(tmpDir, "atomic.txt")
	data := []byte("atomic write test data")

	if err := fs.Save(ctx, path, data, 0o644); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify the target file exists with correct content.
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if string(got) != "atomic write test data" {
		t.Errorf("expected correct data, got %s", string(got))
	}

	// Verify no temp files remain in the directory.
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if matched, _ := filepath.Match(".qrcode-tmp-*", entry.Name()); matched {
			t.Errorf("temp file left behind: %s", entry.Name())
		}
	}
}

func TestFileSystem_SaveLargeData(t *testing.T) {
	ctx := context.Background()
	fs := NewFileSystem()
	tmpDir := t.TempDir()

	// 1 MB of data.
	data := make([]byte, 1<<20)
	for i := range data {
		data[i] = byte(i % 256)
	}

	path := filepath.Join(tmpDir, "large.bin")
	if err := fs.Save(ctx, path, data, 0o644); err != nil {
		t.Fatalf("Save large file failed: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if len(got) != len(data) {
		t.Errorf("expected %d bytes, got %d", len(data), len(got))
	}
}

func TestFileSystem_SaveConcurrent(t *testing.T) {
	ctx := context.Background()
	fs := NewFileSystem()
	tmpDir := t.TempDir()

	const workers = 10
	var wg sync.WaitGroup
	errs := make([]error, workers)

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			path := filepath.Join(tmpDir, "concurrent.txt")
			data := []byte("data-from-worker")
			errs[idx] = fs.Save(ctx, path, data, 0o644)
		}(i)
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Errorf("worker %d: %v", i, err)
		}
	}

	// Verify file exists.
	got, err := os.ReadFile(filepath.Join(tmpDir, "concurrent.txt"))
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if len(got) == 0 {
		t.Error("concurrent file should not be empty")
	}
}

func TestFileSystem_SaveContextCancelledDuringWrite(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	fs := NewFileSystem()
	tmpDir := t.TempDir()

	// Save first succeeds.
	path1 := filepath.Join(tmpDir, "before.txt")
	if err := fs.Save(ctx, path1, []byte("before"), 0o644); err != nil {
		t.Fatalf("Save before cancel failed: %v", err)
	}

	// Cancel context.
	cancel()

	// Save after cancel should fail.
	path2 := filepath.Join(tmpDir, "after.txt")
	err := fs.Save(ctx, path2, []byte("after"), 0o644)
	if err == nil {
		t.Error("expected error after context cancellation")
	}

	// First file should still be intact.
	got, _ := os.ReadFile(path1)
	if string(got) != "before" {
		t.Errorf("first file should be intact, got %s", string(got))
	}
}
