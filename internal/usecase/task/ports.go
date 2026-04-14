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
	ListByScheduledDate(ctx context.Context, date string) ([]taskdomain.Task, error)
	ExistsByParentAndDate(ctx context.Context, parentID int64, date string) (bool, error)
	DeleteFutureInstancesByParent(ctx context.Context, parentID int64, fromDate string) error
}

type RecurrenceRepository interface {
	Create(ctx context.Context, rule *taskdomain.RecurrenceRule) (*taskdomain.RecurrenceRule, error)
	GetByTaskID(ctx context.Context, taskID int64) (*taskdomain.RecurrenceRule, error)
	Update(ctx context.Context, rule *taskdomain.RecurrenceRule) (*taskdomain.RecurrenceRule, error)
	Delete(ctx context.Context, taskID int64) error
	ListActive(ctx context.Context, date time.Time) ([]taskdomain.RecurrenceRule, error)
}

type Usecase interface {
	Create(ctx context.Context, input CreateInput) (*taskdomain.Task, error)
	GetByID(ctx context.Context, id int64) (*taskdomain.Task, error)
	Update(ctx context.Context, id int64, input UpdateInput) (*taskdomain.Task, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context) ([]taskdomain.Task, error)
	ListByDate(ctx context.Context, date string) ([]taskdomain.Task, error)
	GetRecurrence(ctx context.Context, taskID int64) (*taskdomain.RecurrenceRule, error)
	UpdateRecurrence(ctx context.Context, taskID int64, input RecurrenceInput) (*taskdomain.RecurrenceRule, error)
	DeleteRecurrence(ctx context.Context, taskID int64) error
}

type CreateInput struct {
	Title       string
	Description string
	Status      taskdomain.Status
	Recurrence  *RecurrenceInput // nil = обычная задача
}

type UpdateInput struct {
	Title       string
	Description string
	Status      taskdomain.Status
}

type RecurrenceInput struct {
	Type          taskdomain.RecurrenceType
	IntervalDays  *int
	DayOfMonth    *int
	SpecificDates []string // формат "2006-01-02"
	EvenOdd       *string
	StartDate     string
	EndDate       *string
}
