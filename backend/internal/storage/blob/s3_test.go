package blob

import (
	"bytes"
	"context"
	"os"
	"strings"
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
	store := newS3Store(t, cfg)

	ensureBucketOrSkip(t, ctx, store, cfg)

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

func TestS3StorePutFailsWhenBucketMissing(t *testing.T) {
	cfg := loadS3TestConfig(t)
	cfg.Bucket = "missing-" + uuid.New().String()
	store := newS3Store(t, cfg)

	err := store.Put(context.Background(), "key", []byte("payload"))
	if err == nil || !strings.Contains(err.Error(), "put object") {
		t.Fatalf("expected put error, got %v", err)
	}
}

func TestS3StoreGetFailsWhenBucketMissing(t *testing.T) {
	cfg := loadS3TestConfig(t)
	cfg.Bucket = "missing-" + uuid.New().String()
	store := newS3Store(t, cfg)

	if _, err := store.Get(context.Background(), "key"); err == nil || (!strings.Contains(err.Error(), "get object") && !strings.Contains(err.Error(), "read object")) {
		t.Fatalf("expected get error, got %v", err)
	}
}

func TestS3StoreDeleteFailsWhenBucketMissing(t *testing.T) {
	cfg := loadS3TestConfig(t)
	cfg.Bucket = "missing-" + uuid.New().String()
	store := newS3Store(t, cfg)

	if err := store.Delete(context.Background(), "key"); err == nil || !strings.Contains(err.Error(), "remove object") {
		t.Fatalf("expected delete error, got %v", err)
	}
}

func TestS3StoreGetFailsWithEmptyBucket(t *testing.T) {
	cfg := loadS3TestConfig(t)
	cfg.Bucket = ""
	store := newS3Store(t, cfg)

	if _, err := store.Get(context.Background(), "key"); err == nil || !strings.Contains(err.Error(), "get object") {
		t.Fatalf("expected bucket validation error, got %v", err)
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

func newS3Store(t *testing.T, cfg S3Config) *S3Store {
	t.Helper()

	store, err := NewS3(cfg)
	if err != nil {
		t.Skipf("s3 unavailable at %s: %v", cfg.Endpoint, err)
	}
	return store
}

func ensureBucketOrSkip(t *testing.T, ctx context.Context, store *S3Store, cfg S3Config) {
	t.Helper()
	if err := ensureBucket(ctx, store, cfg); err != nil {
		t.Skipf("ensure bucket %q: %v", cfg.Bucket, err)
	}
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
