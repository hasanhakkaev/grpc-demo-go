BEGIN;

CREATE TYPE state_enum AS ENUM('RECEIVED','PROCESSING','DONE');

ALTER TABLE yqapp_demo_schema.tasks ADD COLUMN state state_enum;

CREATE INDEX IF NOT EXISTS idx_task_state ON yqapp_demo_schema.tasks(state);

COMMIT;