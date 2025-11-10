package blob

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// S3Config hosts configuration for the S3 blob store.
type S3Config struct {
	Bucket    string
	Endpoint  string
	AccessKey string
	SecretKey string
	Region    string
	UseSSL    bool
}

// S3Store persists payloads inside an S3-compatible bucket.
type S3Store struct {
	bucket string
	client *minio.Client
}

// NewS3 initializes an S3 store.
func NewS3(cfg S3Config) (*S3Store, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
		Region: cfg.Region,
	})
	if err != nil {
		log.Fatalln(err)
	}
	return &S3Store{
		bucket: cfg.Bucket,
		client: client,
	}, nil
}

// Put uploads the object to the configured bucket.
func (s *S3Store) Put(ctx context.Context, key string, data []byte) error {
	_, err := s.client.PutObject(ctx, s.bucket, key, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{
		ContentType: "application/octet-stream",
	})
	if err != nil {
		return fmt.Errorf("put object: %w", err)
	}
	return nil
}

// Get downloads an object fully into memory.
func (s *S3Store) Get(ctx context.Context, key string) ([]byte, error) {
	obj, err := s.client.GetObject(ctx, s.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("get object: %w", err)
	}
	defer obj.Close()
	data, readErr := io.ReadAll(obj)
	if readErr != nil {
		return nil, fmt.Errorf("read object: %w", readErr)
	}
	return data, nil
}

// Delete removes an object from the bucket.
func (s *S3Store) Delete(ctx context.Context, key string) error {
	err := s.client.RemoveObject(ctx, s.bucket, key, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("remove object: %w", err)
	}
	return nil
}
