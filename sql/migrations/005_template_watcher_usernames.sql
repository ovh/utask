-- +migrate Up

ALTER TABLE "task_template" ADD COLUMN "allowed_watcher_usernames" JSONB NOT NULL DEFAULT '[]';

-- +migrate Down

ALTER TABLE "task_template" DROP COLUMN "allowed_watcher_usernames";
