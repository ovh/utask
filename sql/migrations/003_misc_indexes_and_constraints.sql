-- +migrate Up

DROP INDEX "task_resolver_usernames_idx", "task_watcher_usernames_idx";

-- See section 8.14.4 relative to jsonb indexing:
-- https://www.postgresql.org/docs/9.4/datatype-json.html
CREATE INDEX "task_watcher_usernames_idx" ON "task" USING gin (watcher_usernames jsonb_path_ops);
CREATE INDEX "task_resolver_usernames_idx" ON "task" USING gin (resolver_usernames jsonb_path_ops);

ALTER TABLE "task_template" ALTER COLUMN "variables" SET DEFAULT 'null', ALTER COLUMN "variables" SET NOT NULL;
ALTER TABLE "task_template" ALTER COLUMN "allowed_resolver_usernames" SET DEFAULT '[]', ALTER COLUMN "allowed_resolver_usernames" SET NOT NULL;

ALTER TABLE "task" ALTER COLUMN "watcher_usernames" SET DEFAULT 'null', ALTER COLUMN "watcher_usernames" SET NOT NULL;
ALTER TABLE "task" ALTER COLUMN "resolver_usernames" SET DEFAULT 'null', ALTER COLUMN "resolver_usernames" SET NOT NULL;

-- +migrate Down

ALTER TABLE "task" ALTER COLUMN "resolver_usernames" DROP DEFAULT, ALTER COLUMN "resolver_usernames" DROP NOT NULL;
ALTER TABLE "task" ALTER COLUMN "watcher_usernames" DROP DEFAULT, ALTER COLUMN "watcher_usernames" DROP NOT NULL;

ALTER TABLE "task_template" ALTER COLUMN "allowed_resolver_usernames" DROP DEFAULT, ALTER COLUMN "allowed_resolver_usernames" DROP NOT NULL;
ALTER TABLE "task_template" ALTER COLUMN "variables" DROP DEFAULT, ALTER COLUMN "variables" DROP NOT NULL;

DROP INDEX "task_resolver_usernames_idx", "task_watcher_usernames_idx";

CREATE INDEX "task_watcher_usernames_idx" ON "task"(watcher_usernames);
CREATE INDEX "task_resolver_usernames_idx" ON "task"(resolver_usernames);
