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
		Periodicity: normalized.Periodicity,
		ScheduledAt: normalized.ScheduledAt,
		Status:      normalized.Status,
		Frequency:   normalized.Frequency,
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
		Periodicity: normalized.Periodicity,
		ScheduledAt: normalized.ScheduledAt,
		Frequency:   normalized.Frequency,
		Status:      normalized.Status,
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

	if input.Periodicity == "" {
		input.Periodicity = taskdomain.PeriodicityOnce
	}

	if input.Periodicity != taskdomain.PeriodicitySetDates {
		input.ScheduledAt = input.ScheduledAt[:1]
	}

	normalizedFreq, err := validateFrequency(input.Periodicity, input.Frequency)
	if err != nil {
		return CreateInput{}, err
	}

	input.Frequency = normalizedFreq

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

	if !input.Periodicity.Valid() {
		return UpdateInput{}, fmt.Errorf("%w: invalid periodicity", ErrInvalidInput)
	}

	if input.Periodicity != taskdomain.PeriodicitySetDates {
		input.ScheduledAt = input.ScheduledAt[:1]
	}

	normalizedFreq, err := validateFrequency(input.Periodicity, input.Frequency)
	if err != nil {
		return UpdateInput{}, err
	}

	input.Frequency = normalizedFreq

	return input, nil
}

func validateFrequency(periodicity taskdomain.Periodicity, frequency int64) (int64, error) {
	newFrequency := frequency
	var err error = nil

	switch periodicity {
	case taskdomain.PeriodicityOnce, taskdomain.PeriodicitySetDates:
		newFrequency = 0
	case taskdomain.PeriodicityInterval:
		break
	case taskdomain.PeriodicityMonthly:
		if frequency < 1 || frequency > 31 {
			newFrequency = -1
			err = fmt.Errorf("%w: invalid frequency value", ErrInvalidInput)
		}
	case taskdomain.PeriodicityParity:
		if frequency < 0 || frequency > 1 {
			newFrequency = -1
			err = fmt.Errorf("%w: invalid frequency value", ErrInvalidInput)
		}
	default:
		newFrequency = -1
		err = fmt.Errorf("%w: invalid frequency value", ErrInvalidInput)
	}
	return newFrequency, err
}
