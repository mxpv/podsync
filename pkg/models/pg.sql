BEGIN;

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

CREATE INDEX patron_id_idx ON pledges(patron_id);

COMMIT;