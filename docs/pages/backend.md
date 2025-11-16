# Backend API Documentation

## Architecture Overview

The backend follows Clean Architecture with these layers:

```
Presentation Layer (handlers/)
    ↓
Application Layer (usecases/)
    ↓
Domain Layer (domain/)
    ↓
Infrastructure Layer (adapters/, infrastructure/)
```

## API Endpoints

| Endpoint                 | Method | Purpose                   | Handler Function              |
| ------------------------ | ------ | ------------------------- | ----------------------------- |
| `/`                      | GET    | Home page                 | `HandleHome()`                |
| `/faq`                   | GET    | FAQ page                  | `HandleFAQ()`                 |
| `/healthz`               | GET    | Health check              | `HandleHealth()`              |
| `/secrets`               | POST   | Create secret             | `HandleCreateSecret()`        |
| `/secrets/:id`           | GET    | Load secret page          | `HandleLoadSecret()`          |
| `/secrets/reveal`        | POST   | Reveal secret             | `HandleRevealSecret()`        |
| `/opaque/register/start` | POST   | Start OPAQUE registration | `HandleOpaqueRegisterStart()` |
| `/opaque/login/start`    | POST   | Start OPAQUE login        | `HandleOpaqueLoginStart()`    |

## Background cleanup job

Backend runs a cleanup worker (`startCleanupWorker` in `app/cmd/api/main.go`) that, upon startup and then every `3h`, invokes `CleanupExpiredSecretsUseCase`. The job queries `SecretRepository` for all records whose `expires_at` has passed, and for each, deletes both metadata and the blob from storage. Errors are logged, but the remaining expired secrets continue to be deleted to prevent accumulation of "dead" records in the database.
