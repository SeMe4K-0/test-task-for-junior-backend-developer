package schedule

import (
	"context"
	"encoding/json"
	scheduledomain "example.com/taskservice/internal/domain/schedule"
)

type Repository interface {
	Create(ctx context.Context, schedule *scheduledomain.Schedule) (*scheduledomain.Schedule, error)
	GetByID(ctx context.Context, id int64) (*scheduledomain.Schedule, error)
	Update(ctx context.Context, schedule *scheduledomain.Schedule) (*scheduledomain.Schedule, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context) ([]scheduledomain.Schedule, error)
}

type Usecase interface {
	Create(ctx context.Context, input CreateInput) (*scheduledomain.Schedule, error)
	GetByID(ctx context.Context, id int64) (*scheduledomain.Schedule, error)
	Update(ctx context.Context, id int64, input UpdateInput) (*scheduledomain.Schedule, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context) ([]scheduledomain.Schedule, error)
}

type CreateInput struct {
	Title            string
	Description      string
	RecurrenceType   scheduledomain.RecurrenceType
	RecurrenceParams json.RawMessage
}

type UpdateInput struct {
	Title            string
	Description      string
	RecurrenceType   scheduledomain.RecurrenceType
	RecurrenceParams json.RawMessage
	IsActive         bool
}
