# Local Postgres from docker-compose.yml. Credentials are throwaway, so they're
# inline on purpose — this env can never point at anything but the local container.
env "local" {
  src = "file://schema.sql"
  dev = "docker://postgres/17/dev?search_path=public"
  url = "postgres://postgres:postgres@localhost:5432/shapehill?search_path=public&sslmode=disable"

  migration {
    dir = "file://migrations"
  }
}

# It must be the *direct* connection (port 5432), not the pooler (6543):
# transaction pooling breaks DDL and locks.
env "supabase" {
  src = "file://schema.sql"
  dev = "docker://postgres/17/dev?search_path=public"
  url = getenv("SUPABASE_DATABASE_URL")

  migration {
    dir = "file://migrations"
  }
}
