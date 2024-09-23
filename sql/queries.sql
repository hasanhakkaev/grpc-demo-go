-- name: CreateTask :one
INSERT INTO tasks (type, value, state, creation_time, last_update_time)
VALUES ($1, $2, $3, $4, $5)
RETURNING id;

-- name: UpdateTaskState :one
UPDATE tasks
SET state = $1, last_update_time = $2
WHERE id = $3
RETURNING id, type, value, state, creation_time, last_update_time;

-- name: GetTasksByState :many
SELECT id, type, value, state, creation_time, last_update_time
FROM tasks
WHERE state = $1;

-- name: GetSumOfTasksByState :many
SELECT state, COUNT(*) AS task_count
FROM tasks
GROUP BY state;

-- name: GetSumOfValues :many
SELECT type, SUM(value) AS total_value
FROM tasks
GROUP BY type;
