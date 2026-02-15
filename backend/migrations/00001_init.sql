-- +goose Up
CREATE TABLE IF NOT EXISTS schema_migrations_audit (
    id BIGSERIAL PRIMARY KEY,
    migration_name TEXT NOT NULL,
    applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS schema_migrations_audit;
