// Package storage provides file system abstraction for QR code output.
// It isolates file I/O from business logic, enabling testing and future
// alternative storage backends.
package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// Storage defines the interface for persisting rendered QR code output.
type Storage interface {
	// Save writes data to the given path, creating parent directories as needed.
	Save(ctx context.Context, path string, data []byte, _ os.FileMode) error
}

// FileSystem implements Storage using the local file system.
// Writes are performed atomically: data is first written to a temporary file
// in the same directory, then renamed to the target path.
type FileSystem struct{}

// NewFileSystem creates a new FileSystem storage backend.
func NewFileSystem() *FileSystem {
	return &FileSystem{}
}

// Save writes data to the specified file path atomically. It creates any
// necessary parent directories and respects context cancellation.
//
// Atomicity is achieved by writing to a temporary file in the target
// directory, then renaming it.  On systems where rename is atomic
// (local POSIX filesystems, modern NTFS), this guarantees that readers
// will either see the old file or the new file — never a partial write.
func (*FileSystem) Save(ctx context.Context, path string, data []byte, _ os.FileMode) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	dir := filepath.Dir(path)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("storage: failed to create directory %s: %w", dir, err)
		}
	}

	// Write to a temp file in the same directory to ensure the rename is atomic
	// (rename across filesystems is not guaranteed to be atomic).
	tmpFile, createErr := os.CreateTemp(dir, ".qrcode-tmp-")
	if createErr != nil {
		return fmt.Errorf("storage: failed to create temp file for %s: %w", path, createErr)
	}
	tmpPath := tmpFile.Name()

	success := false
	defer func() {
		if !success {
			_ = os.Remove(tmpPath)
		}
	}()

	// Write data to the temp file.
	if _, err := tmpFile.Write(data); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("storage: failed to write temp file for %s: %w", path, err)
	}

	// Close before rename to flush buffers.
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("storage: failed to close temp file for %s: %w", path, err)
	}

	// Check context again before the rename.
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Atomic rename.
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("storage: failed to rename temp file to %s: %w", path, err)
	}

	success = true
	return nil
}

// Ensure FileSystem satisfies the Storage interface at compile time.
var _ Storage = (*FileSystem)(nil)
