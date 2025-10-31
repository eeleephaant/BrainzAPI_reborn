CREATE DATABASE brainz_lessons;

\connect brainz_lessons;

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS institutions (
	id              SERIAL PRIMARY KEY,
	created_at      TIMESTAMPTZ DEFAULT NOW(),
	updated_at      TIMESTAMPTZ DEFAULT NOW(),
	deleted_at      TIMESTAMPTZ,
	name            VARCHAR(100) NOT NULL,
	site_link       VARCHAR(255)
);

CREATE INDEX IF NOT EXISTS idx_institution_name ON institutions(name);

CREATE TABLE IF NOT EXISTS groups (
	id               SERIAL PRIMARY KEY,
	created_at       TIMESTAMPTZ DEFAULT NOW(),
	updated_at       TIMESTAMPTZ DEFAULT NOW(),
	deleted_at       TIMESTAMPTZ,
	name             VARCHAR(50) NOT NULL,
	institution_id   INTEGER REFERENCES institutions(id)
		ON UPDATE CASCADE
		ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_group_institution_id ON groups(institution_id);

CREATE TABLE IF NOT EXISTS lessons (
	id               SERIAL PRIMARY KEY,
	created_at       TIMESTAMPTZ DEFAULT NOW(),
	updated_at       TIMESTAMPTZ DEFAULT NOW(),
	deleted_at       TIMESTAMPTZ,
	name             VARCHAR(100) NOT NULL,
	cab_num          VARCHAR(20),
	teacher_name     VARCHAR(100) NOT NULL,
	start_time       TIMESTAMPTZ NOT NULL,
	end_time         TIMESTAMPTZ NOT NULL,
	num              SMALLINT NOT NULL,
	group_id         INTEGER REFERENCES groups(id)
		ON UPDATE CASCADE
		ON DELETE SET NULL,
	institution_id   INTEGER REFERENCES institutions(id)
		ON UPDATE CASCADE
		ON DELETE SET NULL,
	CONSTRAINT unique_lesson UNIQUE (name, teacher_name, start_time, num, group_id)
);

CREATE INDEX IF NOT EXISTS idx_lesson_start_time ON lessons(start_time);
CREATE INDEX IF NOT EXISTS idx_lesson_end_time ON lessons(end_time);
CREATE INDEX IF NOT EXISTS idx_lesson_group_id ON lessons(group_id);
CREATE INDEX IF NOT EXISTS idx_lesson_institution_id ON lessons(institution_id);

RESET search_path;
