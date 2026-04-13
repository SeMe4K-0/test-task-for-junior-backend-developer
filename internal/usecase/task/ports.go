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
	List(ctx context.Context, from, to time.Time) ([]taskdomain.Task, error)
	CreateRule(ctx context.Context, rule *taskdomain.TaskRule) (*taskdomain.TaskRule, error)
	ListActiveRules(ctx context.Context, from, to time.Time) ([]taskdomain.TaskRule, error)
	GetRuleByID(ctx context.Context, id int64) (*taskdomain.TaskRule, error)
	ListRules(ctx context.Context) ([]taskdomain.TaskRule, error)
	DeleteRule(ctx context.Context, id int64) error
}

type Usecase interface {
	Create(ctx context.Context, input CreateInput) (*taskdomain.Task, error)
	GetByID(ctx context.Context, id int64) (*taskdomain.Task, error)
	Materialize(ctx context.Context, input MaterializeInput) (*taskdomain.Task, error)
	Update(ctx context.Context, id int64, input UpdateInput) (*taskdomain.Task, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, from, to time.Time) ([]taskdomain.Task, error)
	ListRules(ctx context.Context) ([]taskdomain.TaskRule, error)
	DeleteRule(ctx context.Context, id int64) error
}

type CreateInput struct {
	Title         string
	Description   string
	Status        taskdomain.Status
	ExecutionDate time.Time 
	Recurrence    *RecurrenceInput
}

type UpdateInput struct {
	Title         string
	Description   string
	Status        taskdomain.Status
	ExecutionDate time.Time
}

type RecurrenceInput struct {
	Type      taskdomain.RecurrenceType
	Value     []byte
	StartDate time.Time
	EndDate   *time.Time
}

type MaterializeInput struct {
	Title         string
	Description   string
	Status        taskdomain.Status
	ExecutionDate time.Time
	RuleID        int64
}
