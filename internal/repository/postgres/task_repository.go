package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"time"

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
	var periodConfJSON []byte
	if task.PeriodConf != nil {
		periodConfJSON, _ = json.Marshal(task.PeriodConf)
	}

	const query = `
		INSERT INTO tasks (title, description, status, period_conf, scheduled_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, title, description, status, period_conf, scheduled_date, created_at, updated_at
	`

	row := r.pool.QueryRow(ctx, query,
		task.Title,
		task.Description,
		task.Status,
		periodConfJSON,
		task.ScheduledDate,
		task.CreatedAt,
		task.UpdatedAt,
	)

	return scanTask(row)
}

func (r *Repository) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	const query = `
		SELECT id, title, description, status, period_conf, scheduled_date, created_at, updated_at
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

func (r *Repository) Update(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	var periodConfJSON []byte
	if task.PeriodConf != nil {
		periodConfJSON, _ = json.Marshal(task.PeriodConf)
	}

	const query = `
		UPDATE tasks
		SET title = $1,
		    description = $2,
		    status = $3,
		    period_conf = $4,
		    scheduled_date = $5,
		    updated_at = $6
		WHERE id = $7
		RETURNING id, title, description, status, period_conf, scheduled_date, created_at, updated_at
	`

	row := r.pool.QueryRow(ctx, query,
		task.Title,
		task.Description,
		task.Status,
		periodConfJSON,
		task.ScheduledDate,
		task.UpdatedAt,
		task.ID,
	)

	return scanTask(row)
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
		SELECT id, title, description, status, period_conf, scheduled_date, created_at, updated_at
		FROM tasks
		ORDER BY id DESC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []taskdomain.Task
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, *task)
	}
	return tasks, rows.Err()
}

// CreateBatch создаёт несколько задач одним батчем
func (r *Repository) CreateBatch(ctx context.Context, tasks []taskdomain.Task) error {
	if len(tasks) == 0 {
		return nil
	}
	batch := &pgx.Batch{}
	for _, t := range tasks {
		var periodConfJSON []byte
		if t.PeriodConf != nil {
			periodConfJSON, _ = json.Marshal(t.PeriodConf)
		}
		query := `
			INSERT INTO tasks (title, description, status, period_conf, scheduled_date, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`
		batch.Queue(query,
			t.Title,
			t.Description,
			t.Status,
			periodConfJSON,
			t.ScheduledDate,
			t.CreatedAt,
			t.UpdatedAt,
		)
	}
	br := r.pool.SendBatch(ctx, batch)
	defer br.Close()
	for range tasks {
		if _, err := br.Exec(); err != nil {
			return err
		}
	}
	return nil
}


func (r *Repository) DeleteScheduledTasks(ctx context.Context, scheduledDate time.Time) error {
	const query = `DELETE FROM tasks WHERE scheduled_date = $1`
	_, err := r.pool.Exec(ctx, query, scheduledDate)
	return err
}

func scanTask(scanner interface {
	Scan(dest ...any) error
}) (*taskdomain.Task, error) {
	var (
		t           taskdomain.Task
		status      string
		rawPeriod   []byte
		schedDate   *time.Time
	)

	err := scanner.Scan(
		&t.ID,
		&t.Title,
		&t.Description,
		&status,
		&rawPeriod,
		&schedDate,
		&t.CreatedAt,
		&t.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	t.Status = taskdomain.Status(status)
	if schedDate != nil {
		t.ScheduledDate = schedDate
	}
	if len(rawPeriod) > 0 {
		var pc taskdomain.PeriodConf
		if err := json.Unmarshal(rawPeriod, &pc); err == nil {
			t.PeriodConf = &pc
		}
	}
	return &t, nil
}
