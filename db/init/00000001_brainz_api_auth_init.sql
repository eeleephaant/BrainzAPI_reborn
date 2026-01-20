CREATE DATABASE brainz_auth;

\connect brainz_auth;

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE api_key (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name          TEXT NOT NULL DEFAULT 'Untitled key',
    expire_at     TIMESTAMPTZ NOT NULL,
    revoked_at    TIMESTAMPTZ,
    developer_id  UUID NOT NULL, -- ID из микросервиса developer portal
    key_hash      BYTEA NOT NULL,
    salt          BYTEA NOT NULL,
    prefix_raw    varchar(10) NOT NULL, -- bcs format is brainz_xxxxxxxxxxxxxxx
    suffix_raw    varchar(4) NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE api_key_permission (
    id     SERIAL PRIMARY KEY,
    title  TEXT NOT NULL UNIQUE
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
    method        TEXT NOT NULL,
    usage_at      TIMESTAMPTZ NOT NULL,
    response_code TEXT NOT NULL
);

CREATE TABLE api_key_permission_grant (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    api_key_id      UUID NOT NULL REFERENCES api_key(id) ON DELETE CASCADE,
    permission_id   INTEGER NOT NULL REFERENCES api_key_permission(id) ON DELETE RESTRICT,
    institution_id  INTEGER NULL,  -- NULL = глобальная, NOT NULL = для конкретной организации
    granted_at      TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT api_key_grant_unique
        UNIQUE (api_key_id, permission_id, institution_id)
);

CREATE INDEX idx_api_key_usage_api_key_id ON api_key_usage(api_key_id);
CREATE INDEX idx_api_key_usage_usage_at ON api_key_usage(usage_at);
CREATE INDEX idx_api_key_usage_endpoint ON api_key_usage(endpoint);
CREATE INDEX idx_api_key_whitelist_api_key_id ON api_key_whitelist(api_key_id);
CREATE INDEX idx_api_key_whitelist_ip_address ON api_key_whitelist(ip_address);
