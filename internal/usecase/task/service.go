package task

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

var ErrInvalidInput = errors.New("invalid task input")

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

const generationHorizon = 30

func (s *Service) Create(ctx context.Context, input CreateInput) (*taskdomain.Task, error) {
	input.Title = strings.TrimSpace(input.Title)
	if input.Title == "" {
		return nil, fmt.Errorf("%w: title required", ErrInvalidInput)
	}
	if input.Status == "" {
		input.Status = taskdomain.StatusNew
	}
	if !input.Status.Valid() {
		return nil, fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}
	if input.PeriodConf != nil {
		if err := input.PeriodConf.Validate(); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrInvalidInput, err)
		}
	}

	now := s.now()
	task := &taskdomain.Task{
		Title:       input.Title,
		Description: input.Description,
		Status:      input.Status,
		PeriodConf:  input.PeriodConf,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	created, err := s.repo.Create(ctx, task)
	if err != nil {
		return nil, err
	}

	if created.PeriodConf != nil {
		_ = s.generateTasksFromTemplate(ctx, created)
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

	if input.Title != "" {
		existing.Title = strings.TrimSpace(input.Title)
	}
	if input.Description != "" {
		existing.Description = strings.TrimSpace(input.Description)
	}
	if input.Status != "" {
		existing.Status = input.Status
	}
	if input.PeriodConf != nil {
		existing.PeriodConf = input.PeriodConf
	}
	existing.UpdatedAt = s.now()

	if existing.Title == "" {
		return nil, fmt.Errorf("%w: title required", ErrInvalidInput)
	}
	if !existing.Status.Valid() {
		return nil, fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}
	if existing.PeriodConf != nil {
		if err := existing.PeriodConf.Validate(); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrInvalidInput, err)
		}
	}

	updated, err := s.repo.Update(ctx, existing)
	if err != nil {
		return nil, err
	}

	if updated.PeriodConf != nil {
		_ = s.regenerateTasksForTemplate(ctx, updated)
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

// generateTasksFromTemplate создаёт задачи на ближайшие generationHorizon дней.
func (s *Service) generateTasksFromTemplate(ctx context.Context, template *taskdomain.Task) error {
	dates := s.calculateDates(template.PeriodConf, s.now(), generationHorizon)
	if len(dates) == 0 {
		return nil
	}

	now := s.now()
	tasks := make([]taskdomain.Task, 0, len(dates))
	for _, d := range dates {
		tasks = append(tasks, taskdomain.Task{
			Title:         template.Title,
			Description:   template.Description,
			Status:        taskdomain.StatusNew,
			PeriodConf:    nil, // у дочерних задач нет PeriodConf
			ScheduledDate: &d,
			CreatedAt:     now,
			UpdatedAt:     now,
		})
	}
	return s.repo.CreateBatch(ctx, tasks)
}

// regenerateTasksForTemplate удаляет все будущие задачи и создаёт новые.
func (s *Service) regenerateTasksForTemplate(ctx context.Context, template *taskdomain.Task) error {
	// удаляем все задачи с scheduled_date >= сегодня
	today := s.now().Truncate(24 * time.Hour)
	if err := s.repo.DeleteScheduledTasks(ctx, today); err != nil {
		return err
	}
	return s.generateTasksFromTemplate(ctx, template)
}

// calculateDates возвращает список дат (utc, начало дня) по PeriodConf.
func (s *Service) calculateDates(conf *taskdomain.PeriodConf, start time.Time, limit int) []time.Time {
	if conf == nil {
		return nil
	}
	start = start.Truncate(24 * time.Hour)
	var dates []time.Time

	switch conf.Type {
	case taskdomain.PeriodDaily:
		for i := 0; i < limit; i++ {
			dates = append(dates, start.AddDate(0, 0, i*conf.IntervalDays))
		}
	case taskdomain.PeriodMonthly:
		for i := 0; i < limit; i++ {
			candidate := start.AddDate(0, 0, i)
			if candidate.Day() == conf.MonthDay {
				dates = append(dates, candidate)
			}
		}
	case taskdomain.PeriodSpecific:
		for _, d := range conf.SpecificDates {
			d = time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, time.UTC)
			if (d.After(start) || d.Equal(start)) && len(dates) < limit {
				dates = append(dates, d)
			}
		}
	case taskdomain.PeriodEvenOdd:
		for i := 0; i < limit; i++ {
			candidate := start.AddDate(0, 0, i)
			day := candidate.Day()
			isEven := day%2 == 0
			if (conf.EvenOddType == taskdomain.Even && isEven) ||
				(conf.EvenOddType == taskdomain.Odd && !isEven) {
				dates = append(dates, candidate)
			}
		}
	}
	return dates
}
