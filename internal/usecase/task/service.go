package task

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	recurrencedomain "example.com/taskservice/internal/domain/recurrence"
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

// Recurrences

func (s *Service) CreateRecurrence(ctx context.Context, input CreateRecurrenceInput) (*recurrencedomain.Recurrence, error) {
	normalized, err := validateRecurrenceInput(input)
	if err != nil {
		return nil, err
	}
	model := recurrencedomain.New(
		normalized.Title,
		normalized.Description,
		normalized.Recurrence.StartDate,
		normalized.Recurrence.EndDate,
		normalized.Recurrence.IntervalDays,
		normalized.Recurrence.MonthDays,
		normalized.Recurrence.SpecificDates,
		normalized.Recurrence.EvenOdd,
	)

	created, err := s.repo.CreateRecurrence(ctx, &model)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (s *Service) CreateRecurrencedTasks(ctx context.Context, from, to time.Time) error {
	if from.After(to) {
		return (fmt.Errorf("'from' must be <= 'to'"))
	}

	recurrences, err := s.repo.ListRecurrence(ctx)
	if err != nil {
		return fmt.Errorf("list recurrences: %w", err)
	}

	for _, rec := range recurrences {
		if !rec.Active {
			continue
		}

		if err := s.generateTasksForRule(ctx, rec, from, to); err != nil {
			return fmt.Errorf("generate tasks for rule %d: %w", rec.ID, err)
		}
	}

	return nil
}

func (s *Service) generateTasksForRule(ctx context.Context, rec recurrencedomain.Recurrence, from, to time.Time) error {
	for day := from; !day.After(to); day = day.AddDate(0, 0, 1) {
		if err := rec.Matches(day); err != nil {
			if errors.Is(err, recurrencedomain.ErrDateNotMatched) || errors.Is(err, recurrencedomain.ErrDateOutOfRange) {
				continue
			}
			return err
		}

		task := buildTask(rec, day, s.now())
		if err := s.repo.CreateIfNotExists(ctx, task); err != nil {
			return fmt.Errorf("create task for date %s: %w", day.Format("2006-01-02"), err)
		}
	}

	return nil
}

func buildTask(rec recurrencedomain.Recurrence, scheduledFor time.Time, now time.Time) *taskdomain.Task {
	day := scheduledFor.UTC().Truncate(24 * time.Hour)
	return &taskdomain.Task{
		RuleID:       &rec.ID,
		Title:        rec.Title,
		Description:  rec.Description,
		Status:       taskdomain.StatusNew,
		ScheduledFor: &day,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

func (s *Service) ListRecurrence(ctx context.Context) ([]recurrencedomain.Recurrence, error) {
	return s.repo.ListRecurrence(ctx)
}

// Tasks

func (s *Service) Create(ctx context.Context, input CreateInput) (*taskdomain.Task, error) {
	normalized, err := validateCreateInput(input)
	if err != nil {
		return nil, err
	}

	model := &taskdomain.Task{
		Title:       normalized.Title,
		Description: normalized.Description,
		Status:      normalized.Status,
		DueDate:     &normalized.DueDate,
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

func validateRecurrenceInput(input CreateRecurrenceInput) (CreateRecurrenceInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return CreateRecurrenceInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if input.Recurrence.StartDate == nil {
		t := time.Now()
		input.Recurrence.StartDate = &t
	}

	if input.Recurrence.EndDate != nil {
		if input.Recurrence.EndDate.Before(*input.Recurrence.StartDate) {
			return CreateRecurrenceInput{}, fmt.Errorf("%w: end_date can't be before start_date", ErrInvalidInput)
		}
	}
	notNilCount := 0

	if input.Recurrence.IntervalDays != nil {
		notNilCount++
	}
	if input.Recurrence.MonthDays != nil {
		notNilCount++
	}
	if input.Recurrence.SpecificDates != nil {
		notNilCount++
	}
	if input.Recurrence.EvenOdd != nil {
		notNilCount++
	}

	if notNilCount == 0 {
		return CreateRecurrenceInput{}, fmt.Errorf("%w: one of recurrence should be specified (interval_days, month_days, specific_dates, even_odd)", ErrInvalidInput)
	}

	if notNilCount > 1 {
		return CreateRecurrenceInput{}, fmt.Errorf("%w: only one recurrence type can be specified, got %d", ErrInvalidRecurrenceInput, notNilCount)
	}

	if input.Recurrence.IntervalDays != nil && *input.Recurrence.IntervalDays <= 0 {
		return CreateRecurrenceInput{}, fmt.Errorf("%w: interval_days can't be less than a 0", ErrInvalidInput)
	}

	if input.Recurrence.MonthDays != nil && len(*input.Recurrence.MonthDays) == 0 {
		return CreateRecurrenceInput{}, fmt.Errorf("%w: month_days cant't be less than a 0", ErrInvalidInput)
	}

	if input.Recurrence.MonthDays != nil {
		for _, d := range *input.Recurrence.MonthDays {
			if d < 1 || d > 30 {
				return CreateRecurrenceInput{}, fmt.Errorf("%w: month_days values must be between 1 and 30", ErrInvalidInput)
			}
		}
	}

	if input.Recurrence.SpecificDates != nil && len(*input.Recurrence.SpecificDates) == 0 {
		return CreateRecurrenceInput{}, fmt.Errorf("%w: specific_dates must not be empty", ErrInvalidInput)
	}

	return input, nil
}
