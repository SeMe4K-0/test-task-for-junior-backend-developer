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
	Delete(ctx context.Context, id int64, now time.Time) error
	Restore(ctx context.Context, id int64, now time.Time) (*taskdomain.Task, error)
	List(ctx context.Context, pagination Pagination) ([]taskdomain.Task, int, error)
	ListWithFilter(ctx context.Context, filter ListFilter, pagination Pagination) ([]taskdomain.Task, int, error)
}

type Usecase interface {
	Create(ctx context.Context, input CreateInput) (*taskdomain.Task, error)
	GetByID(ctx context.Context, id int64) (*taskdomain.Task, error)
	Update(ctx context.Context, id int64, input UpdateInput) (*taskdomain.Task, error)
	Delete(ctx context.Context, id int64) error
	Restore(ctx context.Context, id int64) (*taskdomain.Task, error)
	List(ctx context.Context, pagination Pagination) (*ListOutput, error)
	ListWithFilter(ctx context.Context, filter ListFilter, pagination Pagination) (*ListOutput, error)
}

type CreateInput struct {
	Title       string
	Description string
	Status      taskdomain.Status
}

type UpdateInput struct {
	Title       string
	Description string
	Status      taskdomain.Status
}

type ListFilter struct {
	ScheduledDateFrom *time.Time
	ScheduledDateTo   *time.Time
}

type Pagination struct {
	Page  int
	Limit int
}

const (
	DefaultPage  = 1
	DefaultLimit = 20
	MaxLimit     = 100
)

func (p Pagination) Offset() int {
	return (p.Page - 1) * p.Limit
}

func (p Pagination) Normalize() Pagination {
	if p.Page < 1 {
		p.Page = DefaultPage
	}
	if p.Limit < 1 {
		p.Limit = DefaultLimit
	}
	if p.Limit > MaxLimit {
		p.Limit = MaxLimit
	}
	return p
}

type ListOutput struct {
	Tasks []taskdomain.Task
	Total int
	Page  int
	Limit int
}
