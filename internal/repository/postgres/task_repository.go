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

func (r *Repository) Create(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	const query = `
		INSERT INTO tasks (title, description, status, created_at, updated_at, parent_task_id, scheduled_date, is_template)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, title, description, status, created_at, updated_at, parent_task_id, scheduled_date, is_template
	`

	row := r.pool.QueryRow(ctx, query,
		task.Title, task.Description, task.Status,
		task.CreatedAt, task.UpdatedAt,
		task.ParentTaskID, task.ScheduledDate, task.IsTemplate,
	)

	created, err := scanTask(row)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (r *Repository) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	const query = `
		SELECT id, title, description, status, created_at, updated_at, parent_task_id, scheduled_date, is_template
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
		RETURNING id, title, description, status, created_at, updated_at, parent_task_id, scheduled_date, is_template
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

// List возвращает все задачи, исключая шаблоны (is_template = false).
func (r *Repository) List(ctx context.Context) ([]taskdomain.Task, error) {
	const query = `
		SELECT id, title, description, status, created_at, updated_at, parent_task_id, scheduled_date, is_template
		FROM tasks
		WHERE is_template = FALSE
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

// ListByScheduledDate возвращает задачи, запланированные на конкретную дату.
func (r *Repository) ListByScheduledDate(ctx context.Context, date string) ([]taskdomain.Task, error) {
	const query = `
		SELECT id, title, description, status, created_at, updated_at, parent_task_id, scheduled_date, is_template
		FROM tasks
		WHERE scheduled_date = $1 AND is_template = FALSE
		ORDER BY id DESC
	`

	rows, err := r.pool.Query(ctx, query, date)
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

// ExistsByParentAndDate проверяет, существует ли экземпляр задачи
// для данного шаблона на указанную дату (для предотвращения дубликатов).
func (r *Repository) ExistsByParentAndDate(ctx context.Context, parentID int64, date string) (bool, error) {
	const query = `
		SELECT EXISTS(
			SELECT 1 FROM tasks
			WHERE parent_task_id = $1 AND scheduled_date = $2
		)
	`

	var exists bool
	if err := r.pool.QueryRow(ctx, query, parentID, date).Scan(&exists); err != nil {
		return false, err
	}

	return exists, nil
}

// DeleteFutureInstancesByParent удаляет будущие экземпляры (status='new') для данного шаблона.
func (r *Repository) DeleteFutureInstancesByParent(ctx context.Context, parentID int64, fromDate string) error {
	const query = `
		DELETE FROM tasks
		WHERE parent_task_id = $1
		  AND scheduled_date >= $2
		  AND status = 'new'
	`

	_, err := r.pool.Exec(ctx, query, parentID, fromDate)
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
		&task.CreatedAt,
		&task.UpdatedAt,
		&task.ParentTaskID,
		&task.ScheduledDate,
		&task.IsTemplate,
	); err != nil {
		return nil, err
	}

	task.Status = taskdomain.Status(status)

	return &task, nil
}
