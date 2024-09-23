DROP TYPE if EXISTS state;
CREATE TYPE state AS ENUM('RECEIVED','PROCESSING','DONE');

DROP TABLE if EXISTS tasks;
CREATE TABLE IF NOT EXISTS tasks (
                                     id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
                                     type INT NOT NULL CHECK (type BETWEEN 0 AND 9), -- Task type (between 0 and 9)
                                     value INT NOT NULL CHECK (value BETWEEN 0 AND 99), -- Task value (between 0 and 99)
                                     state STATE NOT NULL,            -- Task state (enum with values 'RECEIVED', 'PROCESSING', 'DONE')
                                     creation_time FLOAT NOT NULL,         -- Creation time as a Unix timestamp (float)
                                     last_update_time FLOAT NOT NULL       -- Last update time as a Unix timestamp (float)              -- Timestamp for the last update to the task
);


CREATE INDEX IF NOT EXISTS idx_task_state ON tasks(state);

CREATE INDEX IF NOT EXISTS idx_task_type ON tasks(type);

