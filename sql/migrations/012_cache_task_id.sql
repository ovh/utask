-- +migrate Up

ALTER TABLE "cache" ADD "owner" UUID;
CREATE INDEX "cache_owner_idx" ON "cache" ("owner");

INSERT INTO "utask_sql_migrations" VALUES ('v1.21.1-migration012');

-- +migrate Down

DROP INDEX "cache_owner_idx";
ALTER TABLE "cache" DROP COLUMN "owner" UUID;

DELETE FROM "utask_sql_migrations" WHERE current_migration_applied = 'v1.21.1-migration012';
