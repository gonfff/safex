package metadata

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"go.etcd.io/bbolt"
)

func TestBoltStoreCRUD(t *testing.T) {
	dbPath := t.TempDir() + "/meta.db"
	store, err := NewBolt(dbPath)
	if err != nil {
		t.Fatalf("new bolt: %v", err)
	}
	rec := MetadataRecord{
		ID:         "id",
		FileName:   "cipher.bin",
		TTLSeconds: 60,
		ExpiresAt:  time.Now().Add(time.Hour),
	}
	ctx := context.Background()
	if err := store.Create(ctx, rec); err != nil {
		t.Fatalf("create: %v", err)
	}
	out, err := store.Get(ctx, rec.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if out.FileName != rec.FileName {
		t.Fatalf("unexpected filename: %s", out.FileName)
	}
	if err := store.Delete(ctx, rec.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := store.Get(ctx, rec.ID); err == nil {
		t.Fatalf("expected error after delete")
	}
}

func TestNewBoltCreateDirError(t *testing.T) {
	restore := swapMkdirAll(func(string, os.FileMode) error {
		return errors.New("mkdir boom")
	})
	defer restore()

	if _, err := NewBolt(filepath.Join(t.TempDir(), "meta.db")); err == nil || !strings.Contains(err.Error(), "create metadata dir") {
		t.Fatalf("expected mkdir error, got %v", err)
	}
}

func TestNewBoltOpenError(t *testing.T) {
	restore := swapOpenBolt(func(string, os.FileMode, *bbolt.Options) (*bbolt.DB, error) {
		return nil, errors.New("open boom")
	})
	defer restore()

	if _, err := NewBolt(filepath.Join(t.TempDir(), "meta.db")); err == nil || !strings.Contains(err.Error(), "open bbolt") {
		t.Fatalf("expected open error, got %v", err)
	}
}

func TestNewBoltInitBucketError(t *testing.T) {
	restore := swapInitBucket(func(*bbolt.DB) error {
		return errors.New("init boom")
	})
	defer restore()

	if _, err := NewBolt(filepath.Join(t.TempDir(), "meta.db")); err == nil || !strings.Contains(err.Error(), "init bucket") {
		t.Fatalf("expected init error, got %v", err)
	}
}

func swapMkdirAll(fn func(string, os.FileMode) error) func() {
	original := mkdirAll
	mkdirAll = fn
	return func() {
		mkdirAll = original
	}
}

func swapOpenBolt(fn func(string, os.FileMode, *bbolt.Options) (*bbolt.DB, error)) func() {
	original := openBolt
	openBolt = fn
	return func() {
		openBolt = original
	}
}

func swapInitBucket(fn func(*bbolt.DB) error) func() {
	original := initBoltBucket
	initBoltBucket = fn
	return func() {
		initBoltBucket = original
	}
}
