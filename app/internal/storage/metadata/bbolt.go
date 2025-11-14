package metadata

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.etcd.io/bbolt"
	bboltErrors "go.etcd.io/bbolt/errors"
)

const bucketName = "secrets"

// BoltStore stores metadata inside an embedded Bolt DB file.
type BoltStore struct {
	db *bbolt.DB
}

// NewBolt opens (or creates) a Bolt DB database for metadata persistence.
func NewBolt(path string) (*BoltStore, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create metadata dir: %w", err)
	}
	db, err := bbolt.Open(path, 0o600, &bbolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, fmt.Errorf("open bbolt: %w", err)
	}

	if err := db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		return err
	}); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("init bucket: %w", err)
	}
	return &BoltStore{db: db}, nil
}

// Close releases the underlying DB.
func (s *BoltStore) Close() error {
	return s.db.Close()
}

// Create inserts a metadata record.
func (s *BoltStore) Create(_ context.Context, rec MetadataRecord) error {
	payload, err := json.Marshal(rec)
	if err != nil {
		return err
	}
	return s.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketName))
		return bucket.Put([]byte(rec.ID), payload)
	})
}

// Get fetches a metadata record and enforces TTL.
func (s *BoltStore) Get(_ context.Context, id string) (MetadataRecord, error) {
	var rec MetadataRecord
	err := s.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketName))
		if bucket == nil {
			return ErrNotFound
		}
		blob := bucket.Get([]byte(id))
		if blob == nil {
			return ErrNotFound
		}
		return json.Unmarshal(blob, &rec)
	})
	if err != nil {
		return MetadataRecord{}, translateError(err)
	}

	return rec, nil
}

// Delete removes a metadata record.
func (s *BoltStore) Delete(_ context.Context, id string) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketName))
		if bucket == nil {
			return nil
		}
		return bucket.Delete([]byte(id))
	})
}

func translateError(err error) error {
	if errors.Is(err, bboltErrors.ErrBucketNotFound) {
		return ErrNotFound
	}
	return err
}
