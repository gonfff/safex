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

func TestNewLocalRequiresRoot(t *testing.T) {
	if _, err := NewLocal(""); err == nil {
		t.Fatalf("expected error for empty root")
	}
}

func TestNewLocalFailsWhenPathIsFile(t *testing.T) {
	dir := t.TempDir()
	occupied := filepath.Join(dir, "file")
	if err := os.WriteFile(occupied, []byte("x"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	if _, err := NewLocal(occupied); err == nil {
		t.Fatalf("expected error when path already occupied by file")
	}
}

func TestLocalStorePutMissingDirectory(t *testing.T) {
	root := filepath.Join(t.TempDir(), "missing")
	store := &LocalStore{root: root}
	if err := store.Put(context.Background(), "key", []byte("data")); err == nil {
		t.Fatalf("expected error when directory is absent")
	}
}

func TestLocalStoreGetMissingFile(t *testing.T) {
	store := &LocalStore{root: t.TempDir()}
	if _, err := store.Get(context.Background(), "missing"); err == nil {
		t.Fatalf("expected error when file is absent")
	}
}

func TestLocalStoreDeleteNonEmptyDirectory(t *testing.T) {
	root := t.TempDir()
	dirKey := filepath.Join("nested", "dir")
	target := filepath.Join(root, dirKey)
	if err := os.MkdirAll(filepath.Join(target, "child"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	store := &LocalStore{root: root}
	if err := store.Delete(context.Background(), dirKey); err == nil {
		t.Fatalf("expected error when removing non-empty directory")
	}
}
