-- +migrate Up

CREATE TABLE "task_metadata" (
    id BIGSERIAL PRIMARY KEY,
    public_id UUID UNIQUE NOT NULL,
    id_task BIGINT NOT NULL REFERENCES "task"(id) ON DELETE CASCADE,
    created TIMESTAMP with time zone DEFAULT now() NOT NULL,
    updated TIMESTAMP with time zone DEFAULT now() NOT NULL,
    "key" TEXT NOT NULL,
    "value" JSONB,
);

INSERT INTO "utask_sql_migrations" VALUES ('v1.21.2-migration011');

-- +migrate Down

DROP TABLE IF EXISTS "task_metadata" CASCADE;

DELETE FROM "utask_sql_migrations" WHERE current_migration_applied = 'v1.21.2-migration011';
