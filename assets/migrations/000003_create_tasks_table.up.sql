CREATE TABLE IF NOT EXISTS yqapp_demo_schema.tasks (
                                     id SERIAL PRIMARY KEY,               -- Unique identifier for the task (auto-incrementing integer)
                                     type INT CHECK (type BETWEEN 0 AND 9), -- Task type (integer between 0 and 9)
                                     value INT CHECK (value BETWEEN 0 AND 99), -- Task value (integer between 0 and 99)
                                     creation_time FLOAT NOT NULL,      -- Timestamp for when the task was created
                                     last_update_time FLOAT NOT NULL    -- Timestamp for the last update to the task
);

CREATE INDEX IF NOT EXISTS idx_task_type ON yqapp_demo_schema.tasks(type);
