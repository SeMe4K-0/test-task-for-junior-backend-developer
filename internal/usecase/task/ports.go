package task

import (
	"context"

	taskdomain "example.com/taskservice/internal/domain/task"
)

const (
	DefaultPageLimit      = 50
	MaxPageLimit          = 100
	DefaultSeedWindowDays = 30
	DefaultSyncPastDays   = 1
)

type Repository interface {
	Ping(ctx context.Context) error

	CreateTemplate(ctx context.Context, template *taskdomain.Template) (*taskdomain.Template, error)
	GetTemplateByID(ctx context.Context, id int64) (*taskdomain.Template, error)
	UpdateTemplate(ctx context.Context, template *taskdomain.Template) (*taskdomain.Template, error)
	DeleteTemplate(ctx context.Context, id int64) error
	ListTemplates(ctx context.Context, filter TemplateListInput) ([]taskdomain.Template, int, error)
	ListTemplatesForSync(ctx context.Context, templateID *int64) ([]taskdomain.Template, error)

	UpsertOccurrences(ctx context.Context, tasks []taskdomain.Task) error
	GetOccurrenceByID(ctx context.Context, id int64) (*taskdomain.Task, error)
	GetOccurrenceByTemplateAndDate(ctx context.Context, templateID int64, scheduledFor taskdomain.Date) (*taskdomain.Task, error)
	UpdateOccurrence(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error)
	DeleteOccurrence(ctx context.Context, id int64) error
	DeleteOccurrencesFrom(ctx context.Context, templateID int64, from taskdomain.Date) error
	CountOccurrencesByTemplate(ctx context.Context, templateID int64) (int, error)
	ListOccurrences(ctx context.Context, filter TaskListInput) ([]taskdomain.Task, int, error)
}

type TaskUsecase interface {
	Create(ctx context.Context, input CreateTaskInput) (*taskdomain.Task, error)
	GetByID(ctx context.Context, id int64) (*taskdomain.Task, error)
	Update(ctx context.Context, id int64, input UpdateTaskInput) (*taskdomain.Task, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, input TaskListInput) (TaskListResult, error)
}

type TemplateUsecase interface {
	CreateTemplate(ctx context.Context, input CreateTemplateInput) (*taskdomain.Template, error)
	GetTemplateByID(ctx context.Context, id int64) (*taskdomain.Template, error)
	UpdateTemplate(ctx context.Context, id int64, input UpdateTemplateInput) (*taskdomain.Template, error)
	DeleteTemplate(ctx context.Context, id int64) error
	ListTemplates(ctx context.Context, input TemplateListInput) (TemplateListResult, error)
}

type CreateTaskInput struct {
	Title       string
	Description string
	Status      taskdomain.Status
	Schedule    *taskdomain.Schedule
}

type UpdateTaskInput struct {
	Title       string
	Description string
	Status      taskdomain.Status
	Schedule    *taskdomain.Schedule
}

type TaskListInput struct {
	Limit      int
	Offset     int
	Status     *taskdomain.Status
	TemplateID *int64
	DateFrom   *taskdomain.Date
	DateTo     *taskdomain.Date
}

type TaskListResult struct {
	Items  []taskdomain.Task
	Total  int
	Limit  int
	Offset int
}

type CreateTemplateInput struct {
	Title       string
	Description string
	Schedule    *taskdomain.Schedule
}

type UpdateTemplateInput struct {
	Title       string
	Description string
	Schedule    *taskdomain.Schedule
}

type TemplateListInput struct {
	Limit        int
	Offset       int
	Search       string
	ScheduleType *taskdomain.ScheduleType
}

type TemplateListResult struct {
	Items  []taskdomain.Template
	Total  int
	Limit  int
	Offset int
}
