# Diff/lint/test against a throwaway Docker Postgres. No `url` on purpose:
# without it, `migrate apply` on this env fails, so it can never touch Supabase.
env "local" {
  src = "file://schema.sql"
  dev = "docker://postgres/17/dev?search_path=public"

  migration {
    dir = "file://migrations"
  }
}

# Applies migrations to Supabase. DATABASE_URL must be the *direct* connection
# (port 5432), not the pooler (6543) — transaction pooling breaks DDL and locks.
env "supabase" {
  src = "file://schema.sql"
  dev = "docker://postgres/17/dev?search_path=public"
  url = getenv("DATABASE_URL")

  migration {
    dir = "file://migrations"
  }
}
