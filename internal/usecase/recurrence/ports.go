package recurrence

import (
	"context"
	"time"

	recurrencedomain "example.com/taskservice/internal/domain/recurrence"
	taskdomain "example.com/taskservice/internal/domain/task"
)

type Repository interface {
	CreateWithTasks(ctx context.Context, rule *recurrencedomain.RecurrenceRule, tasks []taskdomain.Task) (*recurrencedomain.RecurrenceRule, []taskdomain.Task, error)
	GetByID(ctx context.Context, id int64) (*recurrencedomain.RecurrenceRule, error)
	List(ctx context.Context) ([]recurrencedomain.RecurrenceRule, error)
	UpdateWithTasks(ctx context.Context, rule *recurrencedomain.RecurrenceRule, tasks []taskdomain.Task) (*recurrencedomain.RecurrenceRule, []taskdomain.Task, error)
	Delete(ctx context.Context, id int64, deleteTasks bool, now time.Time) error
}

type Usecase interface {
	Create(ctx context.Context, input CreateInput) (*CreateOutput, error)
	GetByID(ctx context.Context, id int64) (*recurrencedomain.RecurrenceRule, error)
	List(ctx context.Context) ([]recurrencedomain.RecurrenceRule, error)
	Update(ctx context.Context, id int64, input UpdateInput) (*UpdateOutput, error)
	Delete(ctx context.Context, id int64, deleteTasks bool) error
}

type RecurrenceParams struct {
	Type          recurrencedomain.RecurrenceType `json:"type"`
	EveryNDays    *int                            `json:"every_n_days,omitempty"`
	DayOfMonth    *int                            `json:"day_of_month,omitempty"`
	Dates         []time.Time                     `json:"dates,omitempty"`
	EvenOdd       recurrencedomain.EvenOddType    `json:"even_odd,omitempty"`
}

type CreateInput struct {
	TaskTitle       string
	TaskDescription string
	Recurrence      RecurrenceParams
	StartDate       time.Time
	EndDate         time.Time
}

type UpdateInput struct {
	TaskTitle       string
	TaskDescription string
	Recurrence      RecurrenceParams
	StartDate       time.Time
	EndDate         time.Time
}

type CreateOutput struct {
	Rule  *recurrencedomain.RecurrenceRule
	Tasks []taskdomain.Task
}

type UpdateOutput struct {
	Rule  *recurrencedomain.RecurrenceRule
	Tasks []taskdomain.Task
}
