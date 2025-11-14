package secret

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/rs/zerolog"

	"github.com/gonfff/safex/backend/internal/storage"
	"github.com/gonfff/safex/backend/internal/storage/metadata"
)

func TestServiceCreateSuccess(t *testing.T) {
	blobStore := newMockBlobStore()
	metaStore := newMockMetadataStore()
	service := newTestService(blobStore, metaStore)

	input := CreateInput{
		FileName:     "data.txt",
		ContentType:  "text/plain",
		Payload:      []byte("super secret"),
		TTL:          30 * time.Second,
		PayloadType:  metadata.PayloadTypeText,
		OpaqueRecord: []byte("opaque"),
	}

	record, err := service.Create(context.Background(), input)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if record.ID == "" {
		t.Fatalf("expected random ID to be generated")
	}
	if record.FileName != input.FileName {
		t.Fatalf("FileName mismatch: got %s want %s", record.FileName, input.FileName)
	}
	if record.ContentType != input.ContentType {
		t.Fatalf("ContentType mismatch: got %s want %s", record.ContentType, input.ContentType)
	}
	if record.Size != int64(len(input.Payload)) {
		t.Fatalf("Size mismatch: got %d want %d", record.Size, len(input.Payload))
	}
	if record.ExpiresAt.Before(time.Now()) {
		t.Fatalf("secret should expire in the future, got %v", record.ExpiresAt)
	}
	if record.PayloadType != metadata.PayloadTypeText {
		t.Fatalf("PayloadType mismatch: got %s want %s", record.PayloadType, metadata.PayloadTypeText)
	}
	if string(record.OpaqueRecord) != "opaque" {
		t.Fatalf("expected opaque record to be preserved")
	}
	if len(blobStore.putCalls) != 1 {
		t.Fatalf("expected blob to be stored once, got %d", len(blobStore.putCalls))
	}
	if len(metaStore.createCalls) != 1 {
		t.Fatalf("expected metadata to be stored once, got %d", len(metaStore.createCalls))
	}
}

func TestServiceCreateBlobError(t *testing.T) {
	blobStore := newMockBlobStore()
	metaStore := newMockMetadataStore()
	service := newTestService(blobStore, metaStore)

	blobStore.putErr = errors.New("disk full")

	_, err := service.Create(context.Background(), CreateInput{TTL: time.Minute, OpaqueRecord: []byte("opaque")})
	if err == nil || err.Error() != "save blob: disk full" {
		t.Fatalf("expected blob error to be returned, got %v", err)
	}
	if len(metaStore.createCalls) != 0 {
		t.Fatalf("metadata should not be stored when blob fails")
	}
}

func TestServiceCreateMetadataError(t *testing.T) {
	blobStore := newMockBlobStore()
	metaStore := newMockMetadataStore()
	service := newTestService(blobStore, metaStore)

	metaStore.createErr = errors.New("db down")

	_, err := service.Create(context.Background(), CreateInput{TTL: time.Minute, OpaqueRecord: []byte("opaque")})
	if err == nil || err.Error() != "store metadata: db down" {
		t.Fatalf("expected metadata error to be returned, got %v", err)
	}
	if len(blobStore.deleteCalls) != 1 {
		t.Fatalf("blob should be rolled back when metadata fails")
	}
}

func TestServiceCreateRequiresOpaqueRecord(t *testing.T) {
	blobStore := newMockBlobStore()
	metaStore := newMockMetadataStore()
	service := newTestService(blobStore, metaStore)

	_, err := service.Create(context.Background(), CreateInput{TTL: time.Minute})
	if err == nil || err.Error() != "opaque record is required" {
		t.Fatalf("expected opaque enforcement error, got %v", err)
	}
}

func TestServiceLoadSuccess(t *testing.T) {
	blobStore := newMockBlobStore()
	metaStore := newMockMetadataStore()
	service := newTestService(blobStore, metaStore)

	id := "secret-id"
	payload := []byte("payload")
	meta := metadata.MetadataRecord{
		ID:        id,
		FileName:  "note.txt",
		Size:      7,
		ExpiresAt: time.Now().Add(time.Minute),
	}
	metaStore.records[id] = meta
	blobStore.setData(id, payload)

	record, gotPayload, err := service.Load(context.Background(), id)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if string(gotPayload) != string(payload) {
		t.Fatalf("payload mismatch: got %q want %q", gotPayload, payload)
	}
	if record.ID != meta.ID ||
		record.FileName != meta.FileName ||
		record.Size != meta.Size ||
		!record.ExpiresAt.Equal(meta.ExpiresAt) {
		t.Fatalf("metadata mismatch: got %+v want %+v", record, meta)
	}
	if len(blobStore.deleteCalls) != 0 {
		t.Fatalf("Load should not delete fresh secrets")
	}
}

func TestServiceLoadExpiredSecret(t *testing.T) {
	blobStore := newMockBlobStore()
	metaStore := newMockMetadataStore()
	service := newTestService(blobStore, metaStore)

	id := "expired"
	blobStore.setData(id, []byte("data"))
	metaStore.getOverride = func(string) (metadata.MetadataRecord, error) {
		return metadata.MetadataRecord{
			ID:        id,
			ExpiresAt: time.Now().Add(-time.Minute),
		}, nil
	}

	_, _, err := service.Load(context.Background(), id)
	if err == nil || !errors.Is(err, metadata.ErrExpired) {
		t.Fatalf("expected expiration error, got %v", err)
	}
	if len(metaStore.deleteCalls) != 1 || metaStore.deleteCalls[0] != id {
		t.Fatalf("expected metadata delete for expired secret, got %v", metaStore.deleteCalls)
	}
	if len(blobStore.deleteCalls) != 1 || blobStore.deleteCalls[0] != id {
		t.Fatalf("expected blob delete for expired secret, got %v", blobStore.deleteCalls)
	}
}

func TestServiceLoadMetadataError(t *testing.T) {
	blobStore := newMockBlobStore()
	metaStore := newMockMetadataStore()
	service := newTestService(blobStore, metaStore)

	metaStore.getErr = errors.New("not found")

	_, _, err := service.Load(context.Background(), "id")
	if err == nil || err.Error() != "load metadata: not found" {
		t.Fatalf("expected metadata error, got %v", err)
	}
}

func TestServiceLoadBlobError(t *testing.T) {
	blobStore := newMockBlobStore()
	metaStore := newMockMetadataStore()
	service := newTestService(blobStore, metaStore)

	id := "secret"
	metaStore.records[id] = metadata.MetadataRecord{
		ID:        id,
		ExpiresAt: time.Now().Add(time.Minute),
	}
	blobStore.getErr = errors.New("missing blob")

	_, _, err := service.Load(context.Background(), id)
	if err == nil || err.Error() != "load blob: missing blob" {
		t.Fatalf("expected blob error, got %v", err)
	}
}

func TestServiceDelete(t *testing.T) {
	blobStore := newMockBlobStore()
	metaStore := newMockMetadataStore()
	service := newTestService(blobStore, metaStore)

	if err := service.Delete(context.Background(), "id"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if len(metaStore.deleteCalls) != 1 {
		t.Fatalf("metadata delete not invoked")
	}
	if len(blobStore.deleteCalls) != 1 {
		t.Fatalf("blob delete not invoked")
	}
}

func TestServiceDeleteMetadataError(t *testing.T) {
	blobStore := newMockBlobStore()
	metaStore := newMockMetadataStore()
	service := newTestService(blobStore, metaStore)

	metaStore.deleteErr = errors.New("meta down")

	if err := service.Delete(context.Background(), "id"); err == nil || err.Error() != "meta down" {
		t.Fatalf("expected metadata delete error, got %v", err)
	}
	if len(blobStore.deleteCalls) != 0 {
		t.Fatalf("blob delete should not run when metadata delete fails")
	}
}

func TestServiceDeleteBlobError(t *testing.T) {
	blobStore := newMockBlobStore()
	metaStore := newMockMetadataStore()
	service := newTestService(blobStore, metaStore)

	blobStore.deleteErr = errors.New("blob down")

	err := service.Delete(context.Background(), "id")
	if err == nil || err.Error() != "blob down" {
		t.Fatalf("expected blob delete error, got %v", err)
	}
}

func newTestService(blob storage.BlobStore, meta storage.MetadataStore) *Service {
	return NewService(blob, meta, zerolog.New(io.Discard))
}
