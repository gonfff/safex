package metadata

import (
	"context"
	"errors"
	"os"
	"strconv"
	"testing"
	"time"
)

const (
	testAddrEnv     = "SAFEX_TEST_REDIS_ADDR"
	testPasswordEnv = "SAFEX_TEST_REDIS_PASSWORD"
	testDBEnv       = "SAFEX_TEST_REDIS_DB"
)

func newRedisStore(t *testing.T) *RedisStore {
	t.Helper()
	cfg := loadRedisTestConfig(t)
	store, err := NewRedis(cfg)
	if err != nil {
		t.Skipf("redis unavailable at %s: %v", cfg.Addr, err)
	}
	ctx := context.Background()
	if err := store.client.FlushDB(ctx).Err(); err != nil {
		t.Fatalf("flush redis: %v", err)
	}
	t.Cleanup(func() {
		_ = store.client.FlushDB(ctx).Err()
		_ = store.client.Close()
	})
	return store
}

func TestRedisStoreCRUD(t *testing.T) {
	store := newRedisStore(t)
	ctx := context.Background()
	rec := MetadataRecord{
		ID:          "redis-id",
		FileName:    "cipher.bin",
		ContentType: "application/octet-stream",
		Size:        64,
		TTLSeconds:  60,
		ExpiresAt:   time.Now().Add(time.Minute),
	}
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
	if _, err := store.Get(ctx, rec.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func loadRedisTestConfig(t *testing.T) RedisConfig {
	t.Helper()

	cfg := RedisConfig{
		Addr:     fallbackEnv(testAddrEnv, "localhost:6379"),
		Password: fallbackEnv(testPasswordEnv, ""),
		DB:       getIntEnv(testDBEnv, 0),
	}
	return cfg
}

func fallbackEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getIntEnv(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		v, err := strconv.Atoi(val)
		if err == nil {
			return v
		}
	}
	return defaultVal
}
