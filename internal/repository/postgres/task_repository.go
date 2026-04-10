package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	taskdomain "example.com/taskservice/internal/domain/task"
	"example.com/taskservice/internal/infrastructure/logger"
	"example.com/taskservice/internal/infrastructure/metrics"
)

type Repository struct {
	pool   *pgxpool.Pool
	logger logger.Logger
}

func New(pool *pgxpool.Pool, logger logger.Logger) *Repository {
	return &Repository{
		pool:   pool,
		logger: logger,
	}
}

func (r *Repository) Create(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	start := time.Now()
	r.logger.Debug("Creating task in database",
		zap.String("title", task.Title),
		zap.String("status", string(task.Status)))

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		r.logger.Error("Failed to begin transaction",
			zap.Error(err))
		metrics.RecordError("repository", "create", "transaction")
		return nil, err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	const taskQuery = `
		INSERT INTO tasks (title, description, status, due_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, title, description, status, due_date, created_at, updated_at
	`

	row := tx.QueryRow(ctx, taskQuery, task.Title, task.Description, task.Status, task.DueDate, task.CreatedAt, task.UpdatedAt)
	created, err := scanTask(row)
	if err != nil {
		return nil, err
	}

	if task.Recurrence != nil {
		const recurrenceQuery = `
			INSERT INTO task_recurrence (
				task_id,
				recur_type,
				interval_days,
				month_day,
				specific_dates,
				parity,
				created_at,
				updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

		_, err = tx.Exec(ctx, recurrenceQuery,
			created.ID,
			task.Recurrence.Type,
			nullableInt(task.Recurrence.IntervalDays),
			nullableInt(task.Recurrence.MonthDay),
			task.Recurrence.SpecificDates,
			nullableParity(task.Recurrence.Parity),
			task.CreatedAt,
			task.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("Failed to create task recurrence",
				zap.Error(err),
				zap.Int64("task_id", created.ID))
			metrics.RecordError("repository", "create", "recurrence")
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		r.logger.Error("Failed to commit transaction",
			zap.Error(err))
		metrics.RecordError("repository", "create", "commit")
		return nil, err
	}

	duration := time.Since(start)
	r.logger.Info("Task created successfully",
		zap.Int64("task_id", created.ID),
		zap.String("title", created.Title),
		zap.Duration("duration", duration))
	metrics.RecordDatabaseQuery("create", "tasks", duration)

	return created, nil
}

func (r *Repository) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	start := time.Now()
	r.logger.Debug("Getting task by ID from database", zap.Int64("task_id", id))

	const query = `
		SELECT
			t.id,
			t.title,
			t.description,
			t.status,
			t.due_date,
			t.created_at,
			t.updated_at,
			r.recur_type,
			r.interval_days,
			r.month_day,
			r.specific_dates,
			r.parity
		FROM tasks t
		LEFT JOIN task_recurrence r ON r.task_id = t.id
		WHERE t.id = $1
	`

	row := r.pool.QueryRow(ctx, query, id)
	found, err := scanTaskWithRecurrence(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.logger.Warn("Task not found",
				zap.Int64("task_id", id))
			metrics.RecordError("repository", "get_by_id", "not_found")
			return nil, taskdomain.ErrNotFound
		}

		r.logger.Error("Failed to get task by ID",
			zap.Error(err),
			zap.Int64("task_id", id))
		metrics.RecordError("repository", "get_by_id", "query")
		return nil, err
	}

	duration := time.Since(start)
	r.logger.Debug("Task retrieved successfully",
		zap.Int64("task_id", found.ID),
		zap.String("title", found.Title),
		zap.Duration("duration", duration))
	metrics.RecordDatabaseQuery("select", "tasks", duration)

	return found, nil
}

func (r *Repository) Update(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	const taskQuery = `
		UPDATE tasks
		SET title = $1,
			description = $2,
			status = $3,
			due_date = $4,
			updated_at = $5
		WHERE id = $6
		RETURNING id, title, description, status, due_date, created_at, updated_at
	`

	row := tx.QueryRow(ctx, taskQuery, task.Title, task.Description, task.Status, task.DueDate, task.UpdatedAt, task.ID)
	updated, err := scanTask(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskdomain.ErrNotFound
		}

		return nil, err
	}

	if task.Recurrence == nil {
		const deleteRecurrence = `DELETE FROM task_recurrence WHERE task_id = $1`
		if _, err := tx.Exec(ctx, deleteRecurrence, task.ID); err != nil {
			return nil, err
		}
	} else {
		const recurrenceQuery = `
			INSERT INTO task_recurrence (
				task_id,
				recur_type,
				interval_days,
				month_day,
				specific_dates,
				parity,
				created_at,
				updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			ON CONFLICT (task_id) DO UPDATE SET
				recur_type = EXCLUDED.recur_type,
				interval_days = EXCLUDED.interval_days,
				month_day = EXCLUDED.month_day,
				specific_dates = EXCLUDED.specific_dates,
				parity = EXCLUDED.parity,
				updated_at = EXCLUDED.updated_at
		`

		_, err = tx.Exec(ctx, recurrenceQuery,
			task.ID,
			task.Recurrence.Type,
			nullableInt(task.Recurrence.IntervalDays),
			nullableInt(task.Recurrence.MonthDay),
			task.Recurrence.SpecificDates,
			nullableParity(task.Recurrence.Parity),
			updated.CreatedAt,
			updated.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
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
		SELECT
			t.id,
			t.title,
			t.description,
			t.status,
			t.due_date,
			t.created_at,
			t.updated_at,
			r.recur_type,
			r.interval_days,
			r.month_day,
			r.specific_dates,
			r.parity
		FROM tasks t
		LEFT JOIN task_recurrence r ON r.task_id = t.id
		ORDER BY t.id DESC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]taskdomain.Task, 0)
	for rows.Next() {
		task, err := scanTaskWithRecurrence(rows)
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
		task    taskdomain.Task
		status  string
		dueDate *time.Time
	)

	if err := scanner.Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&status,
		&dueDate,
		&task.CreatedAt,
		&task.UpdatedAt,
	); err != nil {
		return nil, err
	}

	task.Status = taskdomain.Status(status)
	task.DueDate = dueDate

	return &task, nil
}

func scanTaskWithRecurrence(scanner taskScanner) (*taskdomain.Task, error) {
	var (
		task          taskdomain.Task
		status        string
		dueDate       *time.Time
		recurType     *string
		intervalDays  *int
		monthDay      *int
		specificDates []string
		parity        *string
	)

	if err := scanner.Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&status,
		&dueDate,
		&task.CreatedAt,
		&task.UpdatedAt,
		&recurType,
		&intervalDays,
		&monthDay,
		&specificDates,
		&parity,
	); err != nil {
		return nil, err
	}

	task.Status = taskdomain.Status(status)
	task.DueDate = dueDate

	if recurType != nil {
		task.Recurrence = &taskdomain.Recurrence{
			Type:          taskdomain.RecurrenceType(*recurType),
			SpecificDates: specificDates,
		}
		if intervalDays != nil {
			task.Recurrence.IntervalDays = *intervalDays
		}
		if monthDay != nil {
			task.Recurrence.MonthDay = *monthDay
		}
		if parity != nil {
			task.Recurrence.Parity = taskdomain.Parity(*parity)
		}
	}

	return &task, nil
}

func nullableInt(value int) *int {
	if value == 0 {
		return nil
	}
	return &value
}

func nullableParity(value taskdomain.Parity) *string {
	if value == "" {
		return nil
	}
	s := string(value)
	return &s
}
