-- Source of truth for the `public` schema. Edit here, then:
--   atlas migrate diff --env local "<name>"
--
-- Supabase manages auth/storage/extensions itself; keep everything in `public`
-- and Atlas will leave those alone.

CREATE TABLE users (
  id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  auth_user_id uuid UNIQUE,
  email        text NOT NULL,
  username     text NOT NULL,
  name         text,
  created_at   timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX users_email_key ON users (lower(email));
CREATE UNIQUE INDEX users_username_key ON users (lower(username));

CREATE TABLE hills (
  id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  owner_id    uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  slug        text NOT NULL UNIQUE,
  title       text NOT NULL,
  description text,
  is_public   boolean NOT NULL DEFAULT false,
  starts_on   date,
  ends_on     date,
  created_at  timestamptz NOT NULL DEFAULT now(),
  updated_at  timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT hills_dates_ordered CHECK (ends_on IS NULL OR starts_on IS NULL OR ends_on >= starts_on)
);

CREATE INDEX hills_owner_id_idx ON hills (owner_id);

CREATE TABLE scopes (
  id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  hill_id     uuid NOT NULL REFERENCES hills (id) ON DELETE CASCADE,
  title       text NOT NULL,
  description text,
  color       text NOT NULL DEFAULT '#2F4C64',
  sort_order  smallint NOT NULL DEFAULT 0,
  created_at  timestamptz NOT NULL DEFAULT now(),
  archived_at timestamptz
);

CREATE INDEX scopes_hill_id_idx ON scopes (hill_id);

-- Append-only: moving a dot inserts a row, never updates one. A scope's current
-- position is its newest row; the rows behind it are the trail.
CREATE TABLE scope_positions (
  id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  scope_id   uuid NOT NULL REFERENCES scopes (id) ON DELETE CASCADE,
  position   smallint NOT NULL CHECK (position BETWEEN 0 AND 100),
  note       text,
  moved_by   uuid REFERENCES users (id) ON DELETE SET NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX scope_positions_latest_idx ON scope_positions (scope_id, created_at DESC);
