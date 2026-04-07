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
		Title:             normalized.Title,
		Description:       normalized.Description,
		Status:            normalized.Status,
		PeriodicityType:   normalized.PeriodicityType,
		PeriodicityConfig: taskdomain.PeriodicityConfig{},
	}

	// Установить конфиг периодичности если задана
	if normalized.PeriodicityConfig != nil {
		model.PeriodicityConfig = *normalized.PeriodicityConfig
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
		ID:                id,
		Title:             normalized.Title,
		Description:       normalized.Description,
		Status:            normalized.Status,
		PeriodicityType:   normalized.PeriodicityType,
		PeriodicityConfig: taskdomain.PeriodicityConfig{},
		UpdatedAt:         s.now(),
	}

	// Установить конфиг периодичности если задана
	if normalized.PeriodicityConfig != nil {
		model.PeriodicityConfig = *normalized.PeriodicityConfig
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

	// Валидация периодичности
	if input.PeriodicityType == "" {
		input.PeriodicityType = taskdomain.PeriodicityNone
	}

	if !input.PeriodicityType.Valid() {
		return CreateInput{}, fmt.Errorf("%w: invalid periodicity type", ErrInvalidInput)
	}

	if input.PeriodicityType != taskdomain.PeriodicityNone {
		if input.PeriodicityConfig == nil {
			return CreateInput{}, fmt.Errorf("%w: periodicity_config is required when periodicity_type is set", ErrInvalidInput)
		}

		if err := validatePeriodicityConfig(input.PeriodicityType, *input.PeriodicityConfig); err != nil {
			return CreateInput{}, fmt.Errorf("%w: %v", ErrInvalidInput, err)
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

	// Валидация периодичности
	if input.PeriodicityType == "" {
		input.PeriodicityType = taskdomain.PeriodicityNone
	}

	if !input.PeriodicityType.Valid() {
		return UpdateInput{}, fmt.Errorf("%w: invalid periodicity type", ErrInvalidInput)
	}

	if input.PeriodicityType != taskdomain.PeriodicityNone {
		if input.PeriodicityConfig == nil {
			return UpdateInput{}, fmt.Errorf("%w: periodicity_config is required when periodicity_type is set", ErrInvalidInput)
		}

		if err := validatePeriodicityConfig(input.PeriodicityType, *input.PeriodicityConfig); err != nil {
			return UpdateInput{}, fmt.Errorf("%w: %v", ErrInvalidInput, err)
		}
	}

	return input, nil
}

// validatePeriodicityConfig проверяет корректность параметров периодичности
func validatePeriodicityConfig(pType taskdomain.PeriodicityType, config taskdomain.PeriodicityConfig) error {
	switch pType {
	case taskdomain.PeriodicityDaily:
		if config.Interval <= 0 {
			return fmt.Errorf("interval must be positive for daily periodicity")
		}
		if config.StartDate.IsZero() {
			return fmt.Errorf("start_date is required for daily periodicity")
		}

	case taskdomain.PeriodicityMonthly:
		if config.MonthDay < 1 || config.MonthDay > 30 {
			return fmt.Errorf("month_day must be between 1 and 30 for monthly periodicity")
		}
		if config.StartDate.IsZero() {
			return fmt.Errorf("start_date is required for monthly periodicity")
		}

	case taskdomain.PeriodicitySpecificDate:
		if len(config.SpecificDates) == 0 {
			return fmt.Errorf("specific_dates cannot be empty")
		}
		// Проверить формат дат
		for _, dateStr := range config.SpecificDates {
			if _, err := time.Parse("2006-01-02", dateStr); err != nil {
				return fmt.Errorf("invalid date format in specific_dates: %s (expected YYYY-MM-DD)", dateStr)
			}
		}

	case taskdomain.PeriodicityEvenDays, taskdomain.PeriodicityOddDays:
		if config.StartDate.IsZero() {
			return fmt.Errorf("start_date is required for even/odd days periodicity")
		}

	case taskdomain.PeriodicityNone:
		// Нет требований к конфигу
	}

	return nil
}
