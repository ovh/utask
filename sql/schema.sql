BEGIN;

DROP TABLE IF EXISTS "task_template" CASCADE;
DROP TABLE IF EXISTS "batch" CASCADE;
DROP TABLE IF EXISTS "task" CASCADE;
DROP TABLE IF EXISTS "task_comment" CASCADE;
DROP TABLE IF EXISTS "resolution" CASCADE;
DROP TABLE IF EXISTS "runner_instance" CASCADE;

CREATE TABLE "task_template" (
    id BIGSERIAL PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    description TEXT NOT NULL,
    long_description TEXT,
    doc_link TEXT,
    inputs JSONB NOT NULL,
    resolver_inputs JSONB NOT NULL,
    steps JSONB NOT NULL,
    variables JSONB NOT NULL DEFAULT 'null',
    allowed_resolver_usernames JSONB NOT NULL DEFAULT '[]',
    allow_all_resolver_usernames BOOL NOT NULL DEFAULT false,
    auto_runnable BOOL NOT NULL DEFAULT false,
    blocked BOOL NOT NULL DEFAULT false,
    hidden BOOL NOT NULL DEFAULT false,
    result_format JSONB NOT NULL,
    title_format TEXT NOT NULL,
    retry_max INTEGER,
    base_configurations JSONB NOT NULL,
    tags JSONB NOT NULL DEFAULT 'null'
);

CREATE TABLE "batch" (
    id BIGSERIAL PRIMARY KEY,
    public_id UUID UNIQUE NOT NULL
);

CREATE TABLE "task" (
    id BIGSERIAL PRIMARY KEY,
    public_id UUID UNIQUE NOT NULL,
    id_template BIGINT NOT NULL REFERENCES "task_template"(id),
    id_batch BIGINT REFERENCES "batch"(id),
    title TEXT NOT NULL,
    requester_username TEXT,
    watcher_usernames JSONB NOT NULL DEFAULT 'null',
    resolver_usernames JSONB NOT NULL DEFAULT 'null',
    created TIMESTAMP with time zone DEFAULT now() NOT NULL,
    last_activity TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    state TEXT NOT NULL,
    steps_done INTEGER NOT NULL,
    steps_total INTEGER NOT NULL,
    crypt_key BYTEA NOT NULL,
    encrypted_input BYTEA NOT NULL,
    encrypted_result BYTEA NOT NULL,
    tags JSONB NOT NULL DEFAULT 'null'
);

CREATE INDEX ON "task"(id_template);
CREATE INDEX ON "task"(id_batch);
CREATE INDEX ON "task"(requester_username);
CREATE INDEX ON "task"(state);
CREATE INDEX ON "task"(last_activity DESC);
-- See section 8.14.4 relative to jsonb indexing:
-- https://www.postgresql.org/docs/9.4/datatype-json.html
CREATE INDEX ON "task" USING gin (watcher_usernames jsonb_path_ops);
CREATE INDEX ON "task" USING gin (resolver_usernames jsonb_path_ops);
CREATE INDEX ON "task" USING gin (tags jsonb_path_ops);

CREATE TABLE "task_comment" (
    id BIGSERIAL PRIMARY KEY,
    public_id UUID UNIQUE NOT NULL,
    id_task BIGINT NOT NULL REFERENCES "task"(id) ON DELETE CASCADE,
    username TEXT,
    created TIMESTAMP with time zone DEFAULT now() NOT NULL,
    updated TIMESTAMP with time zone DEFAULT now() NOT NULL,
    content TEXT NOT NULL
);
CREATE INDEX ON "task_comment"(id_task);

CREATE TABLE "resolution" (
    id BIGSERIAL PRIMARY KEY,
    public_id UUID UNIQUE NOT NULL,
    id_task BIGINT UNIQUE NOT NULL REFERENCES "task"(id) ON DELETE CASCADE,
    resolver_username TEXT,
    state TEXT NOT NULL,
    instance_id BIGINT,
    last_start TIMESTAMP with time zone,
    last_stop TIMESTAMP with time zone,
    next_retry TIMESTAMP with time zone,
    run_count INTEGER NOT NULL,
    run_max INTEGER NOT NULL,
    crypt_key BYTEA NOT NULL,
    encrypted_resolver_input BYTEA,
    encrypted_steps BYTEA NOT NULL,
    base_configurations JSONB NOT NULL
);

CREATE INDEX ON "resolution"(resolver_username);
CREATE INDEX ON "resolution"(state);
CREATE INDEX ON "resolution"(instance_id);
CREATE INDEX ON "resolution"(next_retry);

CREATE TABLE "runner_instance" (
    id BIGSERIAL PRIMARY KEY,
    heartbeat TIMESTAMP with time zone DEFAULT now() NOT NULL
);

END;
