-- +migrate Up

CREATE INDEX "task_watcher_usernames_idx" ON "task"(watcher_usernames);
CREATE INDEX "task_resolver_usernames_idx" ON "task"(resolver_usernames);


-- +migrate Down

DROP INDEX "task_resolver_usernames_idx";
DROP INDEX "task_watcher_usernames_idx";
