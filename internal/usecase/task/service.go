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
		Recurrence:  input.Recurrence,
	}
	now := s.now()

	if input.Recurrence != nil {
		next := calculateNextRun(input.Recurrence, now)
		model.NextRunAt = next
	}
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

	return input, nil
}

if input.Recurrence != nil {
	if err := input.Recurrence.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}
}

func calculateNextRun(r *taskdomain.Recurrence, now time.Time) *time.Time {
	switch r.Type {

	case taskdomain.RecurrenceDaily:
		next := now.AddDate(0, 0, *r.IntervalDays)
		return &next

	case taskdomain.RecurrenceEvenDays, taskdomain.RecurrenceOddDays:
		for i := 1; i <= 31; i++ {
			d := now.AddDate(0, 0, i)
			day := d.Day()

			if r.Type == taskdomain.RecurrenceEvenDays && day%2 == 0 {
				return &d
			}
			if r.Type == taskdomain.RecurrenceOddDays && day%2 != 0 {
				return &d
			}
		}

	case taskdomain.RecurrenceMonthly:
		for i := 1; i <= 365; i++ {
			d := now.AddDate(0, 0, i)
			for _, day := range r.DaysOfMonth {
				if d.Day() == day {
					return &d
				}
			}
		}

	case taskdomain.RecurrenceSpecificDates:
		var closest *time.Time
		for _, d := range r.Dates {
			if d.After(now) {
				if closest == nil || d.Before(*closest) {
					tmp := d
					closest = &tmp
				}
			}
		}
		return closest
	}

	return nil
}

func (s *Service) Generate(ctx context.Context) error {
	tasks, err := s.repo.List(ctx)
	if err != nil {
		return err
	}

	now := s.now()

	for _, t := range tasks {
		if t.Recurrence == nil || t.NextRunAt == nil {
			continue
		}

		if t.NextRunAt.After(now) {
			continue
		}

		newTask := &taskdomain.Task{
			Title:       t.Title,
			Description: t.Description,
			Status:      taskdomain.StatusNew,
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		_, err := s.repo.Create(ctx, newTask)
		if err != nil {
			return err
		}

		next := calculateNextRun(t.Recurrence, now)
		t.NextRunAt = next

		_, err = s.repo.Update(ctx, &t)
		if err != nil {
			return err
		}
	}

	return nil
}