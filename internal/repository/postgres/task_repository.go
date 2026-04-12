package postgres

import (
	"context"
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
	ruleType, dailyInterval, monthlyDays, specificDates, weekdayParity := flattenRecurrence(task)

	const query = `
		INSERT INTO tasks (
			title, description, status,
			scheduled_at, parent_task_id,
			recurrence_type, recurrence_daily_interval, recurrence_monthly_days,
			recurrence_specific_dates, recurrence_day_parity,
			created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		RETURNING id, title, description, status,
		          scheduled_at, parent_task_id,
		          recurrence_type, recurrence_daily_interval, recurrence_monthly_days,
		          recurrence_specific_dates, recurrence_day_parity,
		          created_at, updated_at
	`
	row := r.pool.QueryRow(ctx, query,
		task.Title, task.Description, task.Status,
		task.ScheduledAt, task.ParentTaskID,
		ruleType, dailyInterval, monthlyDays,
		specificDates, weekdayParity,
		task.CreatedAt, task.UpdatedAt,
	)
	return scanTask(row)
}

func (r *Repository) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	const query = `
		SELECT id, title, description, status,
		       scheduled_at, parent_task_id,
		       recurrence_type, recurrence_daily_interval, recurrence_monthly_days,
		       recurrence_specific_dates, recurrence_day_parity,
		       created_at, updated_at
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
	ruleType, dailyInterval, monthlyDays, specificDates, weekdayParity := flattenRecurrence(task)

	const query = `
		UPDATE tasks
		SET title                      = $1,
		    description                = $2,
		    status                     = $3,
		    scheduled_at               = $4,
		    recurrence_type            = $5,
		    recurrence_daily_interval  = $6,
		    recurrence_monthly_days    = $7,
		    recurrence_specific_dates  = $8,
		    recurrence_day_parity  = $9,
		    updated_at                 = $10
		WHERE id = $11
		RETURNING id, title, description, status,
		          scheduled_at, parent_task_id,
		          recurrence_type, recurrence_daily_interval, recurrence_monthly_days,
		          recurrence_specific_dates, recurrence_day_parity,
		          created_at, updated_at
	`
	row := r.pool.QueryRow(ctx, query,
		task.Title, task.Description, task.Status,
		task.ScheduledAt,
		ruleType, dailyInterval, monthlyDays,
		specificDates, weekdayParity,
		task.UpdatedAt,
		task.ID,
	)
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

// List возвращает задачи. Если includeTemplates == false, шаблоны исключаются.
func (r *Repository) List(ctx context.Context, includeTemplates bool) ([]taskdomain.Task, error) {
	query := `
		SELECT id, title, description, status,
		       scheduled_at, parent_task_id,
		       recurrence_type, recurrence_daily_interval, recurrence_monthly_days,
		       recurrence_specific_dates, recurrence_day_parity,
		       created_at, updated_at
		FROM tasks
	`
	if !includeTemplates {
		query += ` WHERE NOT (recurrence_type IS NOT NULL AND parent_task_id IS NULL)`
	}
	query += ` ORDER BY id DESC`

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
	return tasks, rows.Err()
}

// ListTemplates возвращает только задачи-шаблоны (используется для проверки горизонта).
func (r *Repository) ListTemplates(ctx context.Context) ([]taskdomain.Task, error) {
	const query = `
		SELECT id, title, description, status,
		       scheduled_at, parent_task_id,
		       recurrence_type, recurrence_daily_interval, recurrence_monthly_days,
		       recurrence_specific_dates, recurrence_day_parity,
		       created_at, updated_at
		FROM tasks
		WHERE recurrence_type IS NOT NULL AND parent_task_id IS NULL
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
	return tasks, rows.Err()
}

// CreateBatch вставляет несколько задач за один round-trip; дубликаты молча пропускаются.
func (r *Repository) CreateBatch(ctx context.Context, tasks []taskdomain.Task) error {
	if len(tasks) == 0 {
		return nil
	}

	const query = `
		INSERT INTO tasks (
			title, description, status,
			scheduled_at, parent_task_id,
			recurrence_type, recurrence_daily_interval, recurrence_monthly_days,
			recurrence_specific_dates, recurrence_day_parity,
			created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		ON CONFLICT (parent_task_id, scheduled_at)
		WHERE parent_task_id IS NOT NULL AND scheduled_at IS NOT NULL
		DO NOTHING
	`
	batch := &pgx.Batch{}
	for _, t := range tasks {
		ruleType, dailyInterval, monthlyDays, specificDates, weekdayParity := flattenRecurrence(&t)
		batch.Queue(query,
			t.Title, t.Description, t.Status,
			t.ScheduledAt, t.ParentTaskID,
			ruleType, dailyInterval, monthlyDays,
			specificDates, weekdayParity,
			t.CreatedAt, t.UpdatedAt,
		)
	}

	results := r.pool.SendBatch(ctx, batch)
	defer results.Close()

	for i := 0; i < len(tasks); i++ {
		if _, err := results.Exec(); err != nil {
			return err
		}
	}
	return nil
}

// DeleteFutureByTemplate удаляет незапущенные экземпляры шаблона с scheduled_at >= from.
func (r *Repository) DeleteFutureByTemplate(ctx context.Context, templateID int64, from time.Time) error {
	const query = `
		DELETE FROM tasks
		WHERE parent_task_id = $1
		  AND scheduled_at >= $2
		  AND status = 'new'
	`
	_, err := r.pool.Exec(ctx, query, templateID, from)
	return err
}


type taskScanner interface {
	Scan(dest ...any) error
}

// flattenRecurrence разворачивает поля периодичности задачи в значения для колонок БД.
func flattenRecurrence(t *taskdomain.Task) (
	ruleType *string,
	dailyInterval *int32,
	monthlyDays []int32,
	specificDates []time.Time,
	weekdayParity *string,
) {
	if t.RecurrenceType == nil {
		return nil, nil, nil, nil, nil
	}

	rt := string(*t.RecurrenceType)
	ruleType = &rt

	if t.RecurrenceDailyInterval != nil {
		i := int32(*t.RecurrenceDailyInterval)
		dailyInterval = &i
	}

	if len(t.RecurrenceMonthlyDays) > 0 {
		monthlyDays = make([]int32, len(t.RecurrenceMonthlyDays))
		for i, d := range t.RecurrenceMonthlyDays {
			monthlyDays[i] = int32(d)
		}
	}

	specificDates = t.RecurrenceSpecificDates

	if t.RecurrenceDayParity != nil {
		p := string(*t.RecurrenceDayParity)
		weekdayParity = &p
	}

	return
}

func scanTask(scanner taskScanner) (*taskdomain.Task, error) {
	var (
		t              taskdomain.Task
		status         string
		ruleType       *string
		dailyInterval  *int32
		monthlyDays    []int32
		specificDates  []time.Time
		weekdayParity  *string
	)

	if err := scanner.Scan(
		&t.ID,
		&t.Title,
		&t.Description,
		&status,
		&t.ScheduledAt,
		&t.ParentTaskID,
		&ruleType,
		&dailyInterval,
		&monthlyDays,
		&specificDates,
		&weekdayParity,
		&t.CreatedAt,
		&t.UpdatedAt,
	); err != nil {
		return nil, err
	}

	t.Status = taskdomain.Status(status)

	if ruleType != nil {
		rt := taskdomain.RecurrenceType(*ruleType)
		t.RecurrenceType = &rt
	}
	if dailyInterval != nil {
		i := int(*dailyInterval)
		t.RecurrenceDailyInterval = &i
	}
	if len(monthlyDays) > 0 {
		t.RecurrenceMonthlyDays = make([]int, len(monthlyDays))
		for i, d := range monthlyDays {
			t.RecurrenceMonthlyDays[i] = int(d)
		}
	}
	t.RecurrenceSpecificDates = specificDates

	if weekdayParity != nil {
		p := taskdomain.Parity(*weekdayParity)
		t.RecurrenceDayParity = &p
	}

	return &t, nil
}
