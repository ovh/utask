-- +migrate Up

CREATE INDEX "task_watcher_usernames_idx" ON "task"(watcher_usernames);
CREATE INDEX "task_resolver_usernames_idx" ON "task"(resolver_usernames);
CREATE INDEX "task_comment_id_task_idx" ON "task_comment"(id_task);

-- +migrate Down

DROP INDEX "task_comment_id_task_idx";
DROP INDEX "task_resolver_usernames_idx";
DROP INDEX "task_watcher_usernames_idx";
