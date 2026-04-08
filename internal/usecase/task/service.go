package task

import (
	"context"
	"fmt"
	"reflect"
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

func (s *Service) Create(ctx context.Context, input CreateTaskInput) (*taskdomain.Task, error) {
	normalized, err := validateCreateTaskInput(input)
	if err != nil {
		return nil, err
	}

	now := s.now()
	template := &taskdomain.Template{
		Title:       normalized.Title,
		Description: normalized.Description,
		Schedule:    normalized.Schedule,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	firstOccurrenceOn, err := firstOccurrenceForCreate(template, now)
	if err != nil {
		return nil, err
	}

	createdTemplate, err := s.repo.CreateTemplate(ctx, template)
	if err != nil {
		return nil, err
	}

	if err := s.syncTemplateOccurrences(ctx, createdTemplate, *firstOccurrenceOn, firstOccurrenceOn.AddDays(DefaultSeedWindowDays), &normalized.Status, firstOccurrenceOn); err != nil {
		_ = s.repo.DeleteTemplate(ctx, createdTemplate.ID)
		return nil, err
	}

	task, err := s.repo.GetOccurrenceByTemplateAndDate(ctx, createdTemplate.ID, *firstOccurrenceOn)
	if err != nil {
		return nil, err
	}
	task.Schedule = createdTemplate.Schedule

	return task, nil
}

func (s *Service) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.repo.GetOccurrenceByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, id int64, input UpdateTaskInput) (*taskdomain.Task, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	normalized, err := validateUpdateTaskInput(input)
	if err != nil {
		return nil, err
	}

	current, err := s.repo.GetOccurrenceByID(ctx, id)
	if err != nil {
		return nil, err
	}

	now := s.now()
	updatedOccurrence := &taskdomain.Task{
		ID:          id,
		Title:       normalized.Title,
		Description: normalized.Description,
		Status:      normalized.Status,
		UpdatedAt:   now,
		CompletedAt: completedAtForStatus(normalized.Status, now),
	}

	if current.Schedule != nil && normalized.Schedule != nil && !reflect.DeepEqual(current.Schedule, normalized.Schedule) {
		return nil, fmt.Errorf("%w: recurring schedule can only be changed via template endpoint", ErrInvalidInput)
	}

	updated, err := s.repo.UpdateOccurrence(ctx, updatedOccurrence)
	if err != nil {
		return nil, err
	}

	updated.Schedule = current.Schedule

	if current.Schedule == nil {
		template := &taskdomain.Template{
			ID:          current.TemplateID,
			Title:       normalized.Title,
			Description: normalized.Description,
			Schedule:    normalized.Schedule,
			UpdatedAt:   now,
		}

		updatedTemplate, err := s.repo.UpdateTemplate(ctx, template)
		if err != nil {
			return nil, err
		}

		updated.Schedule = updatedTemplate.Schedule
		if normalized.Schedule != nil {
			syncFrom := maxDate(updated.ScheduledFor.AddDays(1), updatedTemplate.Schedule.DateFromTime(now))
			if err := s.syncTemplateOccurrences(ctx, updatedTemplate, syncFrom, syncFrom.AddDays(DefaultSeedWindowDays), nil, nil); err != nil {
				return nil, err
			}
		}
	}

	return updated, nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	current, err := s.repo.GetOccurrenceByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.repo.DeleteOccurrence(ctx, id); err != nil {
		return err
	}

	count, err := s.repo.CountOccurrencesByTemplate(ctx, current.TemplateID)
	if err != nil {
		return err
	}

	if count == 0 {
		return s.repo.DeleteTemplate(ctx, current.TemplateID)
	}

	return nil
}

func (s *Service) List(ctx context.Context, input TaskListInput) (TaskListResult, error) {
	normalized, err := validateTaskListInput(input, s.now())
	if err != nil {
		return TaskListResult{}, err
	}

	syncFrom, syncTo := resolveTaskSyncWindow(normalized, s.now())
	templates, err := s.repo.ListTemplatesForSync(ctx, normalized.TemplateID)
	if err != nil {
		return TaskListResult{}, err
	}

	for i := range templates {
		if err := s.syncTemplateOccurrences(ctx, &templates[i], syncFrom, syncTo, nil, nil); err != nil {
			return TaskListResult{}, err
		}
	}

	items, total, err := s.repo.ListOccurrences(ctx, normalized)
	if err != nil {
		return TaskListResult{}, err
	}

	return TaskListResult{
		Items:  items,
		Total:  total,
		Limit:  normalized.Limit,
		Offset: normalized.Offset,
	}, nil
}

func (s *Service) CreateTemplate(ctx context.Context, input CreateTemplateInput) (*taskdomain.Template, error) {
	normalized, err := validateCreateTemplateInput(input)
	if err != nil {
		return nil, err
	}

	now := s.now()
	template := &taskdomain.Template{
		Title:       normalized.Title,
		Description: normalized.Description,
		Schedule:    normalized.Schedule,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	created, err := s.repo.CreateTemplate(ctx, template)
	if err != nil {
		return nil, err
	}

	seedFrom, seedTo := templateSeedWindow(created, now)
	if err := s.syncTemplateOccurrences(ctx, created, seedFrom, seedTo, nil, nil); err != nil {
		_ = s.repo.DeleteTemplate(ctx, created.ID)
		return nil, err
	}

	created.NextOccurrenceOn = nextOccurrenceForTemplate(created, now)

	return created, nil
}

func (s *Service) GetTemplateByID(ctx context.Context, id int64) (*taskdomain.Template, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	template, err := s.repo.GetTemplateByID(ctx, id)
	if err != nil {
		return nil, err
	}

	template.NextOccurrenceOn = nextOccurrenceForTemplate(template, s.now())

	return template, nil
}

func (s *Service) UpdateTemplate(ctx context.Context, id int64, input UpdateTemplateInput) (*taskdomain.Template, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	normalized, err := validateUpdateTemplateInput(input)
	if err != nil {
		return nil, err
	}

	current, err := s.repo.GetTemplateByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if current.Schedule != nil && normalized.Schedule == nil {
		return nil, fmt.Errorf("%w: recurring template cannot be converted to one-off", ErrInvalidInput)
	}

	now := s.now()
	template := &taskdomain.Template{
		ID:          id,
		Title:       normalized.Title,
		Description: normalized.Description,
		Schedule:    normalized.Schedule,
		UpdatedAt:   now,
	}

	updated, err := s.repo.UpdateTemplate(ctx, template)
	if err != nil {
		return nil, err
	}

	switch {
	case current.Schedule == nil && updated.Schedule == nil:
		occurrenceDate := oneOffOccurrenceDate(current)
		task, getErr := s.repo.GetOccurrenceByTemplateAndDate(ctx, current.ID, occurrenceDate)
		if getErr == nil {
			_, err = s.repo.UpdateOccurrence(ctx, &taskdomain.Task{
				ID:          task.ID,
				Title:       updated.Title,
				Description: updated.Description,
				Status:      task.Status,
				UpdatedAt:   now,
				CompletedAt: task.CompletedAt,
			})
			if err != nil {
				return nil, err
			}
		}
	case current.Schedule == nil && updated.Schedule != nil:
		syncFrom := maxDate(oneOffOccurrenceDate(current).AddDays(1), updated.Schedule.DateFromTime(now))
		if err := s.syncTemplateOccurrences(ctx, updated, syncFrom, syncFrom.AddDays(DefaultSeedWindowDays), nil, nil); err != nil {
			return nil, err
		}
	default:
		deleteFrom := updated.Schedule.DateFromTime(now)
		if err := s.repo.DeleteOccurrencesFrom(ctx, updated.ID, deleteFrom); err != nil {
			return nil, err
		}
		if err := s.syncTemplateOccurrences(ctx, updated, deleteFrom, deleteFrom.AddDays(DefaultSeedWindowDays), nil, nil); err != nil {
			return nil, err
		}
	}

	updated.NextOccurrenceOn = nextOccurrenceForTemplate(updated, now)

	return updated, nil
}

func (s *Service) DeleteTemplate(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.repo.DeleteTemplate(ctx, id)
}

func (s *Service) ListTemplates(ctx context.Context, input TemplateListInput) (TemplateListResult, error) {
	normalized, err := validateTemplateListInput(input)
	if err != nil {
		return TemplateListResult{}, err
	}

	items, total, err := s.repo.ListTemplates(ctx, normalized)
	if err != nil {
		return TemplateListResult{}, err
	}

	now := s.now()
	for i := range items {
		items[i].NextOccurrenceOn = nextOccurrenceForTemplate(&items[i], now)
	}

	return TemplateListResult{
		Items:  items,
		Total:  total,
		Limit:  normalized.Limit,
		Offset: normalized.Offset,
	}, nil
}

func validateCreateTaskInput(input CreateTaskInput) (CreateTaskInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return CreateTaskInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if input.Status == "" {
		input.Status = taskdomain.StatusNew
	}

	if !input.Status.Valid() {
		return CreateTaskInput{}, fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}

	if err := input.Schedule.Validate(); err != nil {
		return CreateTaskInput{}, fmt.Errorf("%w: schedule %v", ErrInvalidInput, err)
	}

	return input, nil
}

func validateUpdateTaskInput(input UpdateTaskInput) (UpdateTaskInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return UpdateTaskInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if !input.Status.Valid() {
		return UpdateTaskInput{}, fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}

	if err := input.Schedule.Validate(); err != nil {
		return UpdateTaskInput{}, fmt.Errorf("%w: schedule %v", ErrInvalidInput, err)
	}

	return input, nil
}

func validateTaskListInput(input TaskListInput, now time.Time) (TaskListInput, error) {
	if input.Limit <= 0 {
		input.Limit = DefaultPageLimit
	}
	if input.Limit > MaxPageLimit {
		return TaskListInput{}, fmt.Errorf("%w: limit must be between 1 and %d", ErrInvalidInput, MaxPageLimit)
	}
	if input.Offset < 0 {
		return TaskListInput{}, fmt.Errorf("%w: offset must be non-negative", ErrInvalidInput)
	}
	if input.Status != nil && !input.Status.Valid() {
		return TaskListInput{}, fmt.Errorf("%w: invalid status filter", ErrInvalidInput)
	}
	if input.TemplateID != nil && *input.TemplateID <= 0 {
		return TaskListInput{}, fmt.Errorf("%w: template_id must be positive", ErrInvalidInput)
	}

	switch {
	case input.DateFrom != nil && input.DateTo == nil:
		dateTo := input.DateFrom.AddDays(DefaultSeedWindowDays)
		input.DateTo = &dateTo
	case input.DateFrom == nil && input.DateTo != nil:
		dateFrom := input.DateTo.AddDays(-DefaultSeedWindowDays)
		input.DateFrom = &dateFrom
	case input.DateFrom == nil && input.DateTo == nil:
		_ = now
	}

	if input.DateFrom != nil && input.DateTo != nil && input.DateTo.Before(input.DateFrom.Time) {
		return TaskListInput{}, fmt.Errorf("%w: date_from must be before or equal to date_to", ErrInvalidInput)
	}

	return input, nil
}

func validateCreateTemplateInput(input CreateTemplateInput) (CreateTemplateInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return CreateTemplateInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if err := input.Schedule.Validate(); err != nil {
		return CreateTemplateInput{}, fmt.Errorf("%w: schedule %v", ErrInvalidInput, err)
	}

	return input, nil
}

func validateUpdateTemplateInput(input UpdateTemplateInput) (UpdateTemplateInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return UpdateTemplateInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if err := input.Schedule.Validate(); err != nil {
		return UpdateTemplateInput{}, fmt.Errorf("%w: schedule %v", ErrInvalidInput, err)
	}

	return input, nil
}

func validateTemplateListInput(input TemplateListInput) (TemplateListInput, error) {
	input.Search = strings.TrimSpace(input.Search)

	if input.Limit <= 0 {
		input.Limit = DefaultPageLimit
	}
	if input.Limit > MaxPageLimit {
		return TemplateListInput{}, fmt.Errorf("%w: limit must be between 1 and %d", ErrInvalidInput, MaxPageLimit)
	}
	if input.Offset < 0 {
		return TemplateListInput{}, fmt.Errorf("%w: offset must be non-negative", ErrInvalidInput)
	}
	if input.ScheduleType != nil {
		switch *input.ScheduleType {
		case taskdomain.ScheduleDaily, taskdomain.ScheduleMonthly, taskdomain.ScheduleSpecific, taskdomain.ScheduleOddDays, taskdomain.ScheduleEvenDays:
		default:
			return TemplateListInput{}, fmt.Errorf("%w: invalid schedule_type filter", ErrInvalidInput)
		}
	}

	return input, nil
}

func (s *Service) syncTemplateOccurrences(ctx context.Context, template *taskdomain.Template, from, to taskdomain.Date, firstStatus *taskdomain.Status, firstDate *taskdomain.Date) error {
	if to.Before(from.Time) {
		return nil
	}

	occurrences := buildOccurrences(template, from, to, firstStatus, firstDate, s.now())
	if len(occurrences) == 0 {
		return nil
	}

	return s.repo.UpsertOccurrences(ctx, occurrences)
}

func buildOccurrences(template *taskdomain.Template, from, to taskdomain.Date, firstStatus *taskdomain.Status, firstDate *taskdomain.Date, now time.Time) []taskdomain.Task {
	if template == nil {
		return nil
	}

	dates := make([]taskdomain.Date, 0)
	switch {
	case template.Schedule == nil:
		oneOffDate := oneOffOccurrenceDate(template)
		if !oneOffDate.Before(from.Time) && !oneOffDate.After(to.Time) {
			dates = append(dates, oneOffDate)
		}
	default:
		dates = append(dates, template.Schedule.OccurrencesBetween(from, to, template.CreatedAt)...)
	}

	occurrences := make([]taskdomain.Task, 0, len(dates))
	for _, occurrenceDate := range dates {
		status := taskdomain.StatusNew
		if firstStatus != nil && firstDate != nil && occurrenceDate.Equal(firstDate.Time) {
			status = *firstStatus
		}

		occurrence := taskdomain.Task{
			TemplateID:   template.ID,
			Title:        template.Title,
			Description:  template.Description,
			Status:       status,
			ScheduledFor: occurrenceDate,
			Schedule:     template.Schedule,
			CreatedAt:    now,
			UpdatedAt:    now,
		}
		if status == taskdomain.StatusDone {
			completedAt := occurrence.UpdatedAt
			occurrence.CompletedAt = &completedAt
		}
		occurrences = append(occurrences, occurrence)
	}

	return occurrences
}

func oneOffOccurrenceDate(template *taskdomain.Template) taskdomain.Date {
	return taskdomain.NewDate(template.CreatedAt)
}

func firstOccurrenceForCreate(template *taskdomain.Template, now time.Time) (*taskdomain.Date, error) {
	if template.Schedule == nil {
		date := oneOffOccurrenceDate(template)
		return &date, nil
	}

	referenceDate := template.Schedule.DateFromTime(now)
	next := template.Schedule.NextOccurrence(referenceDate, template.CreatedAt)
	if next == nil {
		return nil, fmt.Errorf("%w: schedule has no upcoming occurrences", ErrInvalidInput)
	}

	return next, nil
}

func nextOccurrenceForTemplate(template *taskdomain.Template, now time.Time) *taskdomain.Date {
	if template.Schedule == nil {
		date := oneOffOccurrenceDate(template)
		if date.Before(taskdomain.NewDate(now).Time) {
			return nil
		}
		return &date
	}

	referenceDate := template.Schedule.DateFromTime(now)
	return template.Schedule.NextOccurrence(referenceDate, template.CreatedAt)
}

func templateSeedWindow(template *taskdomain.Template, now time.Time) (taskdomain.Date, taskdomain.Date) {
	if template.Schedule == nil {
		date := oneOffOccurrenceDate(template)
		return date, date
	}

	from := template.Schedule.DateFromTime(now)
	return from, from.AddDays(DefaultSeedWindowDays)
}

func resolveTaskSyncWindow(input TaskListInput, now time.Time) (taskdomain.Date, taskdomain.Date) {
	switch {
	case input.DateFrom != nil && input.DateTo != nil:
		return *input.DateFrom, *input.DateTo
	case input.DateFrom != nil:
		return *input.DateFrom, input.DateFrom.AddDays(DefaultSeedWindowDays)
	case input.DateTo != nil:
		return input.DateTo.AddDays(-DefaultSeedWindowDays), *input.DateTo
	default:
		today := taskdomain.NewDate(now)
		return today.AddDays(-DefaultSyncPastDays), today.AddDays(DefaultSeedWindowDays)
	}
}

func completedAtForStatus(status taskdomain.Status, now time.Time) *time.Time {
	if status != taskdomain.StatusDone {
		return nil
	}

	completedAt := now
	return &completedAt
}

func maxDate(a, b taskdomain.Date) taskdomain.Date {
	if a.Before(b.Time) {
		return b
	}

	return a
}
