// Code generated by sqlc. DO NOT EDIT.
// source: queries.sql

package database

import (
	"context"
	"database/sql"
)

const createTask = `-- name: CreateTask :one
INSERT INTO tasks (id,type, value, state, creation_time, last_update_time)
VALUES ($1, $2, $3, $4, $5, $6)
    RETURNING id, type, value, state, creation_time, last_update_time
`

type CreateTaskParams struct {
	ID             int32
	Type           sql.NullInt32
	Value          sql.NullInt32
	State          State
	CreationTime   float64
	LastUpdateTime float64
}

func (q *Queries) CreateTask(ctx context.Context, arg CreateTaskParams) (Task, error) {
	row := q.db.QueryRowContext(ctx, createTask,
		arg.ID,
		arg.Type,
		arg.Value,
		arg.State,
		arg.CreationTime,
		arg.LastUpdateTime,
	)
	var i Task
	err := row.Scan(
		&i.ID,
		&i.Type,
		&i.Value,
		&i.State,
		&i.CreationTime,
		&i.LastUpdateTime,
	)
	return i, err
}

const getSumOfTasksByState = `-- name: GetSumOfTasksByState :many
SELECT state, COUNT(*)
FROM tasks
GROUP BY state
`

type GetSumOfTasksByStateRow struct {
	State State
	Count int64
}

// Get the total number of tasks per state
func (q *Queries) GetSumOfTasksByState(ctx context.Context) ([]GetSumOfTasksByStateRow, error) {
	rows, err := q.db.QueryContext(ctx, getSumOfTasksByState)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetSumOfTasksByStateRow
	for rows.Next() {
		var i GetSumOfTasksByStateRow
		if err := rows.Scan(&i.State, &i.Count); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getSumOfValues = `-- name: GetSumOfValues :many
SELECT type, SUM(value)
FROM tasks
GROUP BY type
`

type GetSumOfValuesRow struct {
	Type sql.NullInt32
	Sum  int64
}

// Get the total sum of task values for each task type
func (q *Queries) GetSumOfValues(ctx context.Context) ([]GetSumOfValuesRow, error) {
	rows, err := q.db.QueryContext(ctx, getSumOfValues)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetSumOfValuesRow
	for rows.Next() {
		var i GetSumOfValuesRow
		if err := rows.Scan(&i.Type, &i.Sum); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getTask = `-- name: GetTask :one
SELECT id, type, value, state, creation_time, last_update_time
FROM tasks
WHERE id = $1
LIMIT 1
`

func (q *Queries) GetTask(ctx context.Context, id int32) (Task, error) {
	row := q.db.QueryRowContext(ctx, getTask, id)
	var i Task
	err := row.Scan(
		&i.ID,
		&i.Type,
		&i.Value,
		&i.State,
		&i.CreationTime,
		&i.LastUpdateTime,
	)
	return i, err
}

const getTasksByState = `-- name: GetTasksByState :many
SELECT id, type, value, state, creation_time, last_update_time
FROM tasks
WHERE state = $1
`

// Get all tasks in a specific state
func (q *Queries) GetTasksByState(ctx context.Context, state State) ([]Task, error) {
	rows, err := q.db.QueryContext(ctx, getTasksByState, state)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Task
	for rows.Next() {
		var i Task
		if err := rows.Scan(
			&i.ID,
			&i.Type,
			&i.Value,
			&i.State,
			&i.CreationTime,
			&i.LastUpdateTime,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateTask = `-- name: UpdateTask :one
UPDATE tasks
SET state = $1, last_update_time = $2
WHERE id = $3
RETURNING id, type, value, state, creation_time, last_update_time
`

type UpdateTaskParams struct {
	State          State
	LastUpdateTime float64
	ID             int32
}

func (q *Queries) UpdateTask(ctx context.Context, arg UpdateTaskParams) (Task, error) {
	row := q.db.QueryRowContext(ctx, updateTask, arg.State, arg.LastUpdateTime, arg.ID)
	var i Task
	err := row.Scan(
		&i.ID,
		&i.Type,
		&i.Value,
		&i.State,
		&i.CreationTime,
		&i.LastUpdateTime,
	)
	return i, err
}
