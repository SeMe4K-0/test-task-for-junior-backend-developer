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

	now := s.now()
	
	// If it's a recurring task
	if normalized.RecurrenceConfig != nil {
		template := &taskdomain.Task{
			Title:            normalized.Title,
			Description:      normalized.Description,
			Status:           normalized.Status,
			IsTemplate:       true,
			RecurrenceConfig: normalized.RecurrenceConfig,
			CreatedAt:        now,
			UpdatedAt:        now,
		}

		instanceDates := generateDates(normalized.RecurrenceConfig)
		instances := make([]taskdomain.Task, 0, len(instanceDates))
		for _, date := range instanceDates {
			d := date
			instances = append(instances, taskdomain.Task{
				Title:       normalized.Title,
				Description: normalized.Description,
				Status:      taskdomain.StatusNew,
				DueDate:     &d,
				IsTemplate:  false,
				CreatedAt:   now,
				UpdatedAt:   now,
			})
		}

		created, err := s.repo.CreateWithInstances(ctx, template, instances)
		if err != nil {
			return nil, err
		}
		return created, nil
	}

	// Simple one-off task
	model := &taskdomain.Task{
		Title:       normalized.Title,
		Description: normalized.Description,
		Status:      normalized.Status,
		DueDate:     normalized.DueDate,
		IsTemplate:  false,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

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
		DueDate:     normalized.DueDate,
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

	if input.RecurrenceConfig != nil {
		config := input.RecurrenceConfig
		if config.Type == "" {
			return CreateInput{}, fmt.Errorf("%w: recurrence type is required", ErrInvalidInput)
		}
		if config.StartDate.IsZero() {
			return CreateInput{}, fmt.Errorf("%w: recurrence start_date is required", ErrInvalidInput)
		}
		if config.EndDate.IsZero() {
			return CreateInput{}, fmt.Errorf("%w: recurrence end_date is required", ErrInvalidInput)
		}
		if config.EndDate.Before(config.StartDate) {
			return CreateInput{}, fmt.Errorf("%w: recurrence end_date must be after start_date", ErrInvalidInput)
		}

		switch config.Type {
		case taskdomain.RecurrenceDaily:
			if config.Interval != nil && *config.Interval <= 0 {
				return CreateInput{}, fmt.Errorf("%w: daily interval must be positive", ErrInvalidInput)
			}
		case taskdomain.RecurrenceMonthly:
			if config.Day == nil || *config.Day < 1 || *config.Day > 31 {
				return CreateInput{}, fmt.Errorf("%w: monthly day must be between 1 and 31", ErrInvalidInput)
			}
		case taskdomain.RecurrenceSpecificDates:
			if len(config.Dates) == 0 {
				return CreateInput{}, fmt.Errorf("%w: specific dates are required", ErrInvalidInput)
			}
		case taskdomain.RecurrenceParity:
			if config.Parity == nil || (strings.ToLower(*config.Parity) != "even" && strings.ToLower(*config.Parity) != "odd") {
				return CreateInput{}, fmt.Errorf("%w: parity must be 'even' or 'odd'", ErrInvalidInput)
			}
		default:
			return CreateInput{}, fmt.Errorf("%w: unsupported recurrence type: %s", ErrInvalidInput, config.Type)
		}
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

func generateDates(config *taskdomain.RecurrenceConfig) []time.Time {
	if config == nil {
		return nil
	}

	var dates []time.Time
	start := config.StartDate
	end := config.EndDate

	// Safety limit: max 1 year
	if end.After(start.AddDate(1, 0, 0)) {
		end = start.AddDate(1, 0, 0)
	}

	switch config.Type {
	case taskdomain.RecurrenceDaily:
		interval := 1
		if config.Interval != nil && *config.Interval > 0 {
			interval = *config.Interval
		}
		for d := start; !d.After(end); d = d.AddDate(0, 0, interval) {
			dates = append(dates, d)
		}

	case taskdomain.RecurrenceMonthly:
		if config.Day == nil {
			return nil
		}
		day := *config.Day
		// We iterate by day and match the day of month. 
		// This handles months with different lengths correctly (e.g. if day is 31, it only appears in some months).
		for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
			if d.Day() == day {
				dates = append(dates, d)
			}
		}

	case taskdomain.RecurrenceSpecificDates:
		for _, d := range config.Dates {
			if !d.Before(start) && !d.After(end) {
				dates = append(dates, d)
			}
		}

	case taskdomain.RecurrenceParity:
		if config.Parity == nil {
			return nil
		}
		parity := strings.ToLower(*config.Parity)
		for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
			isEven := d.Day()%2 == 0
			if (parity == "even" && isEven) || (parity == "odd" && !isEven) {
				dates = append(dates, d)
			}
		}
	}

	// Safety limit: max 366 instances
	if len(dates) > 366 {
		dates = dates[:366]
	}

	return dates
}
