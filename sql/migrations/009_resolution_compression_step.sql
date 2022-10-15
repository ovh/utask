-- +migrate Up

ALTER TABLE "resolution" ADD COLUMN "steps_compression_alg" TEXT;

INSERT INTO "utask_sql_migrations" VALUES ('v1.21.0-migration009');

-- +migrate Down

ALTER TABLE "resolution" DROP COLUMN "steps_compression_alg";

DELETE FROM "utask_sql_migrations" WHERE current_migration_applied = 'v1.21.0-migration009';
