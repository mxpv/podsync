package storage

//noinspection SpellCheckingInspection
const pgsql = `
BEGIN;

-- Pledges

CREATE TABLE IF NOT EXISTS pledges (
  pledge_id BIGSERIAL PRIMARY KEY,
  patron_id BIGINT NOT NULL UNIQUE,
  created_at TIMESTAMPTZ NOT NULL,
  declined_since TIMESTAMPTZ NULL,
  amount_cents INT NOT NULL,
  total_historical_amount_cents INT,
  outstanding_payment_amount_cents INT,
  is_paused BOOLEAN
);

CREATE INDEX IF NOT EXISTS patron_id_idx ON pledges(patron_id);

-- Feeds

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'provider') THEN
    CREATE TYPE provider AS ENUM ('youtube', 'vimeo');
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'link_type') THEN
    CREATE TYPE link_type AS ENUM ('channel', 'playlist', 'user', 'group');
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'quality') THEN
    CREATE TYPE quality AS ENUM ('low', 'high');
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'format') THEN
    CREATE TYPE format AS ENUM ('video', 'audio');
  END IF;
END
$$;

CREATE TABLE IF NOT EXISTS feeds (
  feed_id BIGSERIAL PRIMARY KEY,
  hash_id VARCHAR(12) NOT NULL UNIQUE,
  user_id VARCHAR(32) NULL,
  item_id VARCHAR(64) NOT NULL CHECK (item_id <> ''),
  provider provider NOT NULL,
  link_type link_type NOT NULL,
  page_size INT NOT NULL DEFAULT 50,
  format format NOT NULL DEFAULT 'video',
  quality quality NOT NULL DEFAULT 'high',
  feature_level INT NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  last_access TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS feeds_hash_id_idx ON feeds(hash_id);
CREATE INDEX IF NOT EXISTS feeds_user_id_idx ON feeds(user_id);

COMMIT;
END;
`
