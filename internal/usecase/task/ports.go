package task

import (
	"context"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type Repository interface {
	Create(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error)
	GetByID(ctx context.Context, id int64) (*taskdomain.Task, error)
	Update(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, includeTemplates bool) ([]taskdomain.Task, error)
	// ListTemplates возвращает только шаблоны (используется планировщиком).
	ListTemplates(ctx context.Context) ([]taskdomain.Task, error)
	// CreateBatch вставляет несколько задач за один запрос; дубликаты молча пропускаются.
	CreateBatch(ctx context.Context, tasks []taskdomain.Task) error
	// DeleteFutureByTemplate удаляет незапущенные экземпляры шаблона с scheduled_at >= from.
	DeleteFutureByTemplate(ctx context.Context, templateID int64, from time.Time) error
}

type Usecase interface {
	Create(ctx context.Context, input CreateInput) (*taskdomain.Task, error)
	GetByID(ctx context.Context, id int64) (*taskdomain.Task, error)
	Update(ctx context.Context, id int64, input UpdateInput) (*taskdomain.Task, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, includeTemplates bool) ([]taskdomain.Task, error)
	RefillTemplates(ctx context.Context) error
}

type CreateInput struct {
	Title       string
	Description string
	Status      taskdomain.Status
	ScheduledAt *time.Time

	RecurrenceType          *taskdomain.RecurrenceType
	RecurrenceDailyInterval *int
	RecurrenceMonthlyDays   []int
	RecurrenceSpecificDates []time.Time
	RecurrenceDayParity *taskdomain.Parity
}

type UpdateInput struct {
	Title       string
	Description string
	Status      taskdomain.Status
	ScheduledAt *time.Time

	// RecurrenceType == nil означает «снять правило» (PUT-семантика полной замены).
	RecurrenceType          *taskdomain.RecurrenceType
	RecurrenceDailyInterval *int
	RecurrenceMonthlyDays   []int
	RecurrenceSpecificDates []time.Time
	RecurrenceDayParity *taskdomain.Parity
}
