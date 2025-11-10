package blob

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
)

const (
	testEndpointEnv = "SAFEX_TEST_S3_ENDPOINT"
	testBucketEnv   = "SAFEX_TEST_S3_BUCKET"
	testAccessEnv   = "SAFEX_TEST_S3_ACCESS_KEY"
	testSecretEnv   = "SAFEX_TEST_S3_SECRET_KEY"
	testRegionEnv   = "SAFEX_TEST_S3_REGION"
	testUseSSLEnv   = "SAFEX_TEST_S3_USE_SSL"
)

func TestS3StoreIntegration(t *testing.T) {
	cfg := loadS3TestConfig(t)

	ctx := context.Background()
	store, err := NewS3(cfg)
	if err != nil {
		t.Fatalf("new s3 store: %v", err)
	}

	if err := ensureBucket(ctx, store, cfg); err != nil {
		t.Fatalf("ensure bucket %q: %v", cfg.Bucket, err)
	}

	key := uuid.New().String()
	payload := []byte("ciphertext payload via s3")

	if err := store.Put(ctx, key, payload); err != nil {
		t.Fatalf("put: %v", err)
	}
	t.Cleanup(func() {
		_ = store.Delete(ctx, key)
	})

	got, err := store.Get(ctx, key)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if !bytes.Equal(got, payload) {
		t.Fatalf("unexpected payload: got %q want %q", got, payload)
	}

	if err := store.Delete(ctx, key); err != nil {
		t.Fatalf("delete: %v", err)
	}

	if _, err := store.Get(ctx, key); err == nil {
		t.Fatalf("expected error when fetching deleted key")
	}
}

func loadS3TestConfig(t *testing.T) S3Config {
	t.Helper()

	cfg := S3Config{
		Bucket:    fallbackEnv(testBucketEnv, "safex-tests"),
		Endpoint:  fallbackEnv(testEndpointEnv, "localhost:9000"),
		AccessKey: fallbackEnv(testAccessEnv, "minioadmin"),
		SecretKey: fallbackEnv(testSecretEnv, "minioadmin"),
		Region:    fallbackEnv(testRegionEnv, "us-east-1"),
		UseSSL:    fallbackEnv(testUseSSLEnv, "false") == "true",
	}
	return cfg
}

func fallbackEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func ensureBucket(ctx context.Context, store *S3Store, cfg S3Config) error {
	exists, err := store.client.BucketExists(ctx, cfg.Bucket)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	return store.client.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{
		Region:        cfg.Region,
		ObjectLocking: false,
	})
}
