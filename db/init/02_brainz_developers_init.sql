CREATE DATABASE brainz_developers;

-- \connect brainz_developers;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE SCHEMA IF NOT EXISTS brainz_developers;

SET
    search_path TO brainz_developers;

CREATE TABLE
    role (id SERIAL PRIMARY KEY, title TEXT NOT NULL UNIQUE);

-- Developers
CREATE TABLE
    developer (
        id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
        email TEXT NOT NULL UNIQUE,
        password_hash TEXT NOT NULL,
        salt TEXT NOT NULL,
        two_factor_secret TEXT,
        two_factor_enabled BOOL NOT NULL DEFAULT FALSE,
        role_id INTEGER REFERENCES role (id) ON UPDATE CASCADE ON DELETE RESTRICT
    );

CREATE TABLE
    developer_session (
        id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
        developer_id UUID REFERENCES developer (id) ON DELETE CASCADE,
        user_agent TEXT NOT NULL,
        ip_address INET NOT NULL,
        token_hash TEXT NOT NULL UNIQUE,
        salt TEXT NOT NULL,
        expires_at TIMESTAMPTZ NOT NULL,
        revoked_at TIMESTAMPTZ
    );