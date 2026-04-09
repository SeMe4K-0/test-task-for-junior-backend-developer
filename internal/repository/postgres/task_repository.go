package postgres

import (
	"context"
	"errors"
	"encoding/json"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	taskdomain "example.com/taskservice/internal/domain/task"
)
//в этом я добавил новый столбец в таблице для бд

type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	//здесь добавляем наш период в таблицу
	const query = `
		INSERT INTO tasks (title, description, status, period_conf, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, title, description, status, period_conf, created_at, updated_at
	`
	var periodConfJSON []byte
	if task.PeriodConf != nil {
		periodConfJSON, _ = json.Marshal(task.PeriodConf)
	}

	row := r.pool.QueryRow(ctx, query, task.Title, task.Description, task.Status, periodConfJSON, task.CreatedAt, task.UpdatedAt)
	created, err := scanTask(row)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (r *Repository) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	//и здесь
	const query = `
		SELECT id, title, description, status, period_conf, created_at, updated_at
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
	//и здесь тоже
	const query = `
		UPDATE tasks
		SET title = $1,
			description = $2,
			status = $3,
			period_conf = $4,
			updated_at = $5
		WHERE id = $6
		RETURNING id, title, description, status, period_conf, created_at, updated_at
	`
	var periodConfJSON []byte
	if task.PeriodConf != nil {
		periodConfJSON, _ = json.Marshal(task.PeriodConf)
	}

	row := r.pool.QueryRow(ctx, query, task.Title, task.Description, task.Status, periodConfJSON, task.UpdatedAt, task.ID)
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
	//здесь также
	const query = `
		SELECT id, title, description, status, period_conf, created_at, updated_at
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

type taskScanner interface {
	Scan(dest ...any) error
}

func scanTask(scanner taskScanner) (*taskdomain.Task, error) {
	var (
		task   taskdomain.Task
		status string
		rawPeriodConf []byte
	)

	if err := scanner.Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&status,
		&rawPeriodConf,
		&task.CreatedAt,
		&task.UpdatedAt,
	); err != nil {
		return nil, err
	}

	task.Status = taskdomain.Status(status)

	//делаем проверку на пустой период

	if rawPeriodConf != nil {
		var pc taskdomain.PeriodConf
		if err := json.Unmarshal(rawPeriodConf, &pc); err != nil {
			return nil, err
		}
		task.PeriodConf = &pc
	}

	return &task, nil
}
