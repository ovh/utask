-- +migrate Up

ALTER TABLE "task_template" ADD COLUMN "allowed_resolver_groups" JSONB NOT NULL DEFAULT '[]';
ALTER TABLE "task" ADD COLUMN "resolver_groups" JSONB NOT NULL DEFAULT 'null';
ALTER TABLE "task" ADD COLUMN "watcher_groups" JSONB NOT NULL DEFAULT 'null';

CREATE INDEX "task_resolver_groups_idx" ON "task" USING gin (resolver_groups jsonb_path_ops);
CREATE INDEX "task_watcher_groups_idx" ON "task" USING gin (watcher_groups jsonb_path_ops);

INSERT INTO "utask_sql_migrations" VALUES ('v1.19.0-migration007');

-- +migrate Down

ALTER TABLE "task_template" DROP COLUMN "allowed_resolver_groups";
ALTER TABLE "task" DROP COLUMN "resolver_groups";
ALTER TABLE "task" DROP COLUMN "watcher_groups";

DELETE FROM "utask_sql_migrations" WHERE current_migration_applied = 'v1.19.0-migration007';
