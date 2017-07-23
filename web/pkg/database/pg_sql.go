package database

const installScript = `
BEGIN;

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'quality') THEN
    CREATE TYPE quality AS ENUM ('high', 'low');
  END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'format') THEN
    CREATE TYPE format AS ENUM ('audio', 'video');
  END IF;
END
$$;

CREATE TABLE IF NOT EXISTS feeds (
	id BIGSERIAL PRIMARY KEY,
	hash_id VARCHAR(12) NOT NULL CHECK (hash_id <> ''),
	user_id VARCHAR(32) NULL,
	url VARCHAR(64) NOT NULL CHECK (url <> ''),
	page_size INT NOT NULL DEFAULT 50,
	quality quality NOT NULL DEFAULT 'high',
	format format NOT NULL DEFAULT 'video'
);

COMMIT;
`
