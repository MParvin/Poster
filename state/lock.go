package state

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

// Lock represents an exclusive process lock.
type Lock struct {
	file *os.File
	path string
}

// AcquireLock creates/opens path and takes an exclusive flock.
func AcquireLock(path string) (*Lock, error) {
	if path == "" {
		return nil, fmt.Errorf("lock path is empty")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, fmt.Errorf("create lock directory: %w", err)
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, fmt.Errorf("open lock file %s: %w", path, err)
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("another social_poster process holds the lock (%s): %w", path, err)
	}
	return &Lock{file: f, path: path}, nil
}

// Release unlocks and closes the lock file.
func (l *Lock) Release() error {
	if l == nil || l.file == nil {
		return nil
	}
	err := syscall.Flock(int(l.file.Fd()), syscall.LOCK_UN)
	closeErr := l.file.Close()
	l.file = nil
	if err != nil {
		return err
	}
	return closeErr
}

// ResolveLockPath returns the process lock path beside the state file.
func ResolveLockPath(stateFilePath, repoPath string) string {
	if stateFilePath != "" {
		return stateFilePath + ".lock"
	}
	if repoPath != "" {
		return filepath.Join(repoPath, ".social_poster.lock")
	}
	return ".social_poster.lock"
}
