package storage

const installScript = `
BEGIN;

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'provider') THEN
    CREATE TYPE provider AS ENUM ('youtube', 'vimeo');
  END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'link_type') THEN
    CREATE TYPE link_type AS ENUM ('channel', 'playlist', 'user', 'group');
  END IF;
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
	hash_id VARCHAR(12) NOT NULL CHECK (hash_id <> '') UNIQUE,
	user_id VARCHAR(32) NULL,
	item_id VARCHAR(32) NOT NULL CHECK (item_id <> ''),
	provider provider NOT NULL,
	link_type link_type NOT NULL,
	page_size INT NOT NULL DEFAULT 50,
	format format NOT NULL DEFAULT 'video',
	quality quality NOT NULL DEFAULT 'high',
	feature_level INT NOT NULL DEFAULT 0,
	last_access timestamp WITHOUT TIME ZONE NOT NULL
);

COMMIT;
`
