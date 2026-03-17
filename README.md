# Config Manager

A centralized configuration management service with a Go SDK for managing application configs across multiple services and environments with full version control and rollback capabilities.

## Features

- ✅ **Multi-service & Multi-environment** - Manage configs for multiple services across dev, staging, prod
- ✅ **Type-safe values** - Native support for string, int, float, bool, and JSON types
- ✅ **Version control** - Full history of all config changes with rollback support
- ✅ **Audit logging** - Track who changed what and when
- ✅ **Go SDK** - Client library with local caching and background polling
- ✅ **RESTful API** - Clean HTTP API for all operations
- ✅ **PostgreSQL backed** - Reliable storage with ACID transactions

## Quick Start

### 1. Start the Server

```bash
# Start with Docker Compose
docker-compose up -d

# Or run locally
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_NAME=config_service
export SERVER_PORT=8080

go run cmd/server/main.go
```

### 2. Use the Go SDK

```bash
go get github.com/sachin/config-manager/pkg/sdk
```

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/sachin/config-manager/pkg/sdk"
)

func main() {
    // Initialize client
    client := sdk.NewClient(sdk.Config{
        BaseURL:       "http://localhost:8080",
        APIKey:        "your-api-key",
        EnvironmentID: "prod-env-id",
        PollInterval:  30 * time.Second, // Auto-refresh every 30s
    })
    defer client.Close()

    ctx := context.Background()
    
    // Start background polling
    client.StartPolling(ctx)

    // Get typed values
    dbHost, _ := client.GetString(ctx, "database.host")
    maxConn, _ := client.GetInt(ctx, "database.max_connections")
    debug, _ := client.GetBool(ctx, "app.debug_mode")

    log.Printf("DB: %s, MaxConn: %d, Debug: %v", dbHost, maxConn, debug)

    // Get with defaults
    timeout := client.GetIntWithDefault(ctx, "request.timeout", 30)

    // Set config
    client.Set(ctx, "app.version", "v2.0.0", "admin")
}
```

## SDK API Reference

### Client Initialization

```go
client := sdk.NewClient(sdk.Config{
    BaseURL:       "http://localhost:8080",       // Config service URL
    APIKey:        "your-api-key",                // Authentication key
    EnvironmentID: "environment-uuid",            // Target environment
    PollInterval:  30 * time.Second,              // Auto-refresh interval (0 = disabled)
    HTTPTimeout:   10 * time.Second,              // HTTP request timeout
})
```

### Reading Configs

```go
// Type-safe getters
value, err := client.Get(ctx, "key")              // Returns interface{}
str, err := client.GetString(ctx, "key")          // Returns string
num, err := client.GetInt(ctx, "key")             // Returns int
flag, err := client.GetBool(ctx, "key")           // Returns bool
flt, err := client.GetFloat(ctx, "key")           // Returns float64
json, err := client.GetJSON(ctx, "key")           // Returns interface{} (map/slice)

// With defaults (no error returned)
str := client.GetStringWithDefault(ctx, "key", "default")
num := client.GetIntWithDefault(ctx, "key", 100)
flag := client.GetBoolWithDefault(ctx, "key", false)

// List all configs
configs, err := client.ListAll(ctx)
for key, config := range configs {
    fmt.Printf("%s = %v (type: %s)\n", key, config.Value, config.ValueType)
}
```

### Writing Configs

```go
// Set any type - automatically infers type
err := client.Set(ctx, "database.host", "postgres.example.com", "admin")
err := client.Set(ctx, "database.port", 5432, "admin")
err := client.Set(ctx, "app.debug", true, "admin")

// Set JSON
config := map[string]interface{}{
    "host": "localhost",
    "port": 5432,
}
err := client.Set(ctx, "database.config", config, "admin")

// Delete config
err := client.Delete(ctx, "old.config.key", "admin")
```

### Background Polling

```go
// Start automatic background refresh
client.StartPolling(ctx)

// Stop polling
client.StopPolling()

// Close and cleanup
client.Close()
```

## REST API Endpoints

### Services
- `POST /api/v1/services` - Create service
- `GET /api/v1/services` - List services
- `GET /api/v1/services/{id}` - Get service
- `PUT /api/v1/services/{id}` - Update service
- `DELETE /api/v1/services/{id}` - Delete service

### Environments
- `POST /api/v1/services/{serviceId}/environments` - Create environment
- `GET /api/v1/services/{serviceId}/environments` - List environments
- `GET /api/v1/environments/{id}` - Get environment
- `PUT /api/v1/environments/{id}` - Update environment
- `DELETE /api/v1/environments/{id}` - Delete environment

### Configs
- `GET /api/v1/configs/{envId}` - List all configs
- `GET /api/v1/configs/{envId}/{key}` - Get config
- `POST /api/v1/configs/{envId}/{key}` - Create/update config
- `DELETE /api/v1/configs/{envId}/{key}` - Delete config
- `GET /api/v1/configs/{envId}/{key}/versions` - Version history
- `POST /api/v1/configs/{envId}/{key}/rollback` - Rollback to version

### Audit
- `GET /api/v1/audit/{id}` - Get audit log
- `GET /api/v1/services/{serviceId}/audit` - List audit logs

## API Examples

### Create Service
```bash
curl -X POST http://localhost:8080/api/v1/services \
  -H "Content-Type: application/json" \
  -d '{"name": "payment-service"}'
```

### Create Environment
```bash
curl -X POST http://localhost:8080/api/v1/services/{serviceId}/environments \
  -H "Content-Type: application/json" \
  -d '{"name": "production"}'
```

### Set Config
```bash
curl -X POST http://localhost:8080/api/v1/configs/{envId}/database.host \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key" \
  -d '{"value": "postgres.example.com", "performed_by": "admin"}'
```

### Get Config
```bash
curl http://localhost:8080/api/v1/configs/{envId}/database.host \
  -H "X-API-Key: your-api-key"
```

## Architecture

```
┌─────────────────────────────────────────────┐
│           HTTP Layer (Controller)           │
│  - Request validation                       │
│  - Response formatting                      │
│  - Middleware (Auth, CORS, Logging)         │
└──────────────────┬──────────────────────────┘
                   │
┌──────────────────▼──────────────────────────┐
│              Service Layer                   │
│  - Business logic                            │
│  - Transaction management                    │
│  - Versioning & audit logging                │
└──────────────────┬──────────────────────────┘
                   │
┌──────────────────▼──────────────────────────┐
│           Repository Layer                   │
│  - Database queries                          │
│  - Type conversion (DB ↔ Go)                 │
└──────────────────┬──────────────────────────┘
                   │
┌──────────────────▼──────────────────────────┐
│          PostgreSQL Database                 │
│  - ACID transactions                         │
│  - Indexes for performance                   │
└──────────────────────────────────────────────┘
```

## Database Schema

- **services** - Registered applications
- **environments** - Deployment environments (dev, staging, prod)
- **config_keys** - Current config values (denormalized for fast reads)
- **config_versions** - Immutable version history
- **audit_logs** - Change audit trail

## Performance Features

- **Denormalized reads** - Config values cached in `config_keys` table (no JOIN needed)
- **SDK local cache** - In-memory store reduces server requests
- **Background polling** - Automatic config refresh without blocking
- **Indexes** - Optimized for common query patterns
- **Connection pooling** - Reusable database connections

## Development

```bash
# Run tests
go test ./...

# Build
go build ./...

# Run server
go run cmd/server/main.go

# Build Docker image
docker build -t config-manager .
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER_HOST` | `0.0.0.0` | Server bind address |
| `SERVER_PORT` | `8080` | Server port |
| `DB_HOST` | `localhost` | PostgreSQL host |
| `DB_PORT` | `5432` | PostgreSQL port |
| `DB_USER` | `postgres` | Database user |
| `DB_PASSWORD` | - | Database password |
| `DB_NAME` | `config_service` | Database name |
| `DB_SSLMODE` | `disable` | SSL mode |

## License

MIT
