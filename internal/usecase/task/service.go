package task

import (
	"context"
	"fmt"
	"strings"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

// materializationHorizon — горизонт опережающего создания экземпляров.
const materializationHorizon = 60 * 24 * time.Hour

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
	title, desc, status, err := validateInput(input.Title, input.Description, input.Status, true)
	if err != nil {
		return nil, err
	}

	now := s.now()
	model := &taskdomain.Task{
		Title:                   title,
		Description:             desc,
		Status:                  status,
		ScheduledAt:             input.ScheduledAt,
		RecurrenceType:          input.RecurrenceType,
		RecurrenceDailyInterval: input.RecurrenceDailyInterval,
		RecurrenceMonthlyDays:   input.RecurrenceMonthlyDays,
		RecurrenceSpecificDates: input.RecurrenceSpecificDates,
		RecurrenceDayParity: input.RecurrenceDayParity,
		CreatedAt:               now,
		UpdatedAt:               now,
	}

	if err := taskdomain.ValidateRecurrence(model); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidInput, err)
	}

	created, err := s.repo.Create(ctx, model)
	if err != nil {
		return nil, err
	}

	if created.IsTemplate() {
		if err := materialize(ctx, s.repo,created, now); err != nil {
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

	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	title, desc, status, err := validateInput(input.Title, input.Description, input.Status, false)
	if err != nil {
		return nil, err
	}

	now := s.now()
	wasTemplate := existing.IsTemplate()

	existing.Title = title
	existing.Description = desc
	existing.Status = status
	existing.ScheduledAt = input.ScheduledAt
	// PUT: nil снимает правило
	existing.RecurrenceType = input.RecurrenceType
	existing.RecurrenceDailyInterval = input.RecurrenceDailyInterval
	existing.RecurrenceMonthlyDays = input.RecurrenceMonthlyDays
	existing.RecurrenceSpecificDates = input.RecurrenceSpecificDates
	existing.RecurrenceDayParity = input.RecurrenceDayParity
	existing.UpdatedAt = now

	if err := taskdomain.ValidateRecurrence(existing); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidInput, err)
	}

	updated, err := s.repo.Update(ctx, existing)
	if err != nil {
		return nil, err
	}

	// Если задача была шаблоном — удаляем будущие незапущенные экземпляры.
	if wasTemplate {
		today := taskdomain.TruncateToDay(now)
		if err := s.repo.DeleteFutureByTemplate(ctx, updated.ID, today); err != nil {
			return nil, err
		}
	}
	// Если задача остаётся или становится шаблоном — материализуем экземпляры.
	if updated.IsTemplate() {
		if err := materialize(ctx, s.repo,updated, now); err != nil {
			return nil, err
		}
	}

	return updated, nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}
	return s.repo.Delete(ctx, id)
}

func (s *Service) List(ctx context.Context, includeTemplates bool) ([]taskdomain.Task, error) {
	return s.repo.List(ctx, includeTemplates)
}

// RefillTemplates создаёт задачи по всем активным шаблонам на 60 дней вперёд.
// Вызывается планировщиком; безопасен при повторном вызове (дубли игнорируются).
func (s *Service) RefillTemplates(ctx context.Context) error {
	templates, err := s.repo.ListTemplates(ctx)
	if err != nil {
		return err
	}

	now := s.now()
	for i := range templates {
		if err := materialize(ctx, s.repo, &templates[i], now); err != nil {
			return err
		}
	}

	return nil
}

// materialize создаёт задачи по шаблону на горизонт now..now+60d.
// Дубликаты обрабатываются уникальным индексом БД (ON CONFLICT DO NOTHING).
func materialize(ctx context.Context, repo Repository, template *taskdomain.Task, now time.Time) error {
	from := taskdomain.TruncateToDay(now)
	to := taskdomain.TruncateToDay(now.Add(materializationHorizon))

	occurrences := taskdomain.GenerateOccurrences(template, from, to)
	if len(occurrences) == 0 {
		return nil
	}

	instances := make([]taskdomain.Task, 0, len(occurrences))
	for _, occ := range occurrences {
		occCopy := occ
		instances = append(instances, taskdomain.Task{
			Title:        template.Title,
			Description:  template.Description,
			Status:       taskdomain.StatusNew,
			ParentTaskID: &template.ID,
			ScheduledAt:  &occCopy,
			CreatedAt:    now,
			UpdatedAt:    now,
		})
	}

	return repo.CreateBatch(ctx, instances)
}

// validateInput нормализует и проверяет поля title, description, status.
// Если defaultStatus == true, пустой статус заменяется на StatusNew (для Create).
func validateInput(title, description string, status taskdomain.Status, defaultStatus bool) (string, string, taskdomain.Status, error) {
	title = strings.TrimSpace(title)
	description = strings.TrimSpace(description)

	if title == "" {
		return "", "", "", fmt.Errorf("%w: title is required", ErrInvalidInput)
	}
	if defaultStatus && status == "" {
		status = taskdomain.StatusNew
	}
	if !status.Valid() {
		return "", "", "", fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}
	return title, description, status, nil
}
