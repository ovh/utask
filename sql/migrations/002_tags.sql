-- +migrate Up

ALTER TABLE "task" ADD COLUMN "tags" JSONB;
ALTER TABLE "task_template" ADD COLUMN "tags" JSONB;

-- See section 8.14.4 relative to jsonb indexing:
-- https://www.postgresql.org/docs/9.4/datatype-json.html
CREATE INDEX "task_tags_idx" ON "task" USING gin (tags jsonb_path_ops);

-- +migrate Down

ALTER TABLE "task" DROP COLUMN "tags";
ALTER TABLE "task_template" DROP COLUMN "tags";

DROP INDEX task_tags_idx;
