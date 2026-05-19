# Secret Server

A REST API for storing and sharing secrets via one-time URLs. Secrets expire after a configurable number of views or a TTL, and are stored encrypted at rest.

## Getting Started

### Local (requires MongoDB)

```bash
go build -o secretserver
ENCRYPTION_KEY=your-key ./secretserver
```

Server listens on `:8080`. MongoDB defaults to `localhost:27017`.

### Docker

```bash
docker-compose up --build
```

Starts the app and a MongoDB instance together.

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `ENCRYPTION_KEY` | `default-key-change-in-production!!` | AES-256-GCM encryption key |
| `MONGO_URI` | `mongodb://localhost:27017` | MongoDB connection URI |

## API

Full spec: [`swagger.yaml`](swagger.yaml)

### Store a secret

```
POST /secret
Content-Type: application/x-www-form-urlencoded

secret=<text>&expireAfterViews=<int>&expireAfter=<minutes>
```

- `expireAfterViews` — must be greater than 0; secret is deleted after this many reads
- `expireAfter` — minutes until expiry; `0` means never expires

Returns the created secret with its `hash`.

### Retrieve a secret

```
GET /secret/:hash
```

Returns the secret as JSON and decrements the remaining view count. The secret is permanently deleted when the last view is consumed or the TTL has passed. Returns `404` if not found or expired.

### Metrics

```
GET /metrics
```

Returns per-endpoint request counts and average latency (ms):

```json
{
  "GET /secret/:hash": { "requests": 10, "avg_latency_ms": 2.4 },
  "POST /secret":      { "requests": 5,  "avg_latency_ms": 8.1 }
}
```

## Running Tests

```bash
go test ./...
```

Tests cover AES-256-GCM encryption/decryption (`internal/crypto`), request metrics tracking (`internal/metrics`), and hash generation (`secret`).

## Architecture

```
main.go
├── external/mongodb/    — MongoDB singleton; URI from MONGO_URI env var
├── internal/crypto/     — AES-256-GCM encrypt/decrypt; key from ENCRYPTION_KEY
├── internal/metrics/    — in-memory request counter + latency tracker
├── server/              — Echo HTTP server, middleware wiring, route registration
└── secret/              — handlers, model (Secret + DoHash), repository
```

Secrets are encrypted before insertion and decrypted on retrieval. Each secret carries a cryptographically random 128-bit hex hash as its identifier.
