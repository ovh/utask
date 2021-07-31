-- +migrate Up

ALTER TABLE "task_template" ADD COLUMN "allowed_resolver_groups" JSONB NOT NULL DEFAULT '[]';

INSERT INTO "utask_sql_migrations" VALUES ('v1.18.2-migration007');

-- +migrate Down

ALTER TABLE "task_template" DROP COLUMN "allowed_resolver_groups";

UPDATE "utask_sql_migrations" SET "current_migration_applied" = 'v1.18.2-migration007';