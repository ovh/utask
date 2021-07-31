-- +migrate Up

ALTER TABLE "task_template" ADD COLUMN "allowed_resolver_groups" JSONB NOT NULL DEFAULT '[]';

UPDATE "utask_sql_migrations" SET "current_migration_applied" = 'v1.16.1-migration006';

-- +migrate Down

ALTER TABLE "task_template" DROP COLUMN "allowed_resolver_groups";

UPDATE "utask_sql_migrations" SET "current_migration_applied" = 'v1.10.0-migration005';