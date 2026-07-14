-- Create "users" table
CREATE TABLE "users" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "auth_user_id" uuid NULL,
  "email" text NOT NULL,
  "username" text NOT NULL,
  "name" text NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "users_auth_user_id_key" UNIQUE ("auth_user_id")
);
-- Create index "users_email_key" to table: "users"
CREATE UNIQUE INDEX "users_email_key" ON "users" ((lower(email)));
-- Create index "users_username_key" to table: "users"
CREATE UNIQUE INDEX "users_username_key" ON "users" ((lower(username)));
-- Create "hills" table
CREATE TABLE "hills" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "owner_id" uuid NOT NULL,
  "slug" text NOT NULL,
  "title" text NOT NULL,
  "description" text NULL,
  "is_public" boolean NOT NULL DEFAULT false,
  "starts_on" date NULL,
  "ends_on" date NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "hills_slug_key" UNIQUE ("slug"),
  CONSTRAINT "hills_owner_id_fkey" FOREIGN KEY ("owner_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "hills_dates_ordered" CHECK ((ends_on IS NULL) OR (starts_on IS NULL) OR (ends_on >= starts_on))
);
-- Create index "hills_owner_id_idx" to table: "hills"
CREATE INDEX "hills_owner_id_idx" ON "hills" ("owner_id");
-- Create "scopes" table
CREATE TABLE "scopes" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "hill_id" uuid NOT NULL,
  "title" text NOT NULL,
  "description" text NULL,
  "color" text NOT NULL DEFAULT '#2F4C64',
  "sort_order" smallint NOT NULL DEFAULT 0,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "archived_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "scopes_hill_id_fkey" FOREIGN KEY ("hill_id") REFERENCES "hills" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "scopes_hill_id_idx" to table: "scopes"
CREATE INDEX "scopes_hill_id_idx" ON "scopes" ("hill_id");
-- Create "scope_positions" table
CREATE TABLE "scope_positions" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "scope_id" uuid NOT NULL,
  "position" smallint NOT NULL,
  "note" text NULL,
  "moved_by" uuid NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "scope_positions_moved_by_fkey" FOREIGN KEY ("moved_by") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "scope_positions_scope_id_fkey" FOREIGN KEY ("scope_id") REFERENCES "scopes" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "scope_positions_position_check" CHECK (("position" >= 0) AND ("position" <= 100))
);
-- Create index "scope_positions_latest_idx" to table: "scope_positions"
CREATE INDEX "scope_positions_latest_idx" ON "scope_positions" ("scope_id", "created_at" DESC);
