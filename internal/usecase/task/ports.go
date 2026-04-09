package task

import "time"

import (
	"context"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type Repository interface {
	Create(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error)
	// новое
	CreatePereodic(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error)
	//
	GetByID(ctx context.Context, id int64) (*taskdomain.Task, error)
	Update(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context) ([]taskdomain.Task, error)
}

type Usecase interface {
	Create(ctx context.Context, input CreateInput) (*taskdomain.Task, error)
	//
	GetByID(ctx context.Context, id int64) (*taskdomain.Task, error)
	Update(ctx context.Context, id int64, input UpdateInput) (*taskdomain.Task, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context) ([]taskdomain.Task, error)
	// нвоое
	CreatePereodic(ctx context.Context, input CreatePereodicInput) (*taskdomain.Task, error)
	// UpdatePereodic(ctx context.Context, id int64, input CreatePereodicInput) (*taskdomain.Task, error)
	// DeletePereodic(ctx context.Context, id int64) error
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

type CreatePereodicInput struct {
	Title       string
	Description string
	Status      taskdomain.Status
	//
	Repetition  int
	DateAndTime time.Time
}
