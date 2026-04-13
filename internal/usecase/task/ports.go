package task

import (
	"context"
	"time"

	recurrencedomain "example.com/taskservice/internal/domain/recurrence"
	taskdomain "example.com/taskservice/internal/domain/task"
)

type Repository interface {
	Create(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error)
	CreateIfNotExists(ctx context.Context, task *taskdomain.Task) error
	CreateRecurrence(ctx context.Context, task *recurrencedomain.Recurrence) (*recurrencedomain.Recurrence, error)
	ListRecurrence(ctx context.Context) ([]recurrencedomain.Recurrence, error)
	GetByID(ctx context.Context, id int64) (*taskdomain.Task, error)
	Update(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context) ([]taskdomain.Task, error)
}

type Usecase interface {
	Create(ctx context.Context, input CreateInput) (*taskdomain.Task, error)
	CreateRecurrence(ctx context.Context, input CreateRecurrenceInput) (*recurrencedomain.Recurrence, error)
	CreateRecurrencedTasks(ctx context.Context, from time.Time, to time.Time) error
	ListRecurrence(ctx context.Context) ([]recurrencedomain.Recurrence, error)
	GetByID(ctx context.Context, id int64) (*taskdomain.Task, error)
	Update(ctx context.Context, id int64, input UpdateInput) (*taskdomain.Task, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context) ([]taskdomain.Task, error)
}

type Recurrence struct {
	StartDate     *time.Time
	EndDate       *time.Time
	IntervalDays  *int
	MonthDays     *[]int
	SpecificDates *[]time.Time
	EvenOdd       *recurrencedomain.EvenOdd
}

type CreateRecurrenceInput struct {
	Title       string
	Description string
	Recurrence  Recurrence
}

type CreateInput struct {
	Title       string
	Description string
	DueDate     time.Time
	Status      taskdomain.Status
}

type UpdateInput struct {
	Title       string
	Description string
	DueDate     time.Time
	Status      taskdomain.Status
}
