BEGIN;

CREATE TYPE state_enum AS ENUM('RECEIVED','PROCESSING','DONE');

ALTER TABLE tasks ADD COLUMN state state_enum;

CREATE INDEX IF NOT EXISTS idx_task_state ON tasks(state);

COMMIT;