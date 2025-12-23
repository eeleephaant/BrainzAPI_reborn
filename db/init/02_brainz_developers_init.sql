CREATE DATABASE brainz_developers;

\connect brainz_developers;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE
    role (id SERIAL PRIMARY KEY, title TEXT NOT NULL UNIQUE);


CREATE TABLE
    developer_account (
        id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
        email TEXT NOT NULL UNIQUE,
        email_confirmed_at TIMESTAMPTZ DEFAULT NULL,
        password_hash BYTEA NOT NULL,
        salt BYTEA NOT NULL,
        two_factor_secret BYTEA DEFAULT NULL,
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
        banned_at TIMESTAMPTZ DEFAULT NULL,
        role_id INTEGER REFERENCES role (id) ON UPDATE CASCADE ON DELETE RESTRICT
    );


CREATE TABLE email_confirmation_token (
	id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
	developer_id UUID NOT NULL REFERENCES developer_account (id) ON DELETE CASCADE,
	token TEXT NOT NULL UNIQUE,
    numberic_code TEXT NOT NULL,
	expires_at TIMESTAMPTZ NOT NULL,
	used_at TIMESTAMPTZ DEFAULT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE
    developer_session (
        id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
        developer_id UUID REFERENCES developer_account (id) ON DELETE CASCADE,
        user_agent TEXT NOT NULL,
        ip_address TEXT NOT NULL,
        token_hash BYTEA NOT NULL UNIQUE,
        salt BYTEA NOT NULL,
        expires_at TIMESTAMPTZ NOT NULL,
        revoked_at TIMESTAMPTZ
    );
