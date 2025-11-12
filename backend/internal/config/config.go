package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config captures runtime tunables sourced from environment variables.
type Config struct {
	HTTPAddr          string
	MaxPayloadMB      int
	RequestsPerMinute int
	BlobBackend       string
	BlobDir           string
	S3Bucket          string
	S3Endpoint        string
	S3AccessKey       string
	S3SecretKey       string
	S3Region          string
	MetadataBackend   string
	BoltPath          string
	RedisAddr         string
	RedisPassword     string
	RedisDB           int
	DefaultTTL        time.Duration // minutes
	MaxTTLMinutes     int
	Environment       string
}

// Load populates Config from environment variables, applying sane defaults.
func Load() Config {
	cfg := Config{
		HTTPAddr:          getStrEnv("SAFEX_API_ADDR", ":8000"),
		MaxPayloadMB:      getIntEnv("SAFEX_MAX_PAYLOAD_MB", 25),
		RequestsPerMinute: getIntEnv("SAFEX_RATE_LIMIT_PER_MINUTE", 100),
		BlobBackend:       getStrEnv("SAFEX_BLOB_BACKEND", "local"),
		BlobDir:           getStrEnv("SAFEX_BLOB_DIR", "./.storage/blobs"),
		S3Bucket:          os.Getenv("SAFEX_S3_BUCKET"),
		S3Endpoint:        os.Getenv("SAFEX_S3_ENDPOINT"),
		S3AccessKey:       os.Getenv("SAFEX_S3_ACCESS_KEY"),
		S3SecretKey:       os.Getenv("SAFEX_S3_SECRET_KEY"),
		S3Region:          getStrEnv("SAFEX_S3_REGION", "us-east-1"),
		MetadataBackend:   getStrEnv("SAFEX_METADATA_BACKEND", "bbolt"),
		BoltPath:          getStrEnv("SAFEX_BBOLT_PATH", "./.storage/metadata.db"),
		RedisAddr:         os.Getenv("SAFEX_REDIS_ADDR"),
		RedisPassword:     os.Getenv("SAFEX_REDIS_PASSWORD"),
		RedisDB:           getIntEnv("SAFEX_REDIS_DB", 0),
		DefaultTTL:        time.Duration(getIntEnv("SAFEX_DEFAULT_TTL_MINUTES", 15)) * time.Minute,
		Environment:       getStrEnv("SAFEX_ENVIRONMENT", "production"),
	}
	return cfg
}

// MaxPayloadBytes returns the configured payload size limit in bytes.
func (c Config) MaxPayloadBytes() int {
	return c.MaxPayloadMB * 1024 * 1024
}

func getStrEnv(key string, defaultVal string) string {
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

// MustValidate ensures critical config pieces are present and compatible.
func (c Config) MustValidate() error {
	if c.BlobBackend != "local" && c.BlobBackend != "s3" {
		return fmt.Errorf("unsupported blob backend: %s", c.BlobBackend)
	}

	if c.MetadataBackend != "bbolt" && c.MetadataBackend != "redis" {
		return fmt.Errorf("unsupported metadata backend: %s", c.MetadataBackend)
	}

	if c.BlobBackend == "s3" {
		if c.S3Bucket == "" {
			return fmt.Errorf("SAFEX_S3_BUCKET must be set when using s3 blob backend")
		}
		if (c.S3AccessKey == "") != (c.S3SecretKey == "") {
			return fmt.Errorf("SAFEX_S3_ACCESS_KEY and SAFEX_S3_SECRET_KEY must be set together when using s3 credentials")
		}
	}

	if c.MetadataBackend == "redis" && c.RedisAddr == "" {
		return fmt.Errorf("SAFEX_REDIS_ADDR must be set when using redis metadata backend")
	}

	if c.MaxPayloadMB <= 0 {
		return fmt.Errorf("SAFEX_MAX_PAYLOAD_MB must be greater than zero")
	}

	if c.DefaultTTL <= 0 {
		return fmt.Errorf("SAFEX_DEFAULT_TTL_MINUTES must be greater than zero")
	}

	if c.RequestsPerMinute <= 0 {
		return fmt.Errorf("SAFEX_RATE_LIMIT_PER_MINUTE must be greater than zero")
	}

	return nil
}
