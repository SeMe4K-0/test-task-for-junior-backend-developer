package schedule

import (
	"context"
	"encoding/json"
	scheduledomain "example.com/taskservice/internal/domain/schedule"
	"fmt"
	"strings"
	"time"
)

type Service struct {
	repo Repository
	now  func() time.Time
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
		now: func() time.Time {
			return time.Now().UTC()
		},
	}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (*scheduledomain.Schedule, error) {
	normalized, err := validateCreateInput(input)
	if err != nil {
		return nil, err
	}

	model := &scheduledomain.Schedule{
		Title:            normalized.Title,
		Description:      normalized.Description,
		RecurrenceType:   normalized.RecurrenceType,
		RecurrenceParams: normalized.RecurrenceParams,
	}

	now := s.now()
	model.CreatedAt = now
	model.UpdatedAt = now
	model.IsActive = true

	created, err := s.repo.Create(ctx, model)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (s *Service) GetByID(ctx context.Context, id int64) (*scheduledomain.Schedule, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.repo.GetByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, id int64, input UpdateInput) (*scheduledomain.Schedule, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	normalized, err := validateUpdateInput(input)
	if err != nil {
		return nil, err
	}

	model := &scheduledomain.Schedule{
		ID:               id,
		Title:            normalized.Title,
		Description:      normalized.Description,
		RecurrenceType:   normalized.RecurrenceType,
		RecurrenceParams: normalized.RecurrenceParams,
		IsActive:         normalized.IsActive,
		UpdatedAt:        s.now(),
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

func (s *Service) List(ctx context.Context) ([]scheduledomain.Schedule, error) {
	return s.repo.List(ctx)
}

func validateCreateInput(input CreateInput) (CreateInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return CreateInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if !input.RecurrenceType.Valid() {
		return CreateInput{}, fmt.Errorf("%w: invalid recurrence type", ErrInvalidInput)
	}

	if err := validateRecurrenceParams(input.RecurrenceType, input.RecurrenceParams); err != nil {
		return CreateInput{}, err
	}

	return input, nil
}

func validateUpdateInput(input UpdateInput) (UpdateInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return UpdateInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if !input.RecurrenceType.Valid() {
		return UpdateInput{}, fmt.Errorf("%w: invalid recurrence type", ErrInvalidInput)
	}

	if err := validateRecurrenceParams(input.RecurrenceType, input.RecurrenceParams); err != nil {
		return UpdateInput{}, err
	}

	return input, nil
}

func validateRecurrenceParams(t scheduledomain.RecurrenceType, raw json.RawMessage) error {
	switch t {
	case scheduledomain.DailyType:
		var p struct {
			Interval int `json:"interval"`
		}
		if err := json.Unmarshal(raw, &p); err != nil {
			return fmt.Errorf("%w: invalid daily params: %v", ErrInvalidInput, err)
		}
		if p.Interval < 1 {
			return fmt.Errorf("%w: daily interval must be >= 1", ErrInvalidInput)
		}

	case scheduledomain.MonthlyType:
		var p struct {
			Day int `json:"day"`
		}
		if err := json.Unmarshal(raw, &p); err != nil {
			return fmt.Errorf("%w: invalid monthly params: %v", ErrInvalidInput, err)
		}
		if p.Day < 1 || p.Day > 30 {
			return fmt.Errorf("%w: monthly day must be between 1 and 30", ErrInvalidInput)
		}

	case scheduledomain.CustomDatesType:
		var p struct {
			Dates []string `json:"dates"`
		}
		if err := json.Unmarshal(raw, &p); err != nil {
			return fmt.Errorf("%w: invalid custom_dates params: %v", ErrInvalidInput, err)
		}
		if len(p.Dates) == 0 {
			return fmt.Errorf("%w: custom_dates list cannot be empty", ErrInvalidInput)
		}
		for _, d := range p.Dates {
			if _, err := time.Parse("2006-01-02", d); err != nil {
				return fmt.Errorf("%w: invalid date %q, must be YYYY-MM-DD", ErrInvalidInput, d)
			}
		}

	case scheduledomain.EvenType, scheduledomain.OddType, scheduledomain.EndOfMonthType:
		// параметры не требуются
	}

	return nil
}
