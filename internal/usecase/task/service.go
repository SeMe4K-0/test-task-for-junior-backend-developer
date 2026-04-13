package task

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"
	"encoding/json"

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

	if normalized.Recurrence != nil {
		rule := &taskdomain.TaskRule{
			Title:           normalized.Title,
			Description:     normalized.Description,
			RecurrenceType:  normalized.Recurrence.Type,
			RecurrenceValue: normalized.Recurrence.Value,
			StartDate:       normalized.Recurrence.StartDate,
			EndDate:         normalized.Recurrence.EndDate,
			CreatedAt:       now,
		}
		createdRule, err := s.repo.CreateRule(ctx, rule)
		if err != nil {
			return nil, err
		}
		
		firstDate := firstOccurrence(createdRule) //createdRule.StartDate

		return &taskdomain.Task{
			ID:            0, 
			Title:         createdRule.Title,
			Description:   createdRule.Description,
			Status:        taskdomain.StatusNew,
			ExecutionDate: firstDate,
			RuleID:        &createdRule.ID,
			CreatedAt:     createdRule.CreatedAt, 
			UpdatedAt:     createdRule.CreatedAt, 
		}, nil
	}

	model := &taskdomain.Task{
		Title:         normalized.Title,
		Description:   normalized.Description,
		Status:        normalized.Status,
		ExecutionDate: normalized.ExecutionDate,
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
		ID:            id,
		Title:         normalized.Title,
		Description:   normalized.Description,
		Status:        normalized.Status,
		ExecutionDate: normalized.ExecutionDate,
		UpdatedAt:     s.now(),
	}

	return s.repo.Update(ctx, model)
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.repo.Delete(ctx, id)
}

func (s *Service) List(ctx context.Context, from, to time.Time) ([]taskdomain.Task, error) {
	realTasks, err := s.repo.List(ctx, from, to)
	if err != nil {
		return nil, err
	}

	rules, err := s.repo.ListActiveRules(ctx, from, to)
	if err != nil {
		return nil, err
	}

	materializedMap := make(map[string]bool)
	for _, t := range realTasks {
		if t.RuleID != nil {
			key := fmt.Sprintf("%d_%s", *t.RuleID, t.ExecutionDate.Format("2006-01-02"))
			materializedMap[key] = true
		}
	}

	result := make([]taskdomain.Task, 0, len(realTasks))
	result = append(result, realTasks...)

	for _, rule := range rules {
		dates := taskdomain.GenerateDatesForRule(rule, from, to)
		
		for _, d := range dates {
			key := fmt.Sprintf("%d_%s", rule.ID, d.Format("2006-01-02"))
			
			if !materializedMap[key] {
				rID := rule.ID
				result = append(result, taskdomain.Task{
					ID:            0, 
					Title:         rule.Title,
					Description:   rule.Description,
					Status:        taskdomain.StatusNew,
					ExecutionDate: d,
					RuleID:        &rID,
					CreatedAt:     rule.CreatedAt, 
					UpdatedAt:     rule.CreatedAt,
				})
			}
		}
	}

	sort.Slice(result, func(i, j int) bool {
	    if result[i].ExecutionDate.Equal(result[j].ExecutionDate) {
	        return result[i].ID < result[j].ID // вторичный ключ
	    }
	    return result[i].ExecutionDate.Before(result[j].ExecutionDate)
	})

	return result, nil
}

func (s *Service) Materialize(ctx context.Context, input MaterializeInput) (*taskdomain.Task, error) {
	if input.RuleID <= 0 {
		return nil, fmt.Errorf("%w: rule_id is required", ErrInvalidInput)
	}

	if input.ExecutionDate.IsZero() {
		return nil, fmt.Errorf("%w: execution_date is required", ErrInvalidInput)
	}

	normalizedTitle := strings.TrimSpace(input.Title)
	if normalizedTitle == "" {
		return nil, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if !input.Status.Valid() {
		return nil, fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}

	rule, err := s.repo.GetRuleByID(ctx, input.RuleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get rule: %w", err)
	}

	targetDate := taskdomain.MidnightUTC(input.ExecutionDate)
	validDates := taskdomain.GenerateDatesForRule(*rule, targetDate, targetDate)
	
	if len(validDates) == 0 {
		return nil, fmt.Errorf("%w: execution_date %s does not match schedule for rule %d", 
			ErrInvalidInput, targetDate.Format("2006-01-02"), rule.ID)
	}

	now := s.now()
	model := &taskdomain.Task{
		Title:         normalizedTitle,
		Description:   strings.TrimSpace(input.Description),
		Status:        input.Status,
		ExecutionDate: targetDate, 
		RuleID:        &input.RuleID,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	return s.repo.Create(ctx, model)
}

func (s *Service) ListRules(ctx context.Context) ([]taskdomain.TaskRule, error) {
	return s.repo.ListRules(ctx)
}

func (s *Service) DeleteRule(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}
	return s.repo.DeleteRule(ctx, id)
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

	if input.Recurrence == nil && input.ExecutionDate.IsZero() {
		return CreateInput{}, fmt.Errorf("%w: execution_date is required", ErrInvalidInput)
	}

	if input.Recurrence != nil {
		if !input.Recurrence.Type.Valid() {
			return CreateInput{}, fmt.Errorf("%w: invalid recurrence type", ErrInvalidInput)
		}

		if input.Recurrence.StartDate.IsZero() {
			return CreateInput{}, fmt.Errorf("%w: recurrence start_date is required", ErrInvalidInput)
		}

		if input.Recurrence.EndDate != nil && input.Recurrence.EndDate.Before(input.Recurrence.StartDate) {
			return CreateInput{}, fmt.Errorf("%w: end_date cannot be before start_date", ErrInvalidInput)
		}

		switch input.Recurrence.Type {
		case taskdomain.RecurrenceDaily:
			var p taskdomain.DailyRecurrence
			if err := json.Unmarshal(input.Recurrence.Value, &p); err != nil {
				return CreateInput{}, fmt.Errorf("%w: invalid JSON for daily recurrence", ErrInvalidInput)
			}
			if p.Interval <= 0 {
				return CreateInput{}, fmt.Errorf("%w: daily interval must be > 0", ErrInvalidInput)
			}
		case taskdomain.RecurrenceMonthly:
			var p taskdomain.MonthlyRecurrence
			if err := json.Unmarshal(input.Recurrence.Value, &p); err != nil {
				return CreateInput{}, fmt.Errorf("%w: invalid JSON for monthly recurrence", ErrInvalidInput)
			}
			if len(p.Days) == 0 {
				return CreateInput{}, fmt.Errorf("%w: monthly recurrence requires at least one day", ErrInvalidInput)
			}
			for _, day := range p.Days {
				if day < 1 || day > 30 {
					return CreateInput{}, fmt.Errorf("%w: monthly day must be between 1 and 30", ErrInvalidInput)
				}
			}
		case taskdomain.RecurrenceSpecificDates:
			var p taskdomain.SpecificDatesRecurrence
			if err := json.Unmarshal(input.Recurrence.Value, &p); err != nil {
				return CreateInput{}, fmt.Errorf("%w: invalid JSON for specific_dates recurrence", ErrInvalidInput)
			}
			if len(p.Dates) == 0 {
				return CreateInput{}, fmt.Errorf("%w: specific_dates requires at least one date", ErrInvalidInput)
			}
			seen := make(map[string]bool)
			for _, dStr := range p.Dates {
				if _, err := time.Parse("2006-01-02", dStr); err != nil {
					return CreateInput{}, fmt.Errorf("%w: specific date must be in YYYY-MM-DD format", ErrInvalidInput)
				}
				if seen[dStr] {
		            return CreateInput{}, fmt.Errorf("%w: duplicate date %s", ErrInvalidInput, dStr)
		        }
		        seen[dStr] = true
			}
		case taskdomain.RecurrenceEvenDays, taskdomain.RecurrenceOddDays:
			input.Recurrence.Value = []byte("{}")
		}
	}

	return input, nil
}

func firstOccurrence(rule *taskdomain.TaskRule) time.Time {
    window := rule.StartDate.AddDate(1, 0, 0)
    if rule.EndDate != nil && rule.EndDate.Before(window) {
        window = *rule.EndDate
    }

    dates := taskdomain.GenerateDatesForRule(*rule, rule.StartDate, window)
    if len(dates) > 0 {
        return dates[0]
    }

    return rule.StartDate
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

	if input.ExecutionDate.IsZero() {
        return UpdateInput{}, fmt.Errorf("%w: execution_date is required", ErrInvalidInput)
    }

	return input, nil
}
