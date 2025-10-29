CREATE DATABASE brainz_auth;

\connect brainz_auth;

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE SCHEMA IF NOT EXISTS brainz_auth;
SET search_path TO brainz_auth;

CREATE TABLE api_key (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    expire_at     TIMESTAMPTZ NOT NULL,
    revoked_at    TIMESTAMPTZ,
    developer_id  UUID NOT NULL, -- ID из микросервиса developer portal
    key_hash      TEXT NOT NULL
);

CREATE TABLE api_keys_and_permissions (
    id                     UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    api_key_id             UUID NOT NULL REFERENCES api_key(id) ON UPDATE CASCADE ON DELETE CASCADE,
    api_key_permission_id  INTEGER NOT NULL REFERENCES api_key_permission(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT api_key_perm_unique UNIQUE (api_key_id, api_key_permission_id)
);

CREATE TABLE api_key_whitelist (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    api_key_id UUID NOT NULL REFERENCES api_key(id) ON DELETE CASCADE,
    ip_address INET NOT NULL,
    CONSTRAINT api_key_ip_unique UNIQUE (api_key_id, ip_address)
);

CREATE TABLE api_key_usage (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    api_key_id    UUID NOT NULL REFERENCES api_key(id) ON UPDATE CASCADE ON DELETE SET NULL,
    endpoint      TEXT NOT NULL,
    usage_at      TIMESTAMPTZ NOT NULL,
    response_code TEXT NOT NULL
);

CREATE TABLE api_key_institution_permission (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    api_key_id UUID NOT NULL REFERENCES api_key(id) ON DELETE CASCADE,
    institution_id INTEGER NOT NULL,
    permission_id INTEGER NOT NULL REFERENCES api_key_permission(id) ON DELETE RESTRICT,
    granted_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT api_key_institution_perm_unique UNIQUE (api_key_id, institution_id, permission_id)
);

CREATE TABLE api_key_permission (
    id     SERIAL PRIMARY KEY,
    title  TEXT NOT NULL UNIQUE
);

CREATE TABLE api_key_global_permission (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    api_key_id UUID NOT NULL REFERENCES api_key(id) ON DELETE CASCADE,
    permission_id INTEGER NOT NULL REFERENCES api_key_permission(id) ON DELETE RESTRICT,
    granted_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT api_key_global_perm_unique UNIQUE (api_key_id, permission_id)
);

CREATE INDEX idx_api_key_usage_api_key_id ON api_key_usage(api_key_id);
CREATE INDEX idx_api_key_usage_usage_at ON api_key_usage(usage_at);
CREATE INDEX idx_api_key_usage_endpoint ON api_key_usage(endpoint);
CREATE INDEX idx_api_key_whitelist_api_key_id ON api_key_whitelist(api_key_id);
CREATE INDEX idx_api_key_whitelist_ip_address ON api_key_whitelist(ip_address);