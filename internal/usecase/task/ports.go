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
	List(ctx context.Context) ([]taskdomain.Task, error)
}

type Usecase interface {
	Create(ctx context.Context, input CreateInput) (*taskdomain.Task, error)
	GetByID(ctx context.Context, id int64) (*taskdomain.Task, error)
	Update(ctx context.Context, id int64, input UpdateInput) (*taskdomain.Task, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context) ([]taskdomain.Task, error)
	CreateRecurring(ctx context.Context, input CreateRecurringInput) ([]*taskdomain.Task, error)
}

type CreateInput struct {
	Title         string
	Description   string
	Status        taskdomain.Status
	ScheduledDate *time.Time
}

type UpdateInput struct {
	Title       string
	Description string
	Status      taskdomain.Status
}

// CreateRecurringInput defines a recurring task template.
// StartDate and EndDate are required for all recurrence types except RecurrenceTypeDates,
// where the dates are listed explicitly inside Recurrence.Dates.
type CreateRecurringInput struct {
	Title       string
	Description string
	Status      taskdomain.Status
	Recurrence  taskdomain.Recurrence
	StartDate   time.Time
	EndDate     time.Time
}
