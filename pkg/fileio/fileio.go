// Package fileio provides the foundation's first filesystem primitives:
// read one file telling absence from failure, replace one file's whole
// content atomically and durably, and ensure a directory exists. Consuming
// applications own save and conflict rules. See CONTRACT.md.
package fileio

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"

	faults "github.com/0xsj/atelier-wails/pkg/errors"
)

const (
	filePerm fs.FileMode = 0o600
	dirPerm  fs.FileMode = 0o700
)

// Read returns the file's content, or found false when the path is missing.
func Read(path string) ([]byte, bool, error) {
	data, err := os.ReadFile(path)
	if err == nil {
		return data, true, nil
	}
	if errors.Is(err, fs.ErrNotExist) {
		return nil, false, nil
	}
	return nil, false, failure("fileio.read_failed", "file could not be read", path, err)
}

// Replace writes data to a temporary file beside path, flushes it, renames it
// over path and flushes the directory. The target is always either its
// previous or its new complete content. The parent directory must exist.
func Replace(path string, data []byte) error {
	dir, base := filepath.Split(path)
	if dir == "" {
		dir = "."
	}
	temp, err := tempName(dir, base)
	if err != nil {
		return failure("fileio.write_failed", "file could not be written", path, err)
	}
	if err := writeTemp(temp, data); err != nil {
		_ = os.Remove(temp)
		return failure("fileio.write_failed", "file could not be written", path, err)
	}
	if err := os.Rename(temp, path); err != nil {
		_ = os.Remove(temp)
		return failure("fileio.write_failed", "file could not be written", path, err)
	}
	if err := syncDir(dir); err != nil {
		return failure("fileio.write_failed", "file could not be written", path, err)
	}
	return nil
}

// EnsureDir creates path and its parents; an existing directory is a no-op.
func EnsureDir(path string) error {
	if err := os.MkdirAll(path, dirPerm); err != nil {
		return failure("fileio.dir_failed", "directory could not be created", path, err)
	}
	info, err := os.Stat(path)
	if err != nil {
		return failure("fileio.dir_failed", "directory could not be created", path, err)
	}
	if !info.IsDir() {
		return failure("fileio.dir_failed", "directory could not be created", path, errors.New("path is not a directory"))
	}
	return nil
}

func tempName(dir, base string) (string, error) {
	var suffix [8]byte
	if _, err := rand.Read(suffix[:]); err != nil {
		return "", err
	}
	return filepath.Join(dir, "."+base+"."+hex.EncodeToString(suffix[:])+".tmp"), nil
}

func writeTemp(temp string, data []byte) error {
	file, err := os.OpenFile(temp, os.O_WRONLY|os.O_CREATE|os.O_EXCL, filePerm)
	if err != nil {
		return err
	}
	if _, err := file.Write(data); err != nil {
		_ = file.Close()
		return err
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}

// syncDir flushes the directory entry after a rename where the platform
// supports opening directories for sync.
func syncDir(dir string) error {
	if runtime.GOOS == "windows" {
		return nil
	}
	handle, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer handle.Close()
	return handle.Sync()
}

func failure(typ, message, path string, cause error) error {
	return faults.New(faults.Unavailable, message).
		WithType(typ).
		WithDetail("path", path).
		WithCause(cause)
}
