# Low-Level Design (LLD) - Config Manager Service

## 1. System Overview

A centralized configuration management service that enables multiple applications to store, retrieve, version, and rollback configuration values across different environments (dev, staging, production).

---

## 2. Domain Model

### 2.1 Entity Relationship Diagram

```
┌─────────────────┐
│    Service      │
│─────────────────│
│ id: UUID        │
│ name: string    │
│ created_at      │
│ updated_at      │
└────────┬────────┘
         │
         │ 1:N
         │
┌────────▼────────┐
│  Environment    │
│─────────────────│
│ id: UUID        │
│ service_id: FK  │
│ name: string    │
│ created_at      │
│ updated_at      │
└────────┬────────┘
         │
         │ 1:N
         │
┌────────▼────────────────┐         ┌──────────────────┐
│    ConfigKey            │◄────1:1─│  ConfigVersion   │
│─────────────────────────│         │──────────────────│
│ id: UUID                │         │ id: UUID         │
│ environment_id: FK      │         │ config_key_id:FK │
│ key: string             │         │ value: text      │
│ value: text (cached)    │         │ version: int     │
│ value_type: string      │         │ created_at       │
│ active_version_id: FK ──┼────────►│ created_by       │
│ created_at              │         └──────────────────┘
│ updated_at              │                 ▲
└─────────────────────────┘                 │
         │                                  │ 1:N
         │                                  │
         └──────────────────────────────────┘

┌──────────────────┐
│   AuditLog       │
│──────────────────│
│ id: UUID         │
│ service_id: FK   │
│ config_key_id:FK │
│ action: string   │
│ old_value: text  │
│ new_value: text  │
│ performed_by     │
│ performed_at     │
└──────────────────┘
```

### 2.2 Entity Descriptions

#### Service
- **Purpose**: Represents a registered application/service that uses config manager
- **Relationships**: 
  - Has many Environments (1:N)
  - Has many AuditLogs (1:N)
- **Constraints**: 
  - `name` must be unique
  - Cannot be deleted if environments exist (CASCADE delete)

#### Environment
- **Purpose**: Deployment environment (dev, staging, prod) scoped to a service
- **Relationships**:
  - Belongs to Service (N:1)
  - Has many ConfigKeys (1:N)
- **Constraints**:
  - `(service_id, name)` must be unique
  - Cannot be deleted if config keys exist (CASCADE delete)

#### ConfigKey
- **Purpose**: Configuration key with cached current value
- **Key Fields**:
  - `value`: Denormalized cache of active version's value (read optimization)
  - `value_type`: Data type hint (string, int, bool, json)
  - `active_version_id`: Points to current version (NULL if no versions)
- **Relationships**:
  - Belongs to Environment (N:1)
  - Has one active ConfigVersion (N:1 via `active_version_id`)
  - Has many ConfigVersions (1:N history)
- **Constraints**:
  - `(environment_id, key)` must be unique
  - Circular FK relationship with ConfigVersion (handled via deferred constraints)

#### ConfigVersion
- **Purpose**: Immutable version history of config values
- **Key Fields**:
  - `version`: Monotonically increasing per config_key_id
  - `value`: The actual config value for this version
  - `created_by`: User/system that created this version
- **Relationships**:
  - Belongs to ConfigKey (N:1)
- **Constraints**:
  - `(config_key_id, version)` must be unique
  - Versions are immutable (INSERT only, no UPDATE)

#### AuditLog
- **Purpose**: Audit trail for compliance and debugging
- **Key Fields**:
  - `action`: create, update, delete, rollback
  - `old_value`, `new_value`: What changed
  - `performed_by`: Extracted from auth context
- **Relationships**:
  - References Service (N:1)
  - References ConfigKey (N:1, nullable)
- **Constraints**:
  - Append-only (no updates/deletes)
  - Indexed on `performed_at` for time-range queries

---

## 3. API Design

### 3.1 REST Endpoints

#### Service Management
```
POST   /api/v1/services
GET    /api/v1/services
GET    /api/v1/services/:service_id
DELETE /api/v1/services/:service_id

POST   /api/v1/services/:service_id/environments
GET    /api/v1/services/:service_id/environments
```

#### Config Operations
```
GET    /api/v1/configs/:env_id                    # List all configs
GET    /api/v1/configs/:env_id/:key               # Get config by key
POST   /api/v1/configs/:env_id/:key               # Create/update config
DELETE /api/v1/configs/:env_id/:key               # Delete config

GET    /api/v1/configs/:env_id/:key/versions      # Version history
POST   /api/v1/configs/:env_id/:key/rollback      # Rollback to version

GET    /api/v1/audit/:service_id                  # Audit logs
```

### 3.2 Request/Response DTOs

#### Set Config Request
```json
{
  "value": "postgresql://db:5432",
  "value_type": "string",
  "created_by": "user@example.com"
}
```

#### Config Response
```json
{
  "key": "database.url",
  "value": "postgresql://db:5432",
  "value_type": "string",
  "version": 3,
  "updated_at": "2026-02-02T10:30:00Z"
}
```

#### Rollback Request
```json
{
  "version": 2,
  "performed_by": "admin@example.com"
}
```

#### Version History Response
```json
{
  "versions": [
    {
      "version": 3,
      "value": "postgresql://db:5432",
      "created_at": "2026-02-02T10:30:00Z",
      "created_by": "user@example.com"
    },
    {
      "version": 2,
      "value": "postgresql://old:5432",
      "created_at": "2026-02-01T09:00:00Z",
      "created_by": "user@example.com"
    }
  ],
  "total": 2
}
```

---

## 4. Component Architecture

### 4.1 Layer Responsibilities

```
┌──────────────────────────────────────────────┐
│              HTTP Layer                       │
│  (Handler, Middleware, Routes)                │
│  - Request validation                         │
│  - Response formatting                        │
│  - Auth/CORS/Logging middleware               │
└──────────────────┬───────────────────────────┘
                   │
┌──────────────────▼───────────────────────────┐
│              DTO Layer                        │
│  - Request/Response transformation            │
│  - Domain <-> API contract mapping            │
└──────────────────┬───────────────────────────┘
                   │
┌──────────────────▼───────────────────────────┐
│            Usecase Layer                      │
│  - Business logic orchestration               │
│  - Transaction management                     │
│  - Version increment logic                    │
│  - Audit log creation                         │
└──────────────────┬───────────────────────────┘
                   │
┌──────────────────▼───────────────────────────┐
│          Repository Layer                     │
│  - DB queries (CRUD)                          │
│  - SQL composition                            │
│  - Connection pooling                         │
└──────────────────┬───────────────────────────┘
                   │
┌──────────────────▼───────────────────────────┐
│            Database (PostgreSQL)              │
└───────────────────────────────────────────────┘
```

### 4.2 Key Components

#### ConfigUsecase
```go
type ConfigUsecase struct {
    configRepo repository.ConfigRepository
    auditRepo  repository.AuditRepository
}

// GetConfig - Fast read from config_keys (no JOIN)
func (u *ConfigUsecase) GetConfig(ctx, envID, key string) (*ConfigKey, error)

// SetConfig - Atomic write with versioning
// 1. Get current config_key
// 2. Calculate next version = max(version) + 1
// 3. INSERT new config_version
// 4. UPDATE config_keys SET active_version_id, value, updated_at
// 5. INSERT audit_log
// 6. Commit transaction
func (u *ConfigUsecase) SetConfig(ctx, envID, key, value, performedBy string) error

// DeleteConfig - Soft or hard delete
func (u *ConfigUsecase) DeleteConfig(ctx, keyID, performedBy string) error

// ListConfigs - Bulk fetch for environment
func (u *ConfigUsecase) ListConfigs(ctx, envID string) ([]*ConfigKey, error)
```

#### RollbackUsecase
```go
type RollbackUsecase struct {
    configRepo repository.ConfigRepository
    auditRepo  repository.AuditRepository
}

// RollbackToVersion - Revert to older version
// 1. Fetch target version from config_versions
// 2. UPDATE config_keys SET active_version_id = target, value = target.value
// 3. INSERT audit_log with action='rollback'
// 4. Commit transaction
func (u *RollbackUsecase) RollbackToVersion(ctx, keyID string, version int, performedBy string) error

// GetVersionHistory - List all versions for a key
func (u *RollbackUsecase) GetVersionHistory(ctx, keyID string) ([]*ConfigVersion, error)
```

---

## 5. Data Access Patterns

### 5.1 Read Patterns

#### Get Single Config (Hot Path - Optimized)
```sql
-- No JOIN needed due to denormalized value
SELECT id, key, value, value_type, updated_at
FROM config_keys
WHERE environment_id = $1 AND key = $2;
```

#### Get Config with Version Info
```sql
SELECT 
    ck.id, ck.key, ck.value, ck.value_type,
    cv.version, cv.created_by, cv.created_at as version_created_at
FROM config_keys ck
LEFT JOIN config_versions cv ON cv.id = ck.active_version_id
WHERE ck.environment_id = $1 AND ck.key = $2;
```

#### List All Configs for Environment
```sql
SELECT id, key, value, value_type, updated_at
FROM config_keys
WHERE environment_id = $1
ORDER BY key;
```

### 5.2 Write Patterns

#### Create New Config
```sql
BEGIN;

-- Insert config key
INSERT INTO config_keys (environment_id, key, value_type)
VALUES ($1, $2, $3)
RETURNING id;

-- Insert first version
INSERT INTO config_versions (config_key_id, value, version, created_by)
VALUES ($config_key_id, $4, 1, $5)
RETURNING id;

-- Update config_key with active version
UPDATE config_keys
SET active_version_id = $version_id, value = $4, updated_at = NOW()
WHERE id = $config_key_id;

-- Insert audit log
INSERT INTO audit_logs (service_id, config_key_id, action, new_value, performed_by)
VALUES ($6, $config_key_id, 'create', $4, $5);

COMMIT;
```

#### Update Existing Config
```sql
BEGIN;

-- Get next version number
SELECT COALESCE(MAX(version), 0) + 1 as next_version
FROM config_versions
WHERE config_key_id = $1;

-- Insert new version
INSERT INTO config_versions (config_key_id, value, version, created_by)
VALUES ($1, $2, $next_version, $3)
RETURNING id;

-- Get old value for audit
SELECT value FROM config_keys WHERE id = $1;

-- Update config_key
UPDATE config_keys
SET active_version_id = $new_version_id, value = $2, updated_at = NOW()
WHERE id = $1;

-- Insert audit log
INSERT INTO audit_logs (service_id, config_key_id, action, old_value, new_value, performed_by)
VALUES ($4, $1, 'update', $old_value, $2, $3);

COMMIT;
```

#### Rollback to Previous Version
```sql
BEGIN;

-- Get target version
SELECT id, value FROM config_versions
WHERE config_key_id = $1 AND version = $2;

-- Get current value for audit
SELECT value FROM config_keys WHERE id = $1;

-- Update config_key to point to old version
UPDATE config_keys
SET active_version_id = $target_version_id, value = $target_value, updated_at = NOW()
WHERE id = $1;

-- Insert audit log
INSERT INTO audit_logs (service_id, config_key_id, action, old_value, new_value, performed_by)
VALUES ($3, $1, 'rollback', $current_value, $target_value, $4);

COMMIT;
```

---

## 6. SDK Client Design

### 6.1 Client Components

```
┌─────────────────────────────────────┐
│          Client                     │
│  - HTTP communication               │
│  - API key authentication           │
│  - Get/Set operations               │
├─────────────────────────────────────┤
│          Store                      │
│  - In-memory cache (map)            │
│  - Thread-safe (RWMutex)            │
│  - Get/Set/Delete/Clear             │
├─────────────────────────────────────┤
│          Poller                     │
│  - Background goroutine             │
│  - Periodic fetch (configurable)    │
│  - Bulk update local store          │
└─────────────────────────────────────┘
```

### 6.2 Client Usage Flow

```go
// Initialize
client := sdk.NewClient(sdk.Config{
    BaseURL:      "http://config-service:8080",
    APIKey:       "service-api-key",
    Environment:  "production",
    PollInterval: 30 * time.Second,
})
defer client.Close()

// Start background polling
ctx := context.Background()
client.StartPolling(ctx)

// Read flow:
// 1. Check local store (in-memory)
// 2. If miss, fetch from server
// 3. Cache in local store
// 4. Return value

value, err := client.Get(ctx, "production", "database.max_connections")

// Fallback pattern
timeout := client.GetWithDefault(ctx, "prod", "request.timeout", "30s")
```

### 6.3 Polling Strategy

```
Time: 0s     30s    60s    90s
      │      │      │      │
      ▼      ▼      ▼      ▼
    [Fetch][Fetch][Fetch][Fetch]
      │      │      │      │
      └──────┴──────┴──────┴──► Update Store

On each poll:
1. GET /api/v1/configs/:env (bulk fetch all)
2. Replace entire local store with fresh data
3. Log errors but continue polling
4. Client reads always go to local store first
```

---

## 7. Security & Auth

### 7.1 API Key Authentication
- Each service gets unique API key
- API key stored in `X-API-Key` header
- Middleware validates key before handler execution
- Rate limiting per API key

### 7.2 Authorization
- Service can only access its own environments
- Environment ID → Service ID lookup via JOIN
- Audit log tracks `performed_by` from auth context

---

## 8. Performance Considerations

### 8.1 Read Optimization
- **Denormalized value**: No JOIN for 99% of reads
- **Indexes**: B-tree on `(environment_id, key)`
- **SDK caching**: Local in-memory store reduces server load
- **Connection pooling**: Reuse DB connections

### 8.2 Write Optimization
- **Batch inserts**: Version + audit log in single transaction
- **Asynchronous audit**: Consider async queue for audit logs
- **Version pruning**: Optional: delete versions older than N days

### 8.3 Scalability
- **Read replicas**: Route SDK reads to replicas
- **Write master**: Single writer for consistency
- **Sharding**: Partition by `service_id` if needed
- **Caching layer**: Redis for hot configs

---

## 9. Error Handling

### 9.1 Client Errors (4xx)
- `400 Bad Request`: Invalid value_type, missing required fields
- `401 Unauthorized`: Invalid/missing API key
- `403 Forbidden`: Service accessing wrong environment
- `404 Not Found`: Config key doesn't exist
- `409 Conflict`: Version conflict during rollback

### 9.2 Server Errors (5xx)
- `500 Internal Server Error`: DB connection failure, transaction rollback
- `503 Service Unavailable`: DB healthcheck failing

### 9.3 SDK Error Handling
```go
value, err := client.Get(ctx, env, key)
if err != nil {
    // Log error
    log.Printf("config fetch failed: %v", err)
    
    // Fallback to default
    value = defaultValue
}
```

---

## 10. Monitoring & Observability

### 10.1 Metrics
- Config read/write latency (p50, p95, p99)
- SDK cache hit rate
- Failed authentication attempts
- Version rollback frequency

### 10.2 Logging
- Structured logs (JSON)
- Request ID tracing
- Audit log queries for compliance

### 10.3 Health Checks
- `/health`: DB connection status
- `/metrics`: Prometheus metrics endpoint

---

## 11. Database Indexes Summary

```sql
-- Primary lookups
CREATE INDEX idx_config_keys_environment_id ON config_keys(environment_id);
CREATE INDEX idx_config_keys_env_key ON config_keys(environment_id, key);

-- Version queries
CREATE INDEX idx_config_versions_config_key_id ON config_versions(config_key_id);
CREATE INDEX idx_config_versions_key_version ON config_versions(config_key_id, version);

-- Audit queries
CREATE INDEX idx_audit_logs_service_id ON audit_logs(service_id);
CREATE INDEX idx_audit_logs_performed_at ON audit_logs(performed_at DESC);
```

---

## 12. Future Enhancements

1. **Config inheritance**: Environment-specific overrides with fallback to defaults
2. **Secret management**: Encrypt sensitive values at rest
3. **Change approval workflow**: Require approval for prod changes
4. **Webhooks**: Notify services on config changes
5. **Feature flags**: Toggle features without code deployment
6. **Schema validation**: JSON schema validation for structured configs
7. **Multi-region**: Replicate configs across regions
