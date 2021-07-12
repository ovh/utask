-- +migrate Up

ALTER TABLE "task_template" ADD COLUMN "allowed_resolver_groups" JSONB NOT NULL DEFAULT '[]';

-- +migrate Down

ALTER TABLE "task_template" DROP COLUMN "allowed_resolver_groups";