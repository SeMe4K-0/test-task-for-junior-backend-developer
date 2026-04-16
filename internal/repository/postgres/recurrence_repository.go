package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	recurrencedomain "example.com/taskservice/internal/domain/recurrence"
	taskdomain "example.com/taskservice/internal/domain/task"
)

const ruleColumns = `id, type, every_n_days, day_of_month, specific_dates, even_odd,
	start_date, end_date, task_title, task_description, created_at, updated_at`

type RecurrenceRepository struct {
	pool     *pgxpool.Pool
	taskRepo *Repository
}

func NewRecurrenceRepository(pool *pgxpool.Pool, taskRepo *Repository) *RecurrenceRepository {
	return &RecurrenceRepository{pool: pool, taskRepo: taskRepo}
}

func (r *RecurrenceRepository) CreateWithTasks(
	ctx context.Context,
	rule *recurrencedomain.RecurrenceRule,
	tasks []taskdomain.Task,
) (*recurrencedomain.RecurrenceRule, []taskdomain.Task, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, nil, err
	}
	defer tx.Rollback(ctx)

	insertRule := `
		INSERT INTO recurrence_rules (type, every_n_days, day_of_month, specific_dates, even_odd,
			start_date, end_date, task_title, task_description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING ` + ruleColumns

	var specificDates []time.Time
	if len(rule.SpecificDates) > 0 {
		specificDates = rule.SpecificDates
	}

	row := tx.QueryRow(ctx, insertRule,
		rule.Type, rule.EveryNDays, rule.DayOfMonth, specificDates,
		nullableString(string(rule.EvenOdd)),
		rule.StartDate, rule.EndDate,
		rule.TaskTitle, rule.TaskDescription,
		rule.CreatedAt, rule.UpdatedAt,
	)

	created, err := scanRecurrenceRule(row)
	if err != nil {
		return nil, nil, err
	}

	for i := range tasks {
		tasks[i].RecurrenceRuleID = &created.ID
	}

	createdTasks, err := r.taskRepo.CreateBatch(ctx, tx, tasks)
	if err != nil {
		return nil, nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, nil, err
	}

	return created, createdTasks, nil
}

func (r *RecurrenceRepository) GetByID(ctx context.Context, id int64) (*recurrencedomain.RecurrenceRule, error) {
	query := `SELECT ` + ruleColumns + ` FROM recurrence_rules WHERE id = $1 AND deleted_at IS NULL`

	row := r.pool.QueryRow(ctx, query, id)
	found, err := scanRecurrenceRule(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, recurrencedomain.ErrNotFound
		}
		return nil, err
	}
	return found, nil
}

func (r *RecurrenceRepository) List(ctx context.Context) ([]recurrencedomain.RecurrenceRule, error) {
	query := `SELECT ` + ruleColumns + ` FROM recurrence_rules WHERE deleted_at IS NULL ORDER BY id DESC`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rules := make([]recurrencedomain.RecurrenceRule, 0)
	for rows.Next() {
		rule, err := scanRecurrenceRule(rows)
		if err != nil {
			return nil, err
		}
		rules = append(rules, *rule)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return rules, nil
}

func (r *RecurrenceRepository) UpdateWithTasks(
	ctx context.Context,
	rule *recurrencedomain.RecurrenceRule,
	tasks []taskdomain.Task,
) (*recurrencedomain.RecurrenceRule, []taskdomain.Task, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, nil, err
	}
	defer tx.Rollback(ctx)

	var specificDates []time.Time
	if len(rule.SpecificDates) > 0 {
		specificDates = rule.SpecificDates
	}

	updateQuery := `
		UPDATE recurrence_rules
		SET type = $1, every_n_days = $2, day_of_month = $3, specific_dates = $4, even_odd = $5,
			start_date = $6, end_date = $7, task_title = $8, task_description = $9, updated_at = $10
		WHERE id = $11 AND deleted_at IS NULL
		RETURNING ` + ruleColumns

	row := tx.QueryRow(ctx, updateQuery,
		rule.Type, rule.EveryNDays, rule.DayOfMonth, specificDates,
		nullableString(string(rule.EvenOdd)),
		rule.StartDate, rule.EndDate,
		rule.TaskTitle, rule.TaskDescription,
		rule.UpdatedAt, rule.ID,
	)

	updated, err := scanRecurrenceRule(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, recurrencedomain.ErrNotFound
		}
		return nil, nil, err
	}

	if err := r.taskRepo.HardDeleteByRecurrenceRuleID(ctx, tx, rule.ID); err != nil {
		return nil, nil, err
	}

	for i := range tasks {
		tasks[i].RecurrenceRuleID = &updated.ID
	}

	createdTasks, err := r.taskRepo.CreateBatch(ctx, tx, tasks)
	if err != nil {
		return nil, nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, nil, err
	}

	return updated, createdTasks, nil
}

func (r *RecurrenceRepository) Delete(ctx context.Context, id int64, deleteTasks bool, now time.Time) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if deleteTasks {
		if err := r.taskRepo.SoftDeleteByRecurrenceRuleID(ctx, tx, id, now); err != nil {
			return err
		}
	}

	const query = `UPDATE recurrence_rules SET deleted_at = $1, updated_at = $1 WHERE id = $2 AND deleted_at IS NULL`
	result, err := tx.Exec(ctx, query, now, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return recurrencedomain.ErrNotFound
	}

	return tx.Commit(ctx)
}

func scanRecurrenceRule(scanner taskScanner) (*recurrencedomain.RecurrenceRule, error) {
	var (
		rule          recurrencedomain.RecurrenceRule
		ruleType      string
		evenOdd       *string
		specificDates []time.Time
	)

	if err := scanner.Scan(
		&rule.ID,
		&ruleType,
		&rule.EveryNDays,
		&rule.DayOfMonth,
		&specificDates,
		&evenOdd,
		&rule.StartDate,
		&rule.EndDate,
		&rule.TaskTitle,
		&rule.TaskDescription,
		&rule.CreatedAt,
		&rule.UpdatedAt,
	); err != nil {
		return nil, err
	}

	rule.Type = recurrencedomain.RecurrenceType(ruleType)
	if evenOdd != nil {
		rule.EvenOdd = recurrencedomain.EvenOddType(*evenOdd)
	}
	rule.SpecificDates = specificDates

	return &rule, nil
}

func nullableString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
