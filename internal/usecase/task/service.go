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
	model := &taskdomain.Task{
		Title:         normalized.Title,
		Description:   normalized.Description,
		Status:        normalized.Status,
		ScheduledDate: normalized.ScheduledDate,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	return s.repo.Create(ctx, model)
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

	return s.repo.Update(ctx, model)
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

// CreateRecurring generates task instances for every date produced by the
// recurrence rule and persists them in a single pass.
//
// Each instance is an independent task — it can be completed, edited, or deleted
// individually. The recurrence settings are stored on every instance so the
// original schedule is always visible.
func (s *Service) CreateRecurring(ctx context.Context, input CreateRecurringInput) ([]*taskdomain.Task, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return nil, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if input.Status == "" {
		input.Status = taskdomain.StatusNew
	}

	if !input.Status.Valid() {
		return nil, fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}

	if err := input.Recurrence.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidInput, err)
	}

	if input.Recurrence.Type != taskdomain.RecurrenceTypeDates {
		if err := validateDateRange(input.StartDate, input.EndDate); err != nil {
			return nil, fmt.Errorf("%w: %s", ErrInvalidInput, err)
		}
	}

	dates := input.Recurrence.GenerateDates(input.StartDate, input.EndDate)
	if len(dates) == 0 {
		return nil, fmt.Errorf("%w: recurrence produced no dates in the given range", ErrInvalidInput)
	}

	recurrence := input.Recurrence
	now := s.now()

	tasks := make([]*taskdomain.Task, 0, len(dates))
	for _, d := range dates {
		scheduledDate := d
		model := &taskdomain.Task{
			Title:         input.Title,
			Description:   input.Description,
			Status:        input.Status,
			ScheduledDate: &scheduledDate,
			Recurrence:    &recurrence,
			CreatedAt:     now,
			UpdatedAt:     now,
		}
		created, err := s.repo.Create(ctx, model)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, created)
	}

	return tasks, nil
}

func validateDateRange(start, end time.Time) error {
	if start.IsZero() || end.IsZero() {
		return fmt.Errorf("start_date and end_date are required")
	}
	if end.Before(start) {
		return fmt.Errorf("end_date must not be before start_date")
	}
	if int(end.Sub(start).Hours()/24) > taskdomain.MaxDateRangeDays {
		return fmt.Errorf("date range must not exceed %d days", taskdomain.MaxDateRangeDays)
	}
	return nil
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
