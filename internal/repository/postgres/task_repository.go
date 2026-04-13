package postgres

import (
	"context"
	"errors"
	"time"
	
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	const query = `
		INSERT INTO tasks (title, description, status, execution_date, rule_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, title, description, status, execution_date, rule_id, created_at, updated_at
	`

	row := r.pool.QueryRow(ctx, query, task.Title, task.Description, task.Status, task.ExecutionDate, task.RuleID, task.CreatedAt, task.UpdatedAt)
	created, err := scanTask(row)
	if err != nil {
        var pgErr *pgconn.PgError
        if errors.As(err, &pgErr) && pgErr.Code == "23505" {
            return nil, taskdomain.ErrAlreadyExists 
        }
        return nil, err
    }
    return created, nil
}

func (r *Repository) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	const query = `
		SELECT id, title, description, status, execution_date, rule_id, created_at, updated_at
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
			execution_date = $4,
			updated_at = $5
		WHERE id = $6
		RETURNING id, title, description, status, execution_date, rule_id, created_at, updated_at
	`

	row := r.pool.QueryRow(ctx, query, task.Title, task.Description, task.Status, task.ExecutionDate, task.UpdatedAt, task.ID)
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

func (r *Repository) List(ctx context.Context, from, to time.Time) ([]taskdomain.Task, error) {
	const query = `
		SELECT id, title, description, status, execution_date, rule_id, created_at, updated_at
		FROM tasks
		WHERE execution_date >= $1 AND execution_date <= $2
		ORDER BY execution_date ASC
	`

	rows, err := r.pool.Query(ctx, query, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	//tasks := make([]taskdomain.Task, 0)
	var tasks []taskdomain.Task
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

func (r *Repository) CreateRule(ctx context.Context, rule *taskdomain.TaskRule) (*taskdomain.TaskRule, error) {
	const query = `
		INSERT INTO task_rules (title, description, recurrence_type, recurrence_value, start_date, end_date, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, title, description, recurrence_type, recurrence_value, start_date, end_date, created_at
	`
	err := r.pool.QueryRow(ctx, query, rule.Title, rule.Description, rule.RecurrenceType, rule.RecurrenceValue, rule.StartDate, rule.EndDate, rule.CreatedAt).
		Scan(&rule.ID, &rule.Title, &rule.Description, &rule.RecurrenceType, &rule.RecurrenceValue, &rule.StartDate, &rule.EndDate, &rule.CreatedAt)
	return rule, err
}

func (r *Repository) ListActiveRules(ctx context.Context, from, to time.Time) ([]taskdomain.TaskRule, error) {
	const query = `
		SELECT id, title, description, recurrence_type, recurrence_value, start_date, end_date, created_at
		FROM task_rules
		WHERE start_date <= $2 AND (end_date IS NULL OR end_date >= $1)
	`
	rows, err := r.pool.Query(ctx, query, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []taskdomain.TaskRule
	for rows.Next() {
		var rule taskdomain.TaskRule
		if err := rows.Scan(&rule.ID, &rule.Title, &rule.Description, &rule.RecurrenceType, &rule.RecurrenceValue, &rule.StartDate, &rule.EndDate, &rule.CreatedAt); err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}
	return rules, rows.Err()
}

func (r *Repository) GetRuleByID(ctx context.Context, id int64) (*taskdomain.TaskRule, error) {
	const query = `
		SELECT id, title, description, recurrence_type, recurrence_value, start_date, end_date, created_at
		FROM task_rules
		WHERE id = $1
	`
	var rule taskdomain.TaskRule
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&rule.ID, &rule.Title, &rule.Description, &rule.RecurrenceType,
		&rule.RecurrenceValue, &rule.StartDate, &rule.EndDate, &rule.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskdomain.ErrNotFound
		}
		return nil, err
	}
	return &rule, nil
}

func (r *Repository) ListRules(ctx context.Context) ([]taskdomain.TaskRule, error) {
	const query = `
		SELECT id, title, description, recurrence_type, recurrence_value, start_date, end_date, created_at
		FROM task_rules
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []taskdomain.TaskRule
	for rows.Next() {
		var rule taskdomain.TaskRule
		if err := rows.Scan(&rule.ID, &rule.Title, &rule.Description, &rule.RecurrenceType,
			&rule.RecurrenceValue, &rule.StartDate, &rule.EndDate, &rule.CreatedAt); err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}
	return rules, rows.Err()
}

func (r *Repository) DeleteRule(ctx context.Context, id int64) error {
	const query = `DELETE FROM task_rules WHERE id = $1`
	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return taskdomain.ErrNotFound
	}
	return nil
}

type taskScanner interface {
	Scan(dest ...any) error
}

func scanTask(scanner taskScanner) (*taskdomain.Task, error) {
	var task taskdomain.Task
	var status string

	if err := scanner.Scan(
		&task.ID, &task.Title, &task.Description, &status, 
		&task.ExecutionDate, &task.RuleID, &task.CreatedAt, &task.UpdatedAt,
	); err != nil {
		return nil, err
	}
	task.Status = taskdomain.Status(status)
	return &task, nil
}
