package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	taskdomain "example.com/taskservice/internal/domain/task"
)

// RecurrenceRepository обеспечивает доступ к таблице recurrence_rules.
type RecurrenceRepository struct {
	pool *pgxpool.Pool
}

func NewRecurrenceRepository(pool *pgxpool.Pool) *RecurrenceRepository {
	return &RecurrenceRepository{pool: pool}
}

func (r *RecurrenceRepository) Create(ctx context.Context, rule *taskdomain.RecurrenceRule) (*taskdomain.RecurrenceRule, error) {
	const query = `
		INSERT INTO recurrence_rules (task_id, type, interval_days, day_of_month, specific_dates, even_odd, start_date, end_date, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, task_id, type, interval_days, day_of_month, specific_dates, even_odd, start_date, end_date, created_at
	`

	var specificDates []time.Time
	if len(rule.SpecificDates) > 0 {
		specificDates = rule.SpecificDates
	}

	var evenOdd *string
	if rule.EvenOdd != nil {
		s := string(*rule.EvenOdd)
		evenOdd = &s
	}

	row := r.pool.QueryRow(ctx, query,
		rule.TaskID,
		string(rule.Type),
		rule.IntervalDays,
		rule.DayOfMonth,
		specificDates,
		evenOdd,
		rule.StartDate,
		rule.EndDate,
		rule.CreatedAt,
	)

	created, err := scanRecurrenceRule(row)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (r *RecurrenceRepository) GetByTaskID(ctx context.Context, taskID int64) (*taskdomain.RecurrenceRule, error) {
	const query = `
		SELECT id, task_id, type, interval_days, day_of_month, specific_dates, even_odd, start_date, end_date, created_at
		FROM recurrence_rules
		WHERE task_id = $1
	`

	row := r.pool.QueryRow(ctx, query, taskID)
	rule, err := scanRecurrenceRule(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskdomain.ErrNotFound
		}

		return nil, err
	}

	return rule, nil
}

func (r *RecurrenceRepository) Update(ctx context.Context, rule *taskdomain.RecurrenceRule) (*taskdomain.RecurrenceRule, error) {
	const query = `
		UPDATE recurrence_rules
		SET type = $1,
			interval_days = $2,
			day_of_month = $3,
			specific_dates = $4,
			even_odd = $5,
			start_date = $6,
			end_date = $7
		WHERE task_id = $8
		RETURNING id, task_id, type, interval_days, day_of_month, specific_dates, even_odd, start_date, end_date, created_at
	`

	var evenOdd *string
	if rule.EvenOdd != nil {
		s := string(*rule.EvenOdd)
		evenOdd = &s
	}

	row := r.pool.QueryRow(ctx, query,
		string(rule.Type),
		rule.IntervalDays,
		rule.DayOfMonth,
		rule.SpecificDates,
		evenOdd,
		rule.StartDate,
		rule.EndDate,
		rule.TaskID,
	)

	updated, err := scanRecurrenceRule(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskdomain.ErrNotFound
		}

		return nil, err
	}

	return updated, nil
}

func (r *RecurrenceRepository) Delete(ctx context.Context, taskID int64) error {
	const query = `DELETE FROM recurrence_rules WHERE task_id = $1`

	result, err := r.pool.Exec(ctx, query, taskID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return taskdomain.ErrNotFound
	}

	return nil
}

// ListActive возвращает все правила, действующие на указанную дату.
func (r *RecurrenceRepository) ListActive(ctx context.Context, date time.Time) ([]taskdomain.RecurrenceRule, error) {
	const query = `
		SELECT id, task_id, type, interval_days, day_of_month, specific_dates, even_odd, start_date, end_date, created_at
		FROM recurrence_rules
		WHERE start_date <= $1
		  AND (end_date IS NULL OR end_date >= $1)
	`

	rows, err := r.pool.Query(ctx, query, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rules := make([]taskdomain.RecurrenceRule, 0)
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

type recurrenceScanner interface {
	Scan(dest ...any) error
}

func scanRecurrenceRule(scanner recurrenceScanner) (*taskdomain.RecurrenceRule, error) {
	var (
		rule          taskdomain.RecurrenceRule
		recType       string
		specificDates []time.Time
		evenOdd       *string
	)

	if err := scanner.Scan(
		&rule.ID,
		&rule.TaskID,
		&recType,
		&rule.IntervalDays,
		&rule.DayOfMonth,
		&specificDates,
		&evenOdd,
		&rule.StartDate,
		&rule.EndDate,
		&rule.CreatedAt,
	); err != nil {
		return nil, err
	}

	rule.Type = taskdomain.RecurrenceType(recType)
	rule.SpecificDates = specificDates

	if evenOdd != nil {
		eo := taskdomain.EvenOdd(*evenOdd)
		rule.EvenOdd = &eo
	}

	return &rule, nil
}
