-- +migrate Up

UPDATE "resolution" SET "steps_compression_alg" = '' WHERE "steps_compression_alg" IS NULL;

ALTER TABLE "resolution" ALTER COLUMN "steps_compression_alg" SET NOT NULL;
ALTER TABLE "resolution" ALTER COLUMN "steps_compression_alg" SET DEFAULT '';

INSERT INTO "utask_sql_migrations" VALUES ('v1.21.1-migration010');

-- +migrate Down

ALTER TABLE "resolution" ALTER COLUMN "steps_compression_alg" DROP NOT NULL;
ALTER TABLE "resolution" ALTER COLUMN "steps_compression_alg" DROP DEFAULT;

DELETE FROM "utask_sql_migrations" WHERE current_migration_applied = 'v1.21.1-migration010';
