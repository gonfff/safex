package blob

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestLocalStore(t *testing.T) {
	dir := t.TempDir()
	store, err := NewLocal(dir)
	if err != nil {
		t.Fatalf("new local store: %v", err)
	}
	ctx := context.Background()
	key := "abc"
	expected := []byte("ciphertext")
	if err := store.Put(ctx, key, expected); err != nil {
		t.Fatalf("put: %v", err)
	}
	path := filepath.Join(dir, key)
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("stat: %v", err)
	}
	got, err := store.Get(ctx, key)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if string(got) != string(expected) {
		t.Fatalf("unexpected payload: %q", got)
	}
	if err := store.Delete(ctx, key); err != nil {
		t.Fatalf("delete: %v", err)
	}
}
