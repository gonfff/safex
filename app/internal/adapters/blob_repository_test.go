package adapters

import (
	"context"
	"errors"
	"testing"
)

// Mock blob store for testing
type mockBlobStore struct {
	data map[string][]byte
	err  error
}

func newMockBlobStore() *mockBlobStore {
	return &mockBlobStore{
		data: make(map[string][]byte),
	}
}

func (m *mockBlobStore) Put(ctx context.Context, key string, data []byte) error {
	if m.err != nil {
		return m.err
	}
	m.data[key] = data
	return nil
}

func (m *mockBlobStore) Get(ctx context.Context, key string) ([]byte, error) {
	if m.err != nil {
		return nil, m.err
	}
	data, exists := m.data[key]
	if !exists {
		return nil, errors.New("not found")
	}
	return data, nil
}

func (m *mockBlobStore) Delete(ctx context.Context, key string) error {
	if m.err != nil {
		return m.err
	}
	delete(m.data, key)
	return nil
}

func (m *mockBlobStore) setError(err error) {
	m.err = err
}

func TestBlobRepositoryAdapter_Store(t *testing.T) {
	ctx := context.Background()
	mockStore := newMockBlobStore()
	adapter := NewBlobRepositoryAdapter(mockStore)

	data := []byte("test data")
	err := adapter.Store(ctx, "test-id", data)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify data was stored
	if len(mockStore.data) != 1 {
		t.Errorf("expected 1 item, got %d", len(mockStore.data))
	}

	stored := mockStore.data["test-id"]
	if string(stored) != "test data" {
		t.Errorf("expected 'test data', got %s", string(stored))
	}
}

func TestBlobRepositoryAdapter_Store_Error(t *testing.T) {
	ctx := context.Background()
	mockStore := newMockBlobStore()
	mockStore.setError(errors.New("store error"))
	adapter := NewBlobRepositoryAdapter(mockStore)

	err := adapter.Store(ctx, "test-id", []byte("data"))
	if err == nil {
		t.Error("expected error but got none")
	}
}

func TestBlobRepositoryAdapter_Retrieve(t *testing.T) {
	ctx := context.Background()
	mockStore := newMockBlobStore()
	adapter := NewBlobRepositoryAdapter(mockStore)

	// Setup test data
	testData := []byte("test data")
	mockStore.data["test-id"] = testData

	retrieved, err := adapter.Retrieve(ctx, "test-id")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if string(retrieved) != "test data" {
		t.Errorf("expected 'test data', got %s", string(retrieved))
	}
}

func TestBlobRepositoryAdapter_Retrieve_Error(t *testing.T) {
	ctx := context.Background()
	mockStore := newMockBlobStore()
	mockStore.setError(errors.New("retrieve error"))
	adapter := NewBlobRepositoryAdapter(mockStore)

	_, err := adapter.Retrieve(ctx, "test-id")
	if err == nil {
		t.Error("expected error but got none")
	}
}

func TestBlobRepositoryAdapter_Remove(t *testing.T) {
	ctx := context.Background()
	mockStore := newMockBlobStore()
	adapter := NewBlobRepositoryAdapter(mockStore)

	// Setup test data
	mockStore.data["test-id"] = []byte("test data")

	err := adapter.Remove(ctx, "test-id")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify data was removed
	if len(mockStore.data) != 0 {
		t.Errorf("expected 0 items, got %d", len(mockStore.data))
	}
}

func TestBlobRepositoryAdapter_Remove_Error(t *testing.T) {
	ctx := context.Background()
	mockStore := newMockBlobStore()
	mockStore.setError(errors.New("remove error"))
	adapter := NewBlobRepositoryAdapter(mockStore)

	err := adapter.Remove(ctx, "test-id")
	if err == nil {
		t.Error("expected error but got none")
	}
}