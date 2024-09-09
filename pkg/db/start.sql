CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TYPE task_running_status AS ENUM(
    'INITIATED',
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

CREATE SEQUENCE task_status_update_seq;

CREATE TABLE task_status_updates_log(
    "task_status_record_id" uuid primary key default gen_random_uuid(),
    "task_id" uuid references "public"."tasks"("task_id"),
    "status" task_running_status not null,
    "status_order" INT NOT NULL DEFAULT nextval('task_status_update_seq'),
    "data" json,
    "created_at" timestamptz default now(),
    "updated_at" timestamptz default now()
);

CREATE INDEX task_status_updates_log_index 
    ON task_status_updates_log ("task_id", "status") INCLUDE ("status_order");

ALTER TABLE "public"."task_status_updates_log"
ADD CONSTRAINT check_log_status_order
CHECK (
    (status = 'INITIATED' AND status_order = 1) OR
    (status = 'RUNNING' AND status_order = 2) OR
    (status = 'SUCCESS' AND status_order IN (3, 4)) OR
    (status = 'FAILED' AND status_order IN (3, 4))
);

ALTER TABLE "public"."task_status_updates_log" 
ADD CONSTRAINT unique_log_status_order 
UNIQUE (task_id, status_order);

CREATE FUNCTION record_updated_at() RETURNS trigger
   LANGUAGE plpgsql AS
$$BEGIN
   NEW.updated_at := current_timestamp;
   RETURN NEW;
END;$$;

CREATE TRIGGER tasks_updated_at BEFORE UPDATE ON "public"."tasks"
   FOR EACH ROW EXECUTE PROCEDURE record_updated_at();

CREATE TRIGGER task_status_updated_at BEFORE UPDATE ON "public"."task_status_updates_log"
   FOR EACH ROW EXECUTE PROCEDURE record_updated_at();

-- UPDATE tasks whenever the right rows are
-- appended on to task_status_updates_log
CREATE OR REPLACE FUNCTION update_task_status()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.status IN ('INITIATED', 'RUNNING') THEN
        UPDATE tasks
        SET status = 'RUNNING'
        WHERE task_id = NEW.task_id;
    ELSIF NEW.status IN ('SUCCESS', 'FAILED') THEN
        UPDATE tasks
        SET status = 'COMPLETED'
        AND completed_at = now()
        WHERE task_id = NEW.task_id;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_task_status_trigger
AFTER INSERT ON "public"."task_status_updates_log"
FOR EACH ROW
EXECUTE FUNCTION update_task_status();

CREATE TYPE worker_status AS ENUM(
    'RUNNING',
    'ERRORED'
);

CREATE TABLE live_workers(
    worker_id SERIAL PRIMARY KEY,
    service_id text not null,
    "status" worker_status NOT NULL DEFAULT 'RUNNING',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()    
);

CREATE TRIGGER live_workers_updated_at BEFORE UPDATE ON "public"."live_workers"
   FOR EACH ROW EXECUTE PROCEDURE record_updated_at();
