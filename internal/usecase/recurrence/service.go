package recurrence

import (
	"context"
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

func (s *Service) Create(ctx context.Context, input CreateInput) (*CreateOutput, error) {
	if err := validateInput(input.TaskTitle, input.TaskDescription, input.Recurrence, input.StartDate, input.EndDate); err != nil {
		return nil, err
	}

	rule := buildRule(input.TaskTitle, input.TaskDescription, input.Recurrence, input.StartDate, input.EndDate)
	now := s.now()
	rule.CreatedAt = now
	rule.UpdatedAt = now

	dates, err := recurrencedomain.GenerateDates(rule)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidInput, err.Error())
	}

	tasks := buildTasks(rule, dates, now)

	createdRule, createdTasks, err := s.repo.CreateWithTasks(ctx, rule, tasks)
	if err != nil {
		return nil, err
	}

	return &CreateOutput{Rule: createdRule, Tasks: createdTasks}, nil
}

func (s *Service) GetByID(ctx context.Context, id int64) (*recurrencedomain.RecurrenceRule, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}
	return s.repo.GetByID(ctx, id)
}

func (s *Service) List(ctx context.Context) ([]recurrencedomain.RecurrenceRule, error) {
	return s.repo.List(ctx)
}

func (s *Service) Update(ctx context.Context, id int64, input UpdateInput) (*UpdateOutput, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	if err := validateInput(input.TaskTitle, input.TaskDescription, input.Recurrence, input.StartDate, input.EndDate); err != nil {
		return nil, err
	}

	rule := buildRule(input.TaskTitle, input.TaskDescription, input.Recurrence, input.StartDate, input.EndDate)
	rule.ID = id
	now := s.now()
	rule.UpdatedAt = now

	dates, err := recurrencedomain.GenerateDates(rule)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidInput, err.Error())
	}

	tasks := buildTasks(rule, dates, now)

	updatedRule, updatedTasks, err := s.repo.UpdateWithTasks(ctx, rule, tasks)
	if err != nil {
		return nil, err
	}

	return &UpdateOutput{Rule: updatedRule, Tasks: updatedTasks}, nil
}

func (s *Service) Delete(ctx context.Context, id int64, deleteTasks bool) error {
	if id <= 0 {
		return fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}
	return s.repo.Delete(ctx, id, deleteTasks, s.now())
}

func validateInput(title, description string, params RecurrenceParams, startDate, endDate time.Time) error {
	title = strings.TrimSpace(title)
	if title == "" {
		return fmt.Errorf("%w: task_title is required", ErrInvalidInput)
	}

	if !params.Type.Valid() {
		return fmt.Errorf("%w: invalid recurrence type", ErrInvalidInput)
	}

	switch params.Type {
	case recurrencedomain.TypeDaily:
		if params.EveryNDays != nil && *params.EveryNDays < 1 {
			return fmt.Errorf("%w: every_n_days must be >= 1", ErrInvalidInput)
		}
	case recurrencedomain.TypeMonthly:
		if params.DayOfMonth == nil {
			return fmt.Errorf("%w: day_of_month is required for monthly recurrence", ErrInvalidInput)
		}
		if *params.DayOfMonth < 1 || *params.DayOfMonth > 31 {
			return fmt.Errorf("%w: day_of_month must be between 1 and 31", ErrInvalidInput)
		}
	case recurrencedomain.TypeSpecificDates:
		if len(params.Dates) == 0 {
			return fmt.Errorf("%w: dates list is required for specific_dates recurrence", ErrInvalidInput)
		}
	case recurrencedomain.TypeEvenOdd:
		if !params.EvenOdd.Valid() {
			return fmt.Errorf("%w: even_odd must be 'even' or 'odd'", ErrInvalidInput)
		}
	}

	if startDate.IsZero() {
		return fmt.Errorf("%w: start_date is required", ErrInvalidInput)
	}
	if endDate.IsZero() {
		return fmt.Errorf("%w: end_date is required", ErrInvalidInput)
	}

	return nil
}

func buildRule(title, description string, params RecurrenceParams, startDate, endDate time.Time) *recurrencedomain.RecurrenceRule {
	return &recurrencedomain.RecurrenceRule{
		Type:            params.Type,
		EveryNDays:      params.EveryNDays,
		DayOfMonth:      params.DayOfMonth,
		SpecificDates:   params.Dates,
		EvenOdd:         params.EvenOdd,
		StartDate:       startDate,
		EndDate:         endDate,
		TaskTitle:       strings.TrimSpace(title),
		TaskDescription: strings.TrimSpace(description),
	}
}

func buildTasks(rule *recurrencedomain.RecurrenceRule, dates []time.Time, now time.Time) []taskdomain.Task {
	tasks := make([]taskdomain.Task, 0, len(dates))
	for _, d := range dates {
		d := d
		tasks = append(tasks, taskdomain.Task{
			Title:       rule.TaskTitle,
			Description: rule.TaskDescription,
			Status:      taskdomain.StatusNew,
			ScheduledDate: &d,
			CreatedAt:   now,
			UpdatedAt:   now,
		})
	}
	return tasks
}
