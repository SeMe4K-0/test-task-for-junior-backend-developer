package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
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

func (r *Repository) Ping(ctx context.Context) error {
	return r.pool.Ping(ctx)
}

func (r *Repository) CreateTemplate(ctx context.Context, template *taskdomain.Template) (*taskdomain.Template, error) {
	const query = `
		INSERT INTO task_templates (title, description, schedule, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, title, description, schedule, created_at, updated_at
	`

	row := r.pool.QueryRow(ctx, query, template.Title, template.Description, encodeSchedule(template.Schedule), template.CreatedAt, template.UpdatedAt)
	return scanTemplate(row)
}

func (r *Repository) GetTemplateByID(ctx context.Context, id int64) (*taskdomain.Template, error) {
	const query = `
		SELECT id, title, description, schedule, created_at, updated_at
		FROM task_templates
		WHERE id = $1
	`

	row := r.pool.QueryRow(ctx, query, id)
	template, err := scanTemplate(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskdomain.ErrNotFound
		}

		return nil, err
	}

	return template, nil
}

func (r *Repository) UpdateTemplate(ctx context.Context, template *taskdomain.Template) (*taskdomain.Template, error) {
	const query = `
		UPDATE task_templates
		SET title = $1,
			description = $2,
			schedule = $3,
			updated_at = $4
		WHERE id = $5
		RETURNING id, title, description, schedule, created_at, updated_at
	`

	row := r.pool.QueryRow(ctx, query, template.Title, template.Description, encodeSchedule(template.Schedule), template.UpdatedAt, template.ID)
	updated, err := scanTemplate(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskdomain.ErrNotFound
		}

		return nil, err
	}

	return updated, nil
}

func (r *Repository) DeleteTemplate(ctx context.Context, id int64) error {
	const query = `DELETE FROM task_templates WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return taskdomain.ErrNotFound
	}

	return nil
}

func (r *Repository) ListTemplates(ctx context.Context, filter taskusecase.TemplateListInput) ([]taskdomain.Template, int, error) {
	whereSQL, args := buildTemplateWhereClause(filter)

	countQuery := `SELECT COUNT(*) FROM task_templates ` + whereSQL
	total, err := r.count(ctx, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, title, description, schedule, created_at, updated_at
		FROM task_templates
	` + whereSQL + `
		ORDER BY id DESC
		LIMIT $` + fmt.Sprint(len(args)+1) + ` OFFSET $` + fmt.Sprint(len(args)+2)

	rows, err := r.pool.Query(ctx, query, append(args, filter.Limit, filter.Offset)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	templates := make([]taskdomain.Template, 0)
	for rows.Next() {
		template, err := scanTemplate(rows)
		if err != nil {
			return nil, 0, err
		}

		templates = append(templates, *template)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return templates, total, nil
}

func (r *Repository) ListTemplatesForSync(ctx context.Context, templateID *int64) ([]taskdomain.Template, error) {
	query := `
		SELECT id, title, description, schedule, created_at, updated_at
		FROM task_templates
	`
	args := make([]any, 0, 1)
	if templateID != nil {
		query += ` WHERE id = $1`
		args = append(args, *templateID)
	}
	query += ` ORDER BY id DESC`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	templates := make([]taskdomain.Template, 0)
	for rows.Next() {
		template, err := scanTemplate(rows)
		if err != nil {
			return nil, err
		}

		templates = append(templates, *template)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return templates, nil
}

func (r *Repository) UpsertOccurrences(ctx context.Context, tasks []taskdomain.Task) error {
	if len(tasks) == 0 {
		return nil
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	const query = `
		INSERT INTO task_occurrences (
			template_id,
			title,
			description,
			status,
			scheduled_for,
			created_at,
			updated_at,
			completed_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (template_id, scheduled_for) DO NOTHING
	`

	for i := range tasks {
		_, err := tx.Exec(
			ctx,
			query,
			tasks[i].TemplateID,
			tasks[i].Title,
			tasks[i].Description,
			tasks[i].Status,
			tasks[i].ScheduledFor.String(),
			tasks[i].CreatedAt,
			tasks[i].UpdatedAt,
			tasks[i].CompletedAt,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *Repository) GetOccurrenceByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	const query = `
		SELECT
			o.id,
			o.template_id,
			o.title,
			o.description,
			o.status,
			o.scheduled_for,
			o.created_at,
			o.updated_at,
			o.completed_at,
			t.schedule
		FROM task_occurrences o
		INNER JOIN task_templates t ON t.id = o.template_id
		WHERE o.id = $1
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

func (r *Repository) GetOccurrenceByTemplateAndDate(ctx context.Context, templateID int64, scheduledFor taskdomain.Date) (*taskdomain.Task, error) {
	const query = `
		SELECT
			o.id,
			o.template_id,
			o.title,
			o.description,
			o.status,
			o.scheduled_for,
			o.created_at,
			o.updated_at,
			o.completed_at,
			t.schedule
		FROM task_occurrences o
		INNER JOIN task_templates t ON t.id = o.template_id
		WHERE o.template_id = $1 AND o.scheduled_for = $2::date
	`

	row := r.pool.QueryRow(ctx, query, templateID, scheduledFor.String())
	task, err := scanTask(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskdomain.ErrNotFound
		}

		return nil, err
	}

	return task, nil
}

func (r *Repository) UpdateOccurrence(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	const query = `
		UPDATE task_occurrences o
		SET title = $1,
			description = $2,
			status = $3,
			updated_at = $4,
			completed_at = $5
		WHERE o.id = $6
		RETURNING
			o.id,
			o.template_id,
			o.title,
			o.description,
			o.status,
			o.scheduled_for,
			o.created_at,
			o.updated_at,
			o.completed_at,
			(SELECT schedule FROM task_templates WHERE id = o.template_id)
	`

	row := r.pool.QueryRow(ctx, query, task.Title, task.Description, task.Status, task.UpdatedAt, task.CompletedAt, task.ID)
	updated, err := scanTask(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskdomain.ErrNotFound
		}

		return nil, err
	}

	return updated, nil
}

func (r *Repository) DeleteOccurrence(ctx context.Context, id int64) error {
	const query = `DELETE FROM task_occurrences WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return taskdomain.ErrNotFound
	}

	return nil
}

func (r *Repository) DeleteOccurrencesFrom(ctx context.Context, templateID int64, from taskdomain.Date) error {
	const query = `DELETE FROM task_occurrences WHERE template_id = $1 AND scheduled_for >= $2::date`
	_, err := r.pool.Exec(ctx, query, templateID, from.String())
	return err
}

func (r *Repository) CountOccurrencesByTemplate(ctx context.Context, templateID int64) (int, error) {
	return r.count(ctx, `SELECT COUNT(*) FROM task_occurrences WHERE template_id = $1`, templateID)
}

func (r *Repository) ListOccurrences(ctx context.Context, filter taskusecase.TaskListInput) ([]taskdomain.Task, int, error) {
	whereSQL, args := buildOccurrenceWhereClause(filter)

	countQuery := `
		SELECT COUNT(*)
		FROM task_occurrences o
		INNER JOIN task_templates t ON t.id = o.template_id
	` + whereSQL

	total, err := r.count(ctx, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	query := `
		SELECT
			o.id,
			o.template_id,
			o.title,
			o.description,
			o.status,
			o.scheduled_for,
			o.created_at,
			o.updated_at,
			o.completed_at,
			t.schedule
		FROM task_occurrences o
		INNER JOIN task_templates t ON t.id = o.template_id
	` + whereSQL + `
		ORDER BY o.scheduled_for DESC, o.id DESC
		LIMIT $` + fmt.Sprint(len(args)+1) + ` OFFSET $` + fmt.Sprint(len(args)+2)

	rows, err := r.pool.Query(ctx, query, append(args, filter.Limit, filter.Offset)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	tasks := make([]taskdomain.Task, 0)
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, 0, err
		}

		tasks = append(tasks, *task)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return tasks, total, nil
}

type templateScanner interface {
	Scan(dest ...any) error
}

type taskScanner interface {
	Scan(dest ...any) error
}

func scanTemplate(scanner templateScanner) (*taskdomain.Template, error) {
	var (
		template      taskdomain.Template
		scheduleBytes []byte
	)

	if err := scanner.Scan(
		&template.ID,
		&template.Title,
		&template.Description,
		&scheduleBytes,
		&template.CreatedAt,
		&template.UpdatedAt,
	); err != nil {
		return nil, err
	}

	if len(scheduleBytes) > 0 {
		var schedule taskdomain.Schedule
		if err := json.Unmarshal(scheduleBytes, &schedule); err != nil {
			return nil, err
		}
		template.Schedule = &schedule
	}

	return &template, nil
}

func scanTask(scanner taskScanner) (*taskdomain.Task, error) {
	var (
		task          taskdomain.Task
		status        string
		scheduledFor  pgtype.Date
		completedAt   pgtype.Timestamptz
		scheduleBytes []byte
	)

	if err := scanner.Scan(
		&task.ID,
		&task.TemplateID,
		&task.Title,
		&task.Description,
		&status,
		&scheduledFor,
		&task.CreatedAt,
		&task.UpdatedAt,
		&completedAt,
		&scheduleBytes,
	); err != nil {
		return nil, err
	}

	task.Status = taskdomain.Status(status)
	if scheduledFor.Valid {
		task.ScheduledFor = taskdomain.NewDate(scheduledFor.Time)
	}
	if completedAt.Valid {
		value := completedAt.Time
		task.CompletedAt = &value
	}
	if len(scheduleBytes) > 0 {
		var schedule taskdomain.Schedule
		if err := json.Unmarshal(scheduleBytes, &schedule); err != nil {
			return nil, err
		}
		task.Schedule = &schedule
	}

	return &task, nil
}

func buildTemplateWhereClause(filter taskusecase.TemplateListInput) (string, []any) {
	clauses := make([]string, 0, 2)
	args := make([]any, 0, 2)

	if filter.Search != "" {
		args = append(args, "%"+filter.Search+"%")
		position := len(args)
		clauses = append(clauses, fmt.Sprintf("(title ILIKE $%d OR description ILIKE $%d)", position, position))
	}

	if filter.ScheduleType != nil {
		args = append(args, string(*filter.ScheduleType))
		clauses = append(clauses, fmt.Sprintf("schedule ->> 'type' = $%d", len(args)))
	}

	return buildWhereClause(clauses), args
}

func buildOccurrenceWhereClause(filter taskusecase.TaskListInput) (string, []any) {
	clauses := make([]string, 0, 4)
	args := make([]any, 0, 4)

	if filter.Status != nil {
		args = append(args, string(*filter.Status))
		clauses = append(clauses, fmt.Sprintf("o.status = $%d", len(args)))
	}
	if filter.TemplateID != nil {
		args = append(args, *filter.TemplateID)
		clauses = append(clauses, fmt.Sprintf("o.template_id = $%d", len(args)))
	}
	if filter.DateFrom != nil {
		args = append(args, filter.DateFrom.String())
		clauses = append(clauses, fmt.Sprintf("o.scheduled_for >= $%d::date", len(args)))
	}
	if filter.DateTo != nil {
		args = append(args, filter.DateTo.String())
		clauses = append(clauses, fmt.Sprintf("o.scheduled_for <= $%d::date", len(args)))
	}

	return buildWhereClause(clauses), args
}

func buildWhereClause(clauses []string) string {
	if len(clauses) == 0 {
		return ""
	}

	return " WHERE " + strings.Join(clauses, " AND ")
}

func (r *Repository) count(ctx context.Context, query string, args ...any) (int, error) {
	var total int
	if err := r.pool.QueryRow(ctx, query, args...).Scan(&total); err != nil {
		return 0, err
	}

	return total, nil
}

func encodeSchedule(schedule *taskdomain.Schedule) []byte {
	if schedule == nil {
		return nil
	}

	data, err := json.Marshal(schedule)
	if err != nil {
		return nil
	}

	return data
}
