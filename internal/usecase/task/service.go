package task

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"

	taskdomain "example.com/taskservice/internal/domain/task"
	"example.com/taskservice/internal/infrastructure/metrics"
)

type Service struct {
	repo    Repository
	logger  Logger
	now     func() time.Time
}

func NewService(repo Repository, logger Logger) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
		now:    func() time.Time { return time.Now().UTC() },
	}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (*taskdomain.Task, error) {
	s.logger.Info("Creating new task",
		zap.String("title", input.Title),
		zap.String("status", string(input.Status)))

	normalized, err := validateCreateInput(input)
	if err != nil {
		s.logger.Error("Failed to validate create input",
			zap.Error(err),
			zap.String("title", input.Title))
		metrics.RecordError("usecase", "create", "validation")
		return nil, err
	}

	model := &taskdomain.Task{
		Title:       normalized.Title,
		Description: normalized.Description,
		Status:      normalized.Status,
		DueDate:     normalized.DueDate,
		Recurrence:  normalized.Recurrence,
	}
	now := s.now()
	model.CreatedAt = now
	model.UpdatedAt = now

	created, err := s.repo.Create(ctx, model)
	if err != nil {
		s.logger.Error("Failed to create task in repository",
			zap.Error(err),
			zap.String("title", model.Title))
		metrics.RecordError("usecase", "create", "repository")
		return nil, err
	}

	s.logger.Info("Task created successfully",
		zap.Int64("task_id", created.ID),
		zap.String("title", created.Title))
	metrics.RecordTaskOperation("create")

	created.Recurrence = model.Recurrence
	return created, nil
}

func (s *Service) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	s.logger.Debug("Getting task by ID", zap.Int64("task_id", id))

	if id <= 0 {
		s.logger.Warn("Invalid task ID provided",
			zap.Int64("task_id", id))
		metrics.RecordError("usecase", "get_by_id", "validation")
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	task, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get task by ID",
			zap.Error(err),
			zap.Int64("task_id", id))
		metrics.RecordError("usecase", "get_by_id", "repository")
		return nil, err
	}

	s.logger.Debug("Task retrieved successfully",
		zap.Int64("task_id", task.ID),
		zap.String("title", task.Title))
	return task, nil
}

func (s *Service) Update(ctx context.Context, id int64, input UpdateInput) (*taskdomain.Task, error) {
	s.logger.Info("Updating task",
		zap.Int64("task_id", id),
		zap.String("new_title", input.Title),
		zap.String("new_status", string(input.Status)))

	if id <= 0 {
		s.logger.Warn("Invalid task ID for update",
			zap.Int64("task_id", id))
		metrics.RecordError("usecase", "update", "validation")
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	normalized, err := validateUpdateInput(input)
	if err != nil {
		s.logger.Error("Failed to validate update input",
			zap.Error(err),
			zap.Int64("task_id", id))
		metrics.RecordError("usecase", "update", "validation")
		return nil, err
	}

	model := &taskdomain.Task{
		ID:          id,
		Title:       normalized.Title,
		Description: normalized.Description,
		Status:      normalized.Status,
		DueDate:     normalized.DueDate,
		Recurrence:  normalized.Recurrence,
		UpdatedAt:   s.now(),
	}

	updated, err := s.repo.Update(ctx, model)
	if err != nil {
		s.logger.Error("Failed to update task in repository",
			zap.Error(err),
			zap.Int64("task_id", id))
		metrics.RecordError("usecase", "update", "repository")
		return nil, err
	}

	s.logger.Info("Task updated successfully",
		zap.Int64("task_id", updated.ID),
		zap.String("title", updated.Title))
	metrics.RecordTaskOperation("update")

	updated.Recurrence = model.Recurrence
	return updated, nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	s.logger.Info("Deleting task", zap.Int64("task_id", id))

	if id <= 0 {
		s.logger.Warn("Invalid task ID for deletion",
			zap.Int64("task_id", id))
		metrics.RecordError("usecase", "delete", "validation")
		return fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	err := s.repo.Delete(ctx, id)
	if err != nil {
		s.logger.Error("Failed to delete task",
			zap.Error(err),
			zap.Int64("task_id", id))
		metrics.RecordError("usecase", "delete", "repository")
		return err
	}

	s.logger.Info("Task deleted successfully", zap.Int64("task_id", id))
	metrics.RecordTaskOperation("delete")

	return nil
}

func (s *Service) List(ctx context.Context) ([]taskdomain.Task, error) {
	s.logger.Debug("Listing all tasks")

	tasks, err := s.repo.List(ctx)
	if err != nil {
		s.logger.Error("Failed to list tasks",
			zap.Error(err))
		metrics.RecordError("usecase", "list", "repository")
		return nil, err
	}

	s.logger.Debug("Tasks listed successfully",
		zap.Int("count", len(tasks)))
	return tasks, nil
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

	normalizedRecurrence, err := normalizeRecurrence(input.Recurrence)
	if err != nil {
		return CreateInput{}, err
	}
	input.Recurrence = normalizedRecurrence

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

	normalizedRecurrence, err := normalizeRecurrence(input.Recurrence)
	if err != nil {
		return UpdateInput{}, err
	}
	input.Recurrence = normalizedRecurrence

	return input, nil
}

func normalizeRecurrence(recurrence *taskdomain.Recurrence) (*taskdomain.Recurrence, error) {
	if recurrence == nil {
		return nil, nil
	}

	if !recurrence.Type.Valid() {
		return nil, fmt.Errorf("%w: invalid recurrence type", ErrInvalidInput)
	}

	normalized := &taskdomain.Recurrence{Type: recurrence.Type}

	switch recurrence.Type {
	case taskdomain.RecurrenceTypeDaily:
		if recurrence.IntervalDays <= 0 {
			return nil, fmt.Errorf("%w: interval_days is required for daily recurrence", ErrInvalidInput)
		}
		normalized.IntervalDays = recurrence.IntervalDays

	case taskdomain.RecurrenceTypeMonthly:
		if recurrence.MonthDay < 1 || recurrence.MonthDay > 30 {
			return nil, fmt.Errorf("%w: month_day must be between 1 and 30", ErrInvalidInput)
		}
		normalized.MonthDay = recurrence.MonthDay

	case taskdomain.RecurrenceTypeSpecificDates:
		if len(recurrence.SpecificDates) == 0 {
			return nil, fmt.Errorf("%w: specific_dates are required for specific_dates recurrence", ErrInvalidInput)
		}
		for _, rawDate := range recurrence.SpecificDates {
			parsed, err := parseRecurrenceDate(rawDate)
			if err != nil {
				return nil, err
			}
			normalized.SpecificDates = append(normalized.SpecificDates, parsed)
		}

	case taskdomain.RecurrenceTypeParity:
		if !recurrence.Parity.Valid() {
			return nil, fmt.Errorf("%w: invalid parity value", ErrInvalidInput)
		}
		normalized.Parity = recurrence.Parity

	default:
		return nil, fmt.Errorf("%w: unsupported recurrence type", ErrInvalidInput)
	}

	return normalized, nil
}

func parseRecurrenceDate(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("%w: recurrence date is required", ErrInvalidInput)
	}

	if t, err := time.Parse("2006-01-02", value); err == nil {
		return t.Format("2006-01-02"), nil
	}

	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t.UTC().Format("2006-01-02"), nil
	}

	return "", fmt.Errorf("%w: invalid recurrence date format", ErrInvalidInput)
}
