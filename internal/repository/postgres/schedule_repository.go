package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	scheduledomain "example.com/taskservice/internal/domain/schedule"
)

type ScheduleRepository struct {
	pool *pgxpool.Pool
}

func NewScheduleRepository(pool *pgxpool.Pool) *ScheduleRepository {
	return &ScheduleRepository{
		pool: pool,
	}
}

func (r *ScheduleRepository) Create(ctx context.Context, schedule *scheduledomain.Schedule) (*scheduledomain.Schedule, error) {
	const query = `
		INSERT INTO schedules (title, description, recurrence_type, recurrence_params, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, title, description, recurrence_type, recurrence_params, is_active, last_created_date, created_at, updated_at
		
	`

	row := r.pool.QueryRow(ctx, query, schedule.Title, schedule.Description, schedule.RecurrenceType, schedule.RecurrenceParams, schedule.IsActive, schedule.CreatedAt, schedule.UpdatedAt)
	created, err := scanSchedule(row)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (r *ScheduleRepository) GetByID(ctx context.Context, id int64) (*scheduledomain.Schedule, error) {
	const query = `
		SELECT id, title, description, recurrence_type, recurrence_params, is_active, last_created_date, created_at, updated_at
		FROM schedules
		WHERE id = $1
`

	row := r.pool.QueryRow(ctx, query, id)
	found, err := scanSchedule(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, scheduledomain.ErrNotFound
		}

		return nil, err
	}

	return found, nil
}

func (r *ScheduleRepository) Update(ctx context.Context, schedule *scheduledomain.Schedule) (*scheduledomain.Schedule, error) {
	const query = `
		UPDATE schedules
		SET title = $1,
		    description = $2,
		    recurrence_type = $3,
		    recurrence_params = $4,
		    is_active = $5,
		    updated_at = $6
		WHERE id = $7
		RETURNING id, title, description, recurrence_type, recurrence_params, is_active, last_created_date, created_at, updated_at

`

	row := r.pool.QueryRow(ctx, query, schedule.Title, schedule.Description, schedule.RecurrenceType, schedule.RecurrenceParams, schedule.IsActive, schedule.UpdatedAt, schedule.ID)
	updated, err := scanSchedule(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, scheduledomain.ErrNotFound
		}

		return nil, err
	}

	return updated, nil
}

func (r *ScheduleRepository) Delete(ctx context.Context, id int64) error {
	const query = `DELETE FROM schedules WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return scheduledomain.ErrNotFound
	}

	return nil
}

func (r *ScheduleRepository) ListActive(ctx context.Context) ([]scheduledomain.Schedule, error) {
	const query = `
		SELECT id, title, description, recurrence_type, recurrence_params, is_active, last_created_date, created_at, updated_at
		FROM schedules
		WHERE is_active = TRUE
		ORDER BY id ASC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	schedules := make([]scheduledomain.Schedule, 0)
	for rows.Next() {
		schedule, err := scanSchedule(rows)
		if err != nil {
			return nil, err
		}

		schedules = append(schedules, *schedule)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return schedules, nil
}

func (r *ScheduleRepository) UpdateLastCreatedDate(ctx context.Context, id int64, date time.Time) error {
	const query = `
		UPDATE schedules
		SET last_created_date = $1
		WHERE id = $2
	`

	_, err := r.pool.Exec(ctx, query, date, id)
	return err
}

func (r *ScheduleRepository) List(ctx context.Context) ([]scheduledomain.Schedule, error) {
	const query = `
		SELECT id, title, description, recurrence_type, recurrence_params, is_active, last_created_date, created_at, updated_at
		FROM schedules
		ORDER BY id DESC
`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	schedules := make([]scheduledomain.Schedule, 0)
	for rows.Next() {
		schedule, err := scanSchedule(rows)
		if err != nil {
			return nil, err
		}

		schedules = append(schedules, *schedule)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return schedules, nil
}

type scheduleScanner interface {
	Scan(dest ...any) error
}

func scanSchedule(scanner scheduleScanner) (*scheduledomain.Schedule, error) {
	var (
		schedule scheduledomain.Schedule
	)

	if err := scanner.Scan(
		&schedule.ID,
		&schedule.Title,
		&schedule.Description,
		&schedule.RecurrenceType,
		&schedule.RecurrenceParams,
		&schedule.IsActive,
		&schedule.LastCreatedDate,
		&schedule.CreatedAt,
		&schedule.UpdatedAt,
	); err != nil {
		return nil, err
	}

	return &schedule, nil
}
