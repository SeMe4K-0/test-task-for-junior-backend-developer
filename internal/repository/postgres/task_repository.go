package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	taskdomain "example.com/taskservice/internal/domain/task"
)

const (
	CreateQuery = `
		INSERT INTO tasks (title, description, status,
		                   repeated_type, repeated_every_n_days, repeated_day_of_month, repeated_specific_dates,
		                   created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, title, description, status,
		          repeated_type, repeated_every_n_days, repeated_day_of_month, repeated_specific_dates,
		          created_at, updated_at`

	GetByIDQuery = `
		SELECT id, title, description, status,
		       repeated_type, repeated_every_n_days, repeated_day_of_month, repeated_specific_dates,
		       created_at, updated_at
		FROM tasks
		WHERE id = $1`

	UpdateQuery = `
		UPDATE tasks
		SET title = $1, description = $2, status = $3, repeated_type = $4, repeated_every_n_days = $5,
		    repeated_day_of_month = $6, repeated_specific_dates = $7, updated_at = $8
		WHERE id = $9
		RETURNING id, title, description, status, repeated_type, repeated_every_n_days, repeated_day_of_month, 
		repeated_specific_dates, created_at, updated_at`

	DeleteQuery = `DELETE FROM tasks WHERE id = $1`

	ListQuery = `
		SELECT id, title, description, status,
		       repeated_type, repeated_every_n_days, repeated_day_of_month, repeated_specific_dates,
		       created_at, updated_at
		FROM tasks
		ORDER BY id DESC`
)

type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	row := r.pool.QueryRow(ctx, CreateQuery, task.Title, task.Description, task.Status, task.Repeated.Type,
		nullInt(task.Repeated.EveryNDays), nullInt(task.Repeated.DayOfMonth), nullTimes(task.Repeated.SpecificDates),
		task.CreatedAt, task.UpdatedAt)

	created, err := scanTask(row)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (r *Repository) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	row := r.pool.QueryRow(ctx, GetByIDQuery, id)

	found, err := scanTask(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskdomain.ErrNotFound
		}

		return nil, err
	}

	return found, nil
}

func (r *Repository) Update(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	row := r.pool.QueryRow(ctx, UpdateQuery, task.Title, task.Description, task.Status, task.Repeated.Type,
		nullInt(task.Repeated.EveryNDays), nullInt(task.Repeated.DayOfMonth), nullTimes(task.Repeated.SpecificDates),
		task.UpdatedAt, task.ID)

	updated, err := scanTask(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskdomain.ErrNotFound
		}

		return nil, err
	}

	return updated, nil
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	result, err := r.pool.Exec(ctx, DeleteQuery, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return taskdomain.ErrNotFound
	}

	return nil
}

func (r *Repository) List(ctx context.Context) ([]taskdomain.Task, error) {
	rows, err := r.pool.Query(ctx, ListQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]taskdomain.Task, 0)
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, err
		}

		tasks = append(tasks, *task)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

type taskScanner interface {
	Scan(dest ...any) error
}

func scanTask(scanner taskScanner) (*taskdomain.Task, error) {
	var (
		task               taskdomain.Task
		status             string
		repeatedType       string
		repeatedEveryNDays *int32
		repeatedDayOfMonth *int32
		specificDates      []time.Time
	)

	if err := scanner.Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&status,
		&repeatedType,
		&repeatedEveryNDays,
		&repeatedDayOfMonth,
		&specificDates,
		&task.CreatedAt,
		&task.UpdatedAt,
	); err != nil {
		return nil, err
	}

	task.Status = taskdomain.Status(status)
	task.Repeated = taskdomain.Repeated{
		Type:          taskdomain.PeriodType(repeatedType),
		SpecificDates: specificDates,
	}

	if task.Repeated.Type == "" {
		task.Repeated.Type = taskdomain.PeriodNone
	}
	if repeatedEveryNDays != nil {
		task.Repeated.EveryNDays = int(*repeatedEveryNDays)
	}
	if repeatedDayOfMonth != nil {
		task.Repeated.DayOfMonth = int(*repeatedDayOfMonth)
	}

	return &task, nil
}

func nullInt(v int) any {
	if v == 0 {
		return nil
	}

	return v
}

func nullTimes(v []time.Time) any {
	if len(v) == 0 {
		return nil
	}

	return v
}
