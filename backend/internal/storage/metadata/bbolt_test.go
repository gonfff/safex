package metadata

import (
	"context"
	"testing"
	"time"
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
