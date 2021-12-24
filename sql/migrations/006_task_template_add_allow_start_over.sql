-- +migrate Up

ALTER TABLE "task_template" ADD COLUMN "allow_task_start_over" BOOL NOT NULL DEFAULT false;

INSERT INTO "utask_sql_migrations" VALUES ('v1.17.0-migration006');

-- +migrate Down

ALTER TABLE "task_template" DROP COLUMN "allow_task_start_over";

DELETE FROM "utask_sql_migrations" WHERE current_migration_applied = 'v1.17.0-migration006';
