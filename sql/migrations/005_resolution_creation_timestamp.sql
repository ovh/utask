-- +migrate Up

-- Adds the resolution creation timestamp on resolution table and sets its value to last_start
ALTER TABLE "resolution" ADD COLUMN "created" TIMESTAMP with time zone;
UPDATE "resolution" SET created = last_start;
ALTER TABLE "resolution" ALTER COLUMN "created" SET NOT NULL;

-- +migrate Down

ALTER TABLE "resolution" DROP COLUMN "created";