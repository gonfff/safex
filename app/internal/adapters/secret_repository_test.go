package adapters

import (
	"context"
	"testing"
	"time"

	"github.com/gonfff/safex/app/internal/domain"
	"github.com/gonfff/safex/app/internal/storage/metadata"
)

// Mock metadata store for testing
type mockMetadataStore struct {
	records map[string]metadata.MetadataRecord
	err     error
}

func newMockMetadataStore() *mockMetadataStore {
	return &mockMetadataStore{
		records: make(map[string]metadata.MetadataRecord),
	}
}

func (m *mockMetadataStore) Create(ctx context.Context, record metadata.MetadataRecord) error {
	if m.err != nil {
		return m.err
	}
	m.records[record.ID] = record
	return nil
}

func (m *mockMetadataStore) Get(ctx context.Context, id string) (metadata.MetadataRecord, error) {
	if m.err != nil {
		return metadata.MetadataRecord{}, m.err
	}
	record, exists := m.records[id]
	if !exists {
		return metadata.MetadataRecord{}, metadata.ErrNotFound
	}
	return record, nil
}

func (m *mockMetadataStore) Delete(ctx context.Context, id string) error {
	if m.err != nil {
		return m.err
	}
	if _, exists := m.records[id]; !exists {
		return metadata.ErrNotFound
	}
	delete(m.records, id)
	return nil
}

func (m *mockMetadataStore) ListExpired(ctx context.Context, before time.Time) ([]metadata.MetadataRecord, error) {
	if m.err != nil {
		return nil, m.err
	}
	expired := make([]metadata.MetadataRecord, 0)
	for _, rec := range m.records {
		if !rec.ExpiresAt.After(before) {
			expired = append(expired, rec)
		}
	}
	return expired, nil
}

func (m *mockMetadataStore) setError(err error) {
	m.err = err
}

func TestSecretRepositoryAdapter_Create(t *testing.T) {
	ctx := context.Background()
	mockStore := newMockMetadataStore()
	adapter := NewSecretRepositoryAdapter(mockStore)

	secret := &domain.Secret{
		ID:           "test-id",
		FileName:     "test.txt",
		ContentType:  "text/plain",
		Size:         100,
		ExpiresAt:    time.Now().Add(time.Hour),
		PayloadType:  domain.PayloadTypeText,
		OpaqueRecord: []byte("opaque-record"),
	}

	err := adapter.Create(ctx, secret)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify record was stored
	if len(mockStore.records) != 1 {
		t.Errorf("expected 1 record, got %d", len(mockStore.records))
	}

	stored := mockStore.records["test-id"]
	if stored.ID != "test-id" {
		t.Errorf("expected ID 'test-id', got %s", stored.ID)
	}
}

func TestSecretRepositoryAdapter_GetByID(t *testing.T) {
	ctx := context.Background()
	mockStore := newMockMetadataStore()
	adapter := NewSecretRepositoryAdapter(mockStore)

	// Setup test data
	record := metadata.MetadataRecord{
		ID:           "test-id",
		FileName:     "test.txt",
		ContentType:  "text/plain",
		Size:         100,
		ExpiresAt:    time.Now().Add(time.Hour),
		PayloadType:  metadata.PayloadTypeText,
		OpaqueRecord: []byte("opaque-record"),
	}
	mockStore.records["test-id"] = record

	secret, err := adapter.GetByID(ctx, "test-id")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if secret.ID != "test-id" {
		t.Errorf("expected ID 'test-id', got %s", secret.ID)
	}
	if secret.PayloadType != domain.PayloadTypeText {
		t.Errorf("expected PayloadType %s, got %s", domain.PayloadTypeText, secret.PayloadType)
	}
}

func TestSecretRepositoryAdapter_GetByID_NotFound(t *testing.T) {
	ctx := context.Background()
	mockStore := newMockMetadataStore()
	adapter := NewSecretRepositoryAdapter(mockStore)

	_, err := adapter.GetByID(ctx, "nonexistent")
	if err != domain.ErrSecretNotFound {
		t.Errorf("expected ErrSecretNotFound, got %v", err)
	}
}

func TestSecretRepositoryAdapter_Delete(t *testing.T) {
	ctx := context.Background()
	mockStore := newMockMetadataStore()
	adapter := NewSecretRepositoryAdapter(mockStore)

	// Setup test data
	record := metadata.MetadataRecord{ID: "test-id"}
	mockStore.records["test-id"] = record

	err := adapter.Delete(ctx, "test-id")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify record was deleted
	if len(mockStore.records) != 0 {
		t.Errorf("expected 0 records, got %d", len(mockStore.records))
	}
}

func TestSecretRepositoryAdapter_Delete_NotFound(t *testing.T) {
	ctx := context.Background()
	mockStore := newMockMetadataStore()
	adapter := NewSecretRepositoryAdapter(mockStore)

	err := adapter.Delete(ctx, "nonexistent")
	if err != domain.ErrSecretNotFound {
		t.Errorf("expected ErrSecretNotFound, got %v", err)
	}
}

func TestSecretRepositoryAdapter_ListExpired(t *testing.T) {
	ctx := context.Background()
	mockStore := newMockMetadataStore()
	adapter := NewSecretRepositoryAdapter(mockStore)

	now := time.Now()
	mockStore.records["expired-id"] = metadata.MetadataRecord{
		ID:        "expired-id",
		ExpiresAt: now.Add(-time.Minute),
	}
	mockStore.records["valid-id"] = metadata.MetadataRecord{
		ID:        "valid-id",
		ExpiresAt: now.Add(time.Hour),
	}

	records, err := adapter.ListExpired(ctx, now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 expired secret, got %d", len(records))
	}
	if records[0].ID != "expired-id" {
		t.Fatalf("expected expired-id, got %s", records[0].ID)
	}
}
