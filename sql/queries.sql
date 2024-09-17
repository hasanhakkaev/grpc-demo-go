-- name: CreateTask :one
INSERT INTO tasks (type, value, state, creation_time, last_update_time)
VALUES ($1, $2, $3, $4, $5)
    RETURNING id;

-- name: GetTask :one
SELECT id, type, value, state, creation_time, last_update_time
FROM tasks
WHERE id = $1
LIMIT 1;

-- name: UpdateTask :one
UPDATE tasks
SET state = $1, last_update_time = $2
WHERE id = $3
RETURNING *;

-- Get all tasks in a specific state
-- name: GetTasksByState :many
SELECT id, type, value, state, creation_time, last_update_time
FROM tasks
WHERE state = $1;

-- Get the total number of tasks per state
-- name: GetSumOfTasksByState :many
SELECT state, COUNT(*)
FROM tasks
GROUP BY state;

-- Get the total sum of task values for each task type
-- name: GetSumOfValues :many
SELECT type, SUM(value)
FROM tasks
GROUP BY type;
