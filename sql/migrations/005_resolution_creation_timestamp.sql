-- +migrate Up

-- Adds the resolution creation timestamp on resolution table
ALTER TABLE "resolution" ADD COLUMN "created" TIMESTAMP with time zone DEFAULT now() NOT NULL;

-- +migrate Down

ALTER TABLE "resolution" DROP COLUMN "created";