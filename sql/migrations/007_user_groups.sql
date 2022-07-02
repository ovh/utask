-- +migrate Up

ALTER TABLE "task_template" ADD COLUMN "allowed_resolver_groups" JSONB NOT NULL DEFAULT '[]';
ALTER TABLE "task" ADD COLUMN "resolver_groups" JSONB NOT NULL DEFAULT 'null';

CREATE INDEX "task_resolver_groups_idx" ON "task" USING gin (resolver_groups jsonb_path_ops);

INSERT INTO "utask_sql_migrations" VALUES ('v1.18.2-migration007');

-- +migrate Down

ALTER TABLE "task_template" DROP COLUMN "allowed_resolver_groups";
ALTER TABLE "task" DROP COLUMN "resolver_groups";

DROP INDEX "task_resolver_groups_idx";

UPDATE "utask_sql_migrations" SET "current_migration_applied" = 'v1.18.2-migration007';