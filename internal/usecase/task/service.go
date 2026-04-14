package task

import (
	"context"
	"fmt"
	"strings"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

const dateLayout = "2006-01-02"

type Service struct {
	repo           Repository
	recurrenceRepo RecurrenceRepository
	now            func() time.Time
}

func NewService(repo Repository, recurrenceRepo RecurrenceRepository) *Service {
	return &Service{
		repo:           repo,
		recurrenceRepo: recurrenceRepo,
		now:            func() time.Time { return time.Now().UTC() },
	}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (*taskdomain.Task, error) {
	normalized, err := validateCreateInput(input)
	if err != nil {
		return nil, err
	}

	now := s.now()

	model := &taskdomain.Task{
		Title:       normalized.Title,
		Description: normalized.Description,
		Status:      normalized.Status,
		CreatedAt:   now,
		UpdatedAt:   now,
		IsTemplate:  normalized.Recurrence != nil,
	}

	created, err := s.repo.Create(ctx, model)
	if err != nil {
		return nil, err
	}

	// Если есть настройки периодичности — создаём правило.
	if normalized.Recurrence != nil {
		rule, err := buildRecurrenceRule(created.ID, *normalized.Recurrence, now)
		if err != nil {
			return nil, err
		}

		if _, err := s.recurrenceRepo.Create(ctx, rule); err != nil {
			return nil, err
		}
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

func (s *Service) ListByDate(ctx context.Context, date string) ([]taskdomain.Task, error) {
	if _, err := time.Parse(dateLayout, date); err != nil {
		return nil, fmt.Errorf("%w: invalid date format, expected YYYY-MM-DD", ErrInvalidInput)
	}

	return s.repo.ListByScheduledDate(ctx, date)
}

func (s *Service) GetRecurrence(ctx context.Context, taskID int64) (*taskdomain.RecurrenceRule, error) {
	if taskID <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.recurrenceRepo.GetByTaskID(ctx, taskID)
}

func (s *Service) UpdateRecurrence(ctx context.Context, taskID int64, input RecurrenceInput) (*taskdomain.RecurrenceRule, error) {
	if taskID <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	// Проверяем, что задача существует и является шаблоном.
	task, err := s.repo.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}

	if !task.IsTemplate {
		return nil, fmt.Errorf("%w: task is not a template", ErrInvalidInput)
	}

	validatedInput, err := validateRecurrenceInput(input)
	if err != nil {
		return nil, err
	}

	rule, err := buildRecurrenceRule(taskID, validatedInput, s.now())
	if err != nil {
		return nil, err
	}

	updated, err := s.recurrenceRepo.Update(ctx, rule)
	if err != nil {
		return nil, err
	}

	// Удаляем будущие непроделанные экземпляры — scheduler пересоздаст с новым правилом.
	today := s.now().Format(dateLayout)
	if err := s.repo.DeleteFutureInstancesByParent(ctx, taskID, today); err != nil {
		return nil, err
	}

	return updated, nil
}

func (s *Service) DeleteRecurrence(ctx context.Context, taskID int64) error {
	if taskID <= 0 {
		return fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.recurrenceRepo.Delete(ctx, taskID)
}

// --- validation ---

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

	if input.Recurrence != nil {
		validated, err := validateRecurrenceInput(*input.Recurrence)
		if err != nil {
			return CreateInput{}, err
		}

		input.Recurrence = &validated
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

func validateRecurrenceInput(input RecurrenceInput) (RecurrenceInput, error) {
	if !input.Type.Valid() {
		return RecurrenceInput{}, fmt.Errorf("%w: invalid recurrence type", ErrInvalidInput)
	}

	switch input.Type {
	case taskdomain.RecurrenceDaily:
		if input.IntervalDays != nil && *input.IntervalDays <= 0 {
			return RecurrenceInput{}, fmt.Errorf("%w: interval_days must be positive", ErrInvalidInput)
		}

	case taskdomain.RecurrenceMonthly:
		if input.DayOfMonth == nil {
			return RecurrenceInput{}, fmt.Errorf("%w: day_of_month is required for monthly recurrence", ErrInvalidInput)
		}

		if *input.DayOfMonth < 1 || *input.DayOfMonth > 30 {
			return RecurrenceInput{}, fmt.Errorf("%w: day_of_month must be between 1 and 30", ErrInvalidInput)
		}

	case taskdomain.RecurrenceSpecificDates:
		if len(input.SpecificDates) == 0 {
			return RecurrenceInput{}, fmt.Errorf("%w: specific_dates cannot be empty", ErrInvalidInput)
		}

		for _, d := range input.SpecificDates {
			if _, err := time.Parse(dateLayout, d); err != nil {
				return RecurrenceInput{}, fmt.Errorf("%w: invalid date format %q, expected YYYY-MM-DD", ErrInvalidInput, d)
			}
		}

	case taskdomain.RecurrenceEvenOdd:
		if input.EvenOdd == nil {
			return RecurrenceInput{}, fmt.Errorf("%w: even_odd is required for even_odd recurrence", ErrInvalidInput)
		}

		eo := taskdomain.EvenOdd(*input.EvenOdd)
		if !eo.Valid() {
			return RecurrenceInput{}, fmt.Errorf("%w: even_odd must be 'even' or 'odd'", ErrInvalidInput)
		}
	}

	if input.StartDate == "" {
		return RecurrenceInput{}, fmt.Errorf("%w: start_date is required", ErrInvalidInput)
	}

	if _, err := time.Parse(dateLayout, input.StartDate); err != nil {
		return RecurrenceInput{}, fmt.Errorf("%w: invalid start_date format, expected YYYY-MM-DD", ErrInvalidInput)
	}

	if input.EndDate != nil {
		if _, err := time.Parse(dateLayout, *input.EndDate); err != nil {
			return RecurrenceInput{}, fmt.Errorf("%w: invalid end_date format, expected YYYY-MM-DD", ErrInvalidInput)
		}
	}

	return input, nil
}

func buildRecurrenceRule(taskID int64, input RecurrenceInput, now time.Time) (*taskdomain.RecurrenceRule, error) {
	startDate, _ := time.Parse(dateLayout, input.StartDate)

	rule := &taskdomain.RecurrenceRule{
		TaskID:       taskID,
		Type:         input.Type,
		IntervalDays: input.IntervalDays,
		DayOfMonth:   input.DayOfMonth,
		EvenOdd:      nil,
		StartDate:    startDate,
		CreatedAt:    now,
	}

	if input.EvenOdd != nil {
		eo := taskdomain.EvenOdd(*input.EvenOdd)
		rule.EvenOdd = &eo
	}

	if input.EndDate != nil {
		endDate, _ := time.Parse(dateLayout, *input.EndDate)
		rule.EndDate = &endDate
	}

	// Конвертируем строковые даты в time.Time.
	for _, ds := range input.SpecificDates {
		d, _ := time.Parse(dateLayout, ds)
		rule.SpecificDates = append(rule.SpecificDates, d)
	}

	return rule, nil
}
