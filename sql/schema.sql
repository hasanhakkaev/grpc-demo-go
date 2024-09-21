DROP TYPE if EXISTS state;
CREATE TYPE state AS ENUM('RECEIVED','PROCESSING','DONE');

DROP TABLE if EXISTS tasks;
CREATE TABLE IF NOT EXISTS tasks (
    id INT PRIMARY KEY,               -- Unique identifier for the task
    type INT CHECK (type BETWEEN 0 AND 9), -- Task type (integer between 0 and 9)
    value INT CHECK (value BETWEEN 0 AND 99), -- Task value (integer between 0 and 99)
    state STATE NOT NULL,            -- Task state (received, processing, done)
    creation_time FLOAT NOT NULL,      -- Timestamp for when the task was created
    last_update_time FLOAT NOT NULL    -- Timestamp for the last update to the task
    );

CREATE INDEX IF NOT EXISTS idx_task_state ON tasks(state);

CREATE INDEX IF NOT EXISTS idx_task_type ON tasks(type);

