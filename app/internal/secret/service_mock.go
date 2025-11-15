package secret

import (
	"context"
	"fmt"

	"github.com/gonfff/safex/app/internal/storage/metadata"
)

type mockBlobStore struct {
	putCalls    []blobPutCall
	getCalls    []string
	deleteCalls []string

	putErr    error
	getErr    error
	deleteErr error

	data map[string][]byte
}

type blobPutCall struct {
	key  string
	data []byte
}

func newMockBlobStore() *mockBlobStore {
	return &mockBlobStore{data: make(map[string][]byte)}
}

func (s *mockBlobStore) Put(_ context.Context, key string, data []byte) error {
	s.putCalls = append(s.putCalls, blobPutCall{key: key, data: cloneBytes(data)})
	if s.putErr != nil {
		return s.putErr
	}
	s.data[key] = cloneBytes(data)
	return nil
}

func (s *mockBlobStore) Get(_ context.Context, key string) ([]byte, error) {
	s.getCalls = append(s.getCalls, key)
	if s.getErr != nil {
		return nil, s.getErr
	}
	payload, ok := s.data[key]
	if !ok {
		return nil, fmt.Errorf("blob %s not found", key)
	}
	return cloneBytes(payload), nil
}

func (s *mockBlobStore) Delete(_ context.Context, key string) error {
	s.deleteCalls = append(s.deleteCalls, key)
	if s.deleteErr != nil {
		return s.deleteErr
	}
	delete(s.data, key)
	return nil
}

func (s *mockBlobStore) setData(key string, data []byte) {
	s.data[key] = cloneBytes(data)
}

type mockMetadataStore struct {
	createCalls []metadata.MetadataRecord
	getCalls    []string
	deleteCalls []string

	createErr error
	getErr    error
	deleteErr error

	getOverride func(string) (metadata.MetadataRecord, error)

	records map[string]metadata.MetadataRecord
}

func newMockMetadataStore() *mockMetadataStore {
	return &mockMetadataStore{records: make(map[string]metadata.MetadataRecord)}
}

func (s *mockMetadataStore) Create(_ context.Context, record metadata.MetadataRecord) error {
	s.createCalls = append(s.createCalls, record)
	if s.createErr != nil {
		return s.createErr
	}
	s.records[record.ID] = record
	return nil
}

func (s *mockMetadataStore) Get(_ context.Context, id string) (metadata.MetadataRecord, error) {
	s.getCalls = append(s.getCalls, id)
	if s.getOverride != nil {
		return s.getOverride(id)
	}
	if s.getErr != nil {
		return metadata.MetadataRecord{}, s.getErr
	}
	record, ok := s.records[id]
	if !ok {
		return metadata.MetadataRecord{}, fmt.Errorf("metadata %s not found", id)
	}
	return record, nil
}

func (s *mockMetadataStore) Delete(_ context.Context, id string) error {
	s.deleteCalls = append(s.deleteCalls, id)
	if s.deleteErr != nil {
		return s.deleteErr
	}
	delete(s.records, id)
	return nil
}

func cloneBytes(src []byte) []byte {
	if src == nil {
		return nil
	}
	dst := make([]byte, len(src))
	copy(dst, src)
	return dst
}
