CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TYPE task_running_status AS ENUM(
    'RUNNING',
    'SUCCESS',
    'FAILED'
);

CREATE TYPE task_status AS ENUM(
    'CREATED',
    'RUNNING',
    'COMPLETED'
);

CREATE TABLE tasks(
    "task_id" uuid primary key default gen_random_uuid(),
    "command" text not null,
    "status" task_status default 'CREATED',
    "created_at" timestamptz default now(),
    "updated_at" timestamptz default now(),
    "completed_at" timestamptz
);

CREATE INDEX task_id_status_index 
    ON tasks ("task_id", "status");

CREATE TABLE task_status_updates_log(
    "task_status_record_id" uuid primary key default gen_random_uuid(),
    "task_id" uuid references "tasks"("task_id"),
    "status" task_running_status not null,
    "data" json,
    "worker_id" text,
    "created_at" timestamptz default now(),
    "updated_at" timestamptz default now()
);

ALTER TABLE task_status_updates_log 
    ADD CONSTRAINT unique_status_task_id UNIQUE (task_id, status);

CREATE FUNCTION record_updated_at() RETURNS trigger
   LANGUAGE plpgsql AS
$$BEGIN
   NEW.updated_at := current_timestamp;
   RETURN NEW;
END;$$;

CREATE TRIGGER tasks_updated_at BEFORE UPDATE ON "tasks"
   FOR EACH ROW EXECUTE PROCEDURE record_updated_at();

CREATE TRIGGER task_status_updated_at BEFORE UPDATE ON "task_status_updates_log"
   FOR EACH ROW EXECUTE PROCEDURE record_updated_at();

-- UPDATE tasks whenever the right rows are
-- appended on to task_status_updates_log
CREATE OR REPLACE FUNCTION update_task_status()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.status = 'RUNNING' THEN
        UPDATE tasks
        SET status = 'RUNNING'
        WHERE task_id = NEW.task_id AND "status" not in ('COMPLETED');
    ELSIF NEW.status IN ('SUCCESS', 'FAILED') THEN
        UPDATE tasks
        SET "status" = 'COMPLETED', completed_at = now()
        WHERE task_id = NEW.task_id;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_task_status_trigger
AFTER INSERT ON "task_status_updates_log"
FOR EACH ROW EXECUTE FUNCTION update_task_status();

CREATE TYPE worker_status AS ENUM(
    'RUNNING',
    'ERRORED'
);

CREATE TABLE live_workers(
    worker_id SERIAL PRIMARY KEY,
    service_id text not null UNIQUE,
    host text not null,
    port bigint not null,
    "status" worker_status NOT NULL DEFAULT 'RUNNING',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    connected_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (host, port)
);

CREATE TRIGGER live_workers_updated_at BEFORE UPDATE ON "live_workers"
   FOR EACH ROW EXECUTE PROCEDURE record_updated_at();
