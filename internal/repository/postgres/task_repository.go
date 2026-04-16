package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	taskdomain "example.com/taskservice/internal/domain/task"
	taskusecase "example.com/taskservice/internal/usecase/task"
)

type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

const taskColumns = `id, title, description, status, recurrence_rule_id, scheduled_date, created_at, updated_at`

func (r *Repository) Create(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	query := `
		INSERT INTO tasks (title, description, status, recurrence_rule_id, scheduled_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING ` + taskColumns

	row := r.pool.QueryRow(ctx, query,
		task.Title, task.Description, task.Status,
		task.RecurrenceRuleID, task.ScheduledDate,
		task.CreatedAt, task.UpdatedAt,
	)
	return scanTask(row)
}

func (r *Repository) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	query := `SELECT ` + taskColumns + ` FROM tasks WHERE id = $1 AND deleted_at IS NULL`

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
	query := `
		UPDATE tasks
		SET title = $1, description = $2, status = $3, updated_at = $4
		WHERE id = $5 AND deleted_at IS NULL
		RETURNING ` + taskColumns

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

func (r *Repository) Delete(ctx context.Context, id int64, now time.Time) error {
	const query = `UPDATE tasks SET deleted_at = $1, updated_at = $1 WHERE id = $2 AND deleted_at IS NULL`

	result, err := r.pool.Exec(ctx, query, now, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return taskdomain.ErrNotFound
	}
	return nil
}

func (r *Repository) Restore(ctx context.Context, id int64, now time.Time) (*taskdomain.Task, error) {
	query := `
		UPDATE tasks
		SET deleted_at = NULL, updated_at = $1
		WHERE id = $2 AND deleted_at IS NOT NULL
		RETURNING ` + taskColumns

	row := r.pool.QueryRow(ctx, query, now, id)
	restored, err := scanTask(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskdomain.ErrNotFound
		}
		return nil, err
	}

	return restored, nil
}

func (r *Repository) List(ctx context.Context, pagination taskusecase.Pagination) ([]taskdomain.Task, int, error) {
	countQuery := `SELECT COUNT(*) FROM tasks WHERE deleted_at IS NULL`
	var total int
	if err := r.pool.QueryRow(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `SELECT ` + taskColumns + ` FROM tasks WHERE deleted_at IS NULL ORDER BY id DESC`
	query += fmt.Sprintf(" LIMIT %d OFFSET %d", pagination.Limit, pagination.Offset())

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	tasks, err := collectTasks(rows)
	if err != nil {
		return nil, 0, err
	}
	return tasks, total, nil
}

func (r *Repository) ListWithFilter(ctx context.Context, filter taskusecase.ListFilter, pagination taskusecase.Pagination) ([]taskdomain.Task, int, error) {
	where := []string{"deleted_at IS NULL"}
	var args []any
	idx := 1

	if filter.ScheduledDateFrom != nil {
		where = append(where, fmt.Sprintf("scheduled_date >= $%d", idx))
		args = append(args, *filter.ScheduledDateFrom)
		idx++
	}
	if filter.ScheduledDateTo != nil {
		where = append(where, fmt.Sprintf("scheduled_date <= $%d", idx))
		args = append(args, *filter.ScheduledDateTo)
		idx++
	}

	whereClause := " WHERE " + strings.Join(where, " AND ")

	countQuery := "SELECT COUNT(*) FROM tasks" + whereClause
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := "SELECT " + taskColumns + " FROM tasks" + whereClause
	query += " ORDER BY COALESCE(scheduled_date, created_at::date), id"
	query += fmt.Sprintf(" LIMIT %d OFFSET %d", pagination.Limit, pagination.Offset())

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	tasks, err := collectTasks(rows)
	if err != nil {
		return nil, 0, err
	}
	return tasks, total, nil
}

func (r *Repository) CreateBatch(ctx context.Context, tx pgx.Tx, tasks []taskdomain.Task) ([]taskdomain.Task, error) {
	if len(tasks) == 0 {
		return nil, nil
	}

	var (
		placeholders []string
		args         []any
	)

	for i, t := range tasks {
		base := i * 7
		placeholders = append(placeholders, fmt.Sprintf(
			"($%d, $%d, $%d, $%d, $%d, $%d, $%d)",
			base+1, base+2, base+3, base+4, base+5, base+6, base+7,
		))
		args = append(args,
			t.Title, t.Description, t.Status,
			t.RecurrenceRuleID, t.ScheduledDate,
			t.CreatedAt, t.UpdatedAt,
		)
	}

	query := `
		INSERT INTO tasks (title, description, status, recurrence_rule_id, scheduled_date, created_at, updated_at)
		VALUES ` + strings.Join(placeholders, ", ") + `
		RETURNING ` + taskColumns

	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return collectTasks(rows)
}

func (r *Repository) SoftDeleteByRecurrenceRuleID(ctx context.Context, tx pgx.Tx, ruleID int64, now time.Time) error {
	const query = `UPDATE tasks SET deleted_at = $1, updated_at = $1 WHERE recurrence_rule_id = $2 AND deleted_at IS NULL`
	_, err := tx.Exec(ctx, query, now, ruleID)
	return err
}

func (r *Repository) HardDeleteByRecurrenceRuleID(ctx context.Context, tx pgx.Tx, ruleID int64) error {
	const query = `DELETE FROM tasks WHERE recurrence_rule_id = $1`
	_, err := tx.Exec(ctx, query, ruleID)
	return err
}

type taskScanner interface {
	Scan(dest ...any) error
}

func scanTask(scanner taskScanner) (*taskdomain.Task, error) {
	var (
		task   taskdomain.Task
		status string
	)

	if err := scanner.Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&status,
		&task.RecurrenceRuleID,
		&task.ScheduledDate,
		&task.CreatedAt,
		&task.UpdatedAt,
	); err != nil {
		return nil, err
	}

	task.Status = taskdomain.Status(status)

	return &task, nil
}

func collectTasks(rows pgx.Rows) ([]taskdomain.Task, error) {
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
