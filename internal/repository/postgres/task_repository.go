package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	recurrencedomain "example.com/taskservice/internal/domain/recurrence"
	taskdomain "example.com/taskservice/internal/domain/task"
)

type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) CreateRecurrence(ctx context.Context, recurrence *recurrencedomain.Recurrence) (*recurrencedomain.Recurrence, error) {
	const query = `
		INSERT INTO task_generation_rules (title, description, start_date, end_date, interval_days, month_days, specific_dates, even_odd, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id
	`
	err := r.pool.QueryRow(ctx, query, recurrence.Title, recurrence.Description, recurrence.StartDate, recurrence.EndDate, recurrence.IntervalDays, recurrence.MonthDays, recurrence.SpecificDates, recurrence.EvenOdd, recurrence.CreatedAt, recurrence.UpdatedAt).Scan(&recurrence.ID)
	if err != nil {
		return nil, err
	}

	return recurrence, nil
}

func (r *Repository) ListRecurrence(ctx context.Context) ([]recurrencedomain.Recurrence, error) {
	const query = `
		SELECT id, title, description, active, start_date, end_date, interval_days, specific_dates, month_days, even_odd, created_at, updated_at FROM task_generation_rules
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	recurrences := make([]recurrencedomain.Recurrence, 0)
	for rows.Next() {
		recurrence, err := scanRecurrence(rows)
		if err != nil {
			return nil, err
		}

		recurrences = append(recurrences, *recurrence)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return recurrences, nil
}

func (r *Repository) Create(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	const query = `
		INSERT INTO tasks (title, description, due_date, status, created_at, updated_at, rule_id, scheduled_for)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, title, description, status, created_at, updated_at, rule_id, due_date, scheduled_for
	`

	row := r.pool.QueryRow(ctx, query, task.Title, task.Description, task.DueDate, task.Status, task.CreatedAt, task.UpdatedAt, task.RuleID, task.ScheduledFor)
	created, err := scanTask(row)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (r *Repository) CreateIfNotExists(ctx context.Context, task *taskdomain.Task) error {
	const query = `
		INSERT INTO tasks (title, description, status, created_at, updated_at, rule_id, scheduled_for)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (rule_id, scheduled_for) WHERE rule_id IS NOT NULL AND scheduled_for IS NOT NULL
		DO NOTHING
	`
	_, err := r.pool.Exec(ctx, query,
		task.Title, task.Description, task.Status,
		task.CreatedAt, task.UpdatedAt, task.RuleID, task.ScheduledFor,
	)
	return err
}

func (r *Repository) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	const query = `
		SELECT id, title, description, status, created_at, updated_at, rule_id, due_date, scheduled_for
		FROM tasks
		WHERE id = $1
	`

	row := r.pool.QueryRow(ctx, query, id)
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
	const query = `
		UPDATE tasks
		SET title = $1,
			description = $2,
			status = $3,
			updated_at = $4
		WHERE id = $5
		RETURNING id, title, description, status, created_at, updated_at, rule_id, due_date, scheduled_for
	`

	row := r.pool.QueryRow(ctx, query, task.Title, task.Description, task.Status, task.UpdatedAt, task.ID)
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
	const query = `DELETE FROM tasks WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return taskdomain.ErrNotFound
	}

	return nil
}

func (r *Repository) List(ctx context.Context) ([]taskdomain.Task, error) {
	const query = `
		SELECT id, title, description, status, created_at, updated_at, rule_id, due_date, scheduled_for
		FROM tasks
		ORDER BY id DESC
	`

	rows, err := r.pool.Query(ctx, query)
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

type scanner interface {
	Scan(dest ...any) error
}

func scanTask(scanner scanner) (*taskdomain.Task, error) {
	var (
		task   taskdomain.Task
		status string
	)

	if err := scanner.Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&status,
		&task.CreatedAt,
		&task.UpdatedAt,
		&task.RuleID,
		&task.DueDate,
		&task.ScheduledFor,
	); err != nil {
		return nil, err
	}

	task.Status = taskdomain.Status(status)

	return &task, nil
}

func scanRecurrence(scanner scanner) (*recurrencedomain.Recurrence, error) {
	var (
		recurrence recurrencedomain.Recurrence
	)

	if err := scanner.Scan(
		&recurrence.ID,
		&recurrence.Title,
		&recurrence.Description,
		&recurrence.Active,
		&recurrence.StartDate,
		&recurrence.EndDate,
		&recurrence.IntervalDays,
		&recurrence.SpecificDates,
		&recurrence.MonthDays,
		&recurrence.EvenOdd,
		&recurrence.CreatedAt,
		&recurrence.UpdatedAt,
	); err != nil {
		return nil, err
	}

	return &recurrence, nil
}
