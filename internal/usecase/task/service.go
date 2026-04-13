package task

import (
	"context"
	"fmt"
	"strings"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type Service struct {
	repo Repository
	now  func() time.Time
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
		now:  func() time.Time { return time.Now().UTC() },
	}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (*taskdomain.Task, error) {
	normalized, err := validateCreateInput(input)
	if err != nil {
		return nil, err
	}

	model := &taskdomain.Task{
		Title:       normalized.Title,
		Description: normalized.Description,
		Status:      normalized.Status,
		Repeated:    normalized.Repeated,
	}

	now := s.now()
	model.CreatedAt = now
	model.UpdatedAt = now

	created, err := s.repo.Create(ctx, model)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (s *Service) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.repo.GetByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, id int64, input UpdateInput) (*taskdomain.Task, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	normalized, err := validateUpdateInput(input)
	if err != nil {
		return nil, err
	}

	model := &taskdomain.Task{
		ID:          id,
		Title:       normalized.Title,
		Description: normalized.Description,
		Status:      normalized.Status,
		Repeated:    normalized.Repeated,
		UpdatedAt:   s.now(),
	}

	updated, err := s.repo.Update(ctx, model)
	if err != nil {
		return nil, err
	}

	return updated, nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.repo.Delete(ctx, id)
}

func (s *Service) List(ctx context.Context) ([]taskdomain.Task, error) {
	return s.repo.List(ctx)
}

func validateCreateInput(input CreateInput) (CreateInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return CreateInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if input.Status == "" {
		input.Status = taskdomain.StatusNew
	}

	if !input.Status.Valid() {
		return CreateInput{}, fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}

	repeated, err := normalizeRepeated(input.Repeated)
	if err != nil {
		return CreateInput{}, err
	}
	input.Repeated = repeated

	return input, nil
}

func validateUpdateInput(input UpdateInput) (UpdateInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return UpdateInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if !input.Status.Valid() {
		return UpdateInput{}, fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}

	repeated, err := normalizeRepeated(input.Repeated)
	if err != nil {
		return UpdateInput{}, err
	}
	input.Repeated = repeated

	return input, nil
}

func normalizeRepeated(r taskdomain.Repeated) (taskdomain.Repeated, error) {
	if r.Type == "" {
		r.Type = taskdomain.PeriodNone
	}

	if !r.Valid() {
		return taskdomain.Repeated{}, fmt.Errorf("%w: invalid repeated settings", ErrInvalidInput)
	}

	switch r.Type {
	case taskdomain.PeriodNone:
		return taskdomain.Repeated{
			Type: taskdomain.PeriodNone,
		}, nil

	case taskdomain.PeriodEveryNDays:
		return taskdomain.Repeated{
			Type:       taskdomain.PeriodEveryNDays,
			EveryNDays: r.EveryNDays,
		}, nil

	case taskdomain.PeriodMonthlyDay:
		return taskdomain.Repeated{
			Type:       taskdomain.PeriodMonthlyDay,
			DayOfMonth: r.DayOfMonth,
		}, nil

	case taskdomain.PeriodSpecificDates:
		return taskdomain.Repeated{
			Type:          taskdomain.PeriodSpecificDates,
			SpecificDates: normalizeSpecificDates(r.SpecificDates),
		}, nil

	case taskdomain.PeriodEvenDays:
		return taskdomain.Repeated{
			Type: taskdomain.PeriodEvenDays,
		}, nil

	case taskdomain.PeriodOddDays:
		return taskdomain.Repeated{
			Type: taskdomain.PeriodOddDays,
		}, nil

	default:
		return taskdomain.Repeated{}, fmt.Errorf("%w: invalid repeated type", ErrInvalidInput)
	}
}

func normalizeSpecificDates(dates []time.Time) []time.Time {
	if len(dates) == 0 {
		return nil
	}

	result := make([]time.Time, 0, len(dates))
	seen := make(map[string]struct{}, len(dates))

	for _, d := range dates {
		normalized := time.Date(d.UTC().Year(), d.UTC().Month(), d.UTC().Day(), 0, 0, 0, 0, time.UTC)
		key := normalized.Format("2006-01-02")

		if _, ok := seen[key]; ok {
			continue
		}

		seen[key] = struct{}{}
		result = append(result, normalized)
	}

	return result
}
