-- Initial database schema for config-service

-- Services table
CREATE TABLE IF NOT EXISTS services (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Environments table
CREATE TABLE IF NOT EXISTS environments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    service_id UUID NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(service_id, name)
);

-- Config versions table (history) - created first for FK reference
CREATE TABLE IF NOT EXISTS config_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    config_key_id UUID NOT NULL,  -- FK added after config_keys exists
    value TEXT NOT NULL,
    version INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by VARCHAR(255)
);

-- Config keys table
CREATE TABLE IF NOT EXISTS config_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    environment_id UUID NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
    key VARCHAR(255) NOT NULL,
    value TEXT,  -- Denormalized: cached active value for fast reads
    value_type VARCHAR(50) DEFAULT 'string',
    active_version_id UUID REFERENCES config_versions(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(environment_id, key)
);

-- Add FK from config_versions to config_keys (circular reference)
ALTER TABLE config_versions 
    ADD CONSTRAINT fk_config_versions_config_key 
    FOREIGN KEY (config_key_id) REFERENCES config_keys(id) ON DELETE CASCADE;

-- Unique constraint on version per key
ALTER TABLE config_versions 
    ADD CONSTRAINT uq_config_versions_key_version 
    UNIQUE(config_key_id, version);

-- Audit logs table
CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    service_id UUID NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    config_key_id UUID REFERENCES config_keys(id) ON DELETE SET NULL,
    action VARCHAR(50) NOT NULL,
    old_value TEXT,
    new_value TEXT,
    performed_by VARCHAR(255),
    performed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for better query performance
CREATE INDEX idx_environments_service_id ON environments(service_id);
CREATE INDEX idx_config_keys_environment_id ON config_keys(environment_id);
CREATE INDEX idx_config_keys_active_version ON config_keys(active_version_id);
CREATE INDEX idx_config_versions_config_key_id ON config_versions(config_key_id);
CREATE INDEX idx_config_versions_version ON config_versions(config_key_id, version);
CREATE INDEX idx_audit_logs_service_id ON audit_logs(service_id);
CREATE INDEX idx_audit_logs_config_key_id ON audit_logs(config_key_id);
CREATE INDEX idx_audit_logs_performed_at ON audit_logs(performed_at);
