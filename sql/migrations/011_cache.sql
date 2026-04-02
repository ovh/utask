-- +migrate Up

CREATE TABLE "cache" (
    "key" TEXT PRIMARY KEY,
    "value" BYTEA NOT NULL,
    "expires_at" TIMESTAMP WITH TIME ZONE
);

CREATE INDEX "cache_expires_at_idx" ON "cache" ("expires_at") WHERE "expires_at" IS NOT NULL;

INSERT INTO "utask_sql_migrations" VALUES ('v1.21.1-migration011');

-- +migrate Down

DROP INDEX IF EXISTS "cache_expires_at_idx";
DROP TABLE IF EXISTS "cache";

DELETE FROM "utask_sql_migrations" WHERE current_migration_applied = 'v1.21.1-migration011';
