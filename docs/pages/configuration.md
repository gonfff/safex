# Configuration Reference

This page documents all environment variables available for configuring Safex.

## Environment Variables

### Server Configuration

| Variable                      | Description                           | Default      | Example                   |
| ----------------------------- | ------------------------------------- | ------------ | ------------------------- |
| `SAFEX_API_ADDR`              | HTTP server bind address              | `:8000`      | `:8080`, `localhost:8000` |
| `SAFEX_ENVIRONMENT`           | Runtime environment                   | `production` | `development`, `testing`  |
| `SAFEX_MAX_PAYLOAD_MB`        | Maximum payload size in MB            | `10`         | `5`, `50`                 |
| `SAFEX_RATE_LIMIT_PER_MINUTE` | Rate limit POST queries per client IP | `10`         | `20`, `100`               |

### Secret Management

| Variable                    | Description                        | Default      | Example            |
| --------------------------- | ---------------------------------- | ------------ | ------------------ |
| `SAFEX_DEFAULT_TTL_MINUTES` | Default secret lifetime in minutes | `15`         | `60`, `1440` (24h) |
| `SAFEX_MAX_TTL_MINUTES`     | Maximum allowed TTL                | `1440` (24h) | `1440` (24h)       |

### Blob Storage Configuration

| Variable             | Description             | Default            | Example                       |
| -------------------- | ----------------------- | ------------------ | ----------------------------- |
| `SAFEX_BLOB_BACKEND` | Storage backend type    | `local`            | `s3` (Not tested with S3 yet) |
| `SAFEX_BLOB_DIR`     | Local storage directory | `./.storage/blobs` | `/var/lib/safex/blobs`        |

#### S3-Compatible Storage (Not tested with S3 yet)

| Variable              | Description     | Default     | Example                    |
| --------------------- | --------------- | ----------- | -------------------------- |
| `SAFEX_S3_BUCKET`     | S3 bucket name  | -           | `safex-secrets`            |
| `SAFEX_S3_ENDPOINT`   | S3 endpoint URL | -           | `https://s3.amazonaws.com` |
| `SAFEX_S3_ACCESS_KEY` | S3 access key   | -           | `AKIA...`                  |
| `SAFEX_S3_SECRET_KEY` | S3 secret key   | -           | `wJalrXUt...`              |
| `SAFEX_S3_REGION`     | S3 region       | `us-east-1` | `eu-west-1`                |

### Metadata Storage Configuration

| Variable                 | Description           | Default                  | Example                             |
| ------------------------ | --------------------- | ------------------------ | ----------------------------------- |
| `SAFEX_METADATA_BACKEND` | Metadata backend type | `bbolt`                  | `redis` (Not tested with Redis yet) |
| `SAFEX_BBOLT_PATH`       | BoltDB file path      | `./.storage/metadata.db` | `/var/lib/safex/metadata.db`        |

#### Redis Configuration (Not tested with Redis yet)

| Variable               | Description           | Default | Example          |
| ---------------------- | --------------------- | ------- | ---------------- |
| `SAFEX_REDIS_ADDR`     | Redis server address  | -       | `localhost:6379` |
| `SAFEX_REDIS_PASSWORD` | Redis password        | -       | `mypassword`     |
| `SAFEX_REDIS_DB`       | Redis database number | `0`     | `1`              |

### OPAQUE Authentication

| Variable                           | Description                 | Default     | Example              |
| ---------------------------------- | --------------------------- | ----------- | -------------------- |
| `SAFEX_OPAQUE_SERVER_ID`           | OPAQUE server identifier    | `safex`     | `my-safex-instance`  |
| `SAFEX_OPAQUE_PRIVATE_KEY`         | Server private key (base64) | _generated_ | `dGhpc2lzYXRlc3Q...` |
| `SAFEX_OPAQUE_OPRF_SEED`           | OPRF seed (base64)          | _generated_ | `dGhpc2lzYXRlc3Q...` |
| `SAFEX_OPAQUE_SESSION_TTL_SECONDS` | OPAQUE session timeout      | `120`       | `300`                |

### Key Generation

If not provided, the OPAQUE keys are automatically generated at startup:

```go
// Generated private key (32 random bytes, base64 encoded)
func getOpaquePrivateKey() string {
    key := make([]byte, 32)
    rand.Read(key)
    return base64.StdEncoding.EncodeToString(key)
}

// Generated OPRF seed (64 random bytes, base64 encoded)
func getOpaqueOPRFSeed() string {
    seed := make([]byte, 64)
    rand.Read(seed)
    return base64.StdEncoding.EncodeToString(seed)
}
```

**Important**: In production, you should set these explicitly to ensure consistency across restarts and multiple instances.
