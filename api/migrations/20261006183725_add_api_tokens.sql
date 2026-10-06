-- Create "api_tokens" table
CREATE TABLE "api_tokens" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "user_id" uuid NOT NULL,
  "name" text NOT NULL,
  "hint" text NOT NULL,
  "token_hash" bytea NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "last_used_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "api_tokens_token_hash_key" UNIQUE ("token_hash"),
  CONSTRAINT "api_tokens_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "api_tokens_user_id_idx" to table: "api_tokens"
CREATE INDEX "api_tokens_user_id_idx" ON "api_tokens" ("user_id");
