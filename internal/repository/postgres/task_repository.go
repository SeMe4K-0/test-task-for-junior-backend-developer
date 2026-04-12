package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, t *taskdomain.Task) (*taskdomain.Task, error) {
	const query = `
	INSERT INTO tasks (
		title,
		description,
		status,
		created_at,
		updated_at,
		recurrence
	)
	VALUES ($1,$2,$3,$4,$5,$6)
	RETURNING
		id, title, description, status,
		created_at, updated_at,
		recurrence
	`

	recJSON, err := encodeRecurrence(t.Recurrence)
	if err != nil {
		return nil, err
	}

	row := r.pool.QueryRow(ctx, query,
		t.Title,
		t.Description,
		t.Status,
		t.CreatedAt,
		t.UpdatedAt,
		recJSON,
	)

	return scanTask(row)
}

func (r *Repository) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	const query = `
	SELECT
		id, title, description, status,
		created_at, updated_at,
		recurrence
	FROM tasks
	WHERE id = $1
	`

	row := r.pool.QueryRow(ctx, query, id)

	task, err := scanTask(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskdomain.ErrNotFound
		}
		return nil, err
	}

	return task, nil
}

func (r *Repository) Update(ctx context.Context, t *taskdomain.Task) (*taskdomain.Task, error) {
	const query = `
	UPDATE tasks SET
		title = $1,
		description = $2,
		status = $3,
		updated_at = $4,
		recurrence = $5
	WHERE id = $6
	RETURNING
		id, title, description, status,
		created_at, updated_at,
		recurrence
	`

	recJSON, err := encodeRecurrence(t.Recurrence)
	if err != nil {
		return nil, err
	}

	row := r.pool.QueryRow(ctx, query,
		t.Title,
		t.Description,
		t.Status,
		t.UpdatedAt,
		recJSON,
		t.ID,
	)

	return scanTask(row)
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	const query = `DELETE FROM tasks WHERE id = $1`

	res, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if res.RowsAffected() == 0 {
		return taskdomain.ErrNotFound
	}

	return nil
}

func (r *Repository) List(ctx context.Context) ([]taskdomain.Task, error) {
	const query = `
	SELECT
		id, title, description, status,
		created_at, updated_at,
		recurrence
	FROM tasks
	ORDER BY id DESC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []taskdomain.Task

	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, *t)
	}

	return result, rows.Err()
}

type scanner interface {
	Scan(dest ...any) error
}

func scanTask(scanner scanner) (*taskdomain.Task, error) {
	var (
		task   taskdomain.Task
		status string
		recRaw []byte
	)

	err := scanner.Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&status,
		&task.CreatedAt,
		&task.UpdatedAt,
		&recRaw,
	)
	if err != nil {
		return nil, err
	}

	task.Status = taskdomain.Status(status)

	rec, err := decodeRecurrence(recRaw)
	if err != nil {
		return nil, err
	}

	task.Recurrence = &rec

	return &task, nil
}