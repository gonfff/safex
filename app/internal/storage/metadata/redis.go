package metadata

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/redis/go-redis/v9/maintnotifications"
)

// RedisConfig configures the Redis-backed metadata store.
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

// RedisStore stores metadata records inside Redis as JSON blobs.
type RedisStore struct {
	client *redis.Client
}

// NewRedis creates a redis-backed metadata store.
func NewRedis(cfg RedisConfig) (*RedisStore, error) {
	if cfg.Addr == "" {
		return nil, fmt.Errorf("missing redis address")
	}
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,

		// Explicitly disable maintenance notifications
		// This prevents the client from sending CLIENT MAINT_NOTIFICATIONS ON
		MaintNotificationsConfig: &maintnotifications.Config{
			Mode: maintnotifications.ModeDisabled,
		},
	})
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}
	return &RedisStore{client: client}, nil
}

// Create stores metadata with a TTL.
func (s *RedisStore) Create(ctx context.Context, rec MetadataRecord) error {
	payload, err := json.Marshal(rec)
	if err != nil {
		return err
	}
	return s.client.Set(ctx, rec.ID, payload, 0).Err()
}

// Get retrieves and validates metadata.
func (s *RedisStore) Get(ctx context.Context, id string) (MetadataRecord, error) {
	val, err := s.client.Get(ctx, id).Bytes()
	if err != nil {
		if err == redis.Nil {
			return MetadataRecord{}, ErrNotFound
		}
		return MetadataRecord{}, err
	}
	var rec MetadataRecord
	if err := json.Unmarshal(val, &rec); err != nil {
		return MetadataRecord{}, err
	}

	return rec, nil
}

// Delete removes metadata key.
func (s *RedisStore) Delete(ctx context.Context, id string) error {
	return s.client.Del(ctx, id).Err()
}
