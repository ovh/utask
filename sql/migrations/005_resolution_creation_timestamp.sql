-- +migrate Up

-- Adds the resolution creation timestamp on resolution table and sets its value to last_start
ALTER TABLE "resolution" ADD COLUMN "created" TIMESTAMP with time zone;
UPDATE "resolution" SET created = last_start;
UPDATE "resolution" SET created = NOW() WHERE created IS NULL;
ALTER TABLE "resolution" ALTER COLUMN "created" SET NOT NULL, ALTER COLUMN "created" SET DEFAULT now();

CREATE TABLE "utask_sql_migrations" (
    current_migration_applied TEXT PRIMARY KEY
);

INSERT INTO "utask_sql_migrations" VALUES ('v1.10.0-migration005');

-- +migrate Down

ALTER TABLE "resolution" DROP COLUMN "created";
DROP TABLE "utask_sql_migrations";
