-- +migrate Up

CREATE TABLE "callback" (
    id BIGSERIAL PRIMARY KEY,
    created TIMESTAMP with time zone DEFAULT now() NOT NULL,
    updated TIMESTAMP with time zone DEFAULT now() NOT NULL,
    called TIMESTAMP with time zone,
    public_id UUID UNIQUE NOT NULL,
    id_task BIGINT NOT NULL REFERENCES "task"(id) ON DELETE CASCADE,
    id_resolution BIGINT NOT NULL REFERENCES "resolution"(id) ON DELETE CASCADE,
    resolver_username TEXT NOT NULL,
    encrypted_schema BYTEA NOT NULL,
    encrypted_body BYTEA NOT NULL,
    encrypted_secret BYTEA NOT NULL
);
CREATE INDEX ON "callback"(id_task);
CREATE INDEX ON "callback"(id_resolution);

INSERT INTO "utask_sql_migrations" VALUES ('v1.20.0-migration008');

-- +migrate Down

DROP TABLE "callback" CASCADE;

DELETE FROM "utask_sql_migrations" WHERE current_migration_applied = 'v1.20.0-migration008';
