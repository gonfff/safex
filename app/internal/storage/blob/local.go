package blob

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// LocalStore saves ciphertext files on disk.
type LocalStore struct {
	root string
}

// NewLocal creates a filesystem backed blob store.
func NewLocal(root string) (*LocalStore, error) {
	if root == "" {
		return nil, fmt.Errorf("local blob dir must not be empty")
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, fmt.Errorf("create blob dir: %w", err)
	}
	return &LocalStore{root: root}, nil
}

// Put writes the full payload into a file named by key.
func (s *LocalStore) Put(_ context.Context, key string, data []byte) error {
	path := filepath.Join(s.root, key)
	return os.WriteFile(path, data, 0o600)
}

// Get reads payload bytes back.
func (s *LocalStore) Get(_ context.Context, key string) ([]byte, error) {
	path := filepath.Join(s.root, key)
	return os.ReadFile(path)
}

// Delete removes the stored payload.
func (s *LocalStore) Delete(_ context.Context, key string) error {
	path := filepath.Join(s.root, key)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
