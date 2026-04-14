package task

import (
	"context"
	"log/slog"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

const (
	defaultCheckInterval = 1 * time.Hour
	defaultLookAheadDays = 7
)

// Scheduler — фоновый процесс, генерирующий экземпляры задач по правилам периодичности.
type Scheduler struct {
	taskRepo       Repository
	recurrenceRepo RecurrenceRepository
	logger         *slog.Logger
	interval       time.Duration
	lookAheadDays  int
	now            func() time.Time
}

type SchedulerOption func(*Scheduler)

func WithCheckInterval(d time.Duration) SchedulerOption {
	return func(s *Scheduler) {
		s.interval = d
	}
}

func WithLookAheadDays(days int) SchedulerOption {
	return func(s *Scheduler) {
		s.lookAheadDays = days
	}
}

func NewScheduler(
	taskRepo Repository,
	recurrenceRepo RecurrenceRepository,
	logger *slog.Logger,
	opts ...SchedulerOption,
) *Scheduler {
	s := &Scheduler{
		taskRepo:       taskRepo,
		recurrenceRepo: recurrenceRepo,
		logger:         logger,
		interval:       defaultCheckInterval,
		lookAheadDays:  defaultLookAheadDays,
		now:            func() time.Time { return time.Now().UTC() },
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// Run запускает планировщик. Блокирует горутину до отмены контекста.
func (s *Scheduler) Run(ctx context.Context) {
	s.logger.Info("scheduler started",
		"interval", s.interval,
		"look_ahead_days", s.lookAheadDays,
	)

	// Выполняем сразу при старте.
	s.tick(ctx)

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("scheduler stopped")
			return
		case <-ticker.C:
			s.tick(ctx)
		}
	}
}

func (s *Scheduler) tick(ctx context.Context) {
	now := s.now()

	for i := 0; i < s.lookAheadDays; i++ {
		date := now.AddDate(0, 0, i)
		if err := s.generateTasksForDate(ctx, date); err != nil {
			s.logger.Error("failed to generate tasks",
				"date", date.Format(dateLayout),
				"error", err,
			)
		}
	}
}

func (s *Scheduler) generateTasksForDate(ctx context.Context, date time.Time) error {
	rules, err := s.recurrenceRepo.ListActive(ctx, date)
	if err != nil {
		return err
	}

	dateStr := date.Format(dateLayout)

	for i := range rules {
		rule := &rules[i]

		if !rule.ShouldCreateOn(date) {
			continue
		}

		// Проверяем, не создан ли уже экземпляр на эту дату.
		exists, err := s.taskRepo.ExistsByParentAndDate(ctx, rule.TaskID, dateStr)
		if err != nil {
			s.logger.Error("failed to check existing task",
				"task_id", rule.TaskID,
				"date", dateStr,
				"error", err,
			)

			continue
		}

		if exists {
			continue
		}

		// Получаем шаблон.
		template, err := s.taskRepo.GetByID(ctx, rule.TaskID)
		if err != nil {
			s.logger.Error("failed to get template task",
				"task_id", rule.TaskID,
				"error", err,
			)

			continue
		}

		scheduledDate, _ := time.Parse(dateLayout, dateStr)
		now := s.now()

		instance := &taskdomain.Task{
			Title:         template.Title,
			Description:   template.Description,
			Status:        taskdomain.StatusNew,
			ParentTaskID:  &template.ID,
			ScheduledDate: &scheduledDate,
			IsTemplate:    false,
			CreatedAt:     now,
			UpdatedAt:     now,
		}

		if _, err := s.taskRepo.Create(ctx, instance); err != nil {
			s.logger.Error("failed to create task instance",
				"parent_task_id", template.ID,
				"date", dateStr,
				"error", err,
			)

			continue
		}

		s.logger.Info("created task instance",
			"parent_task_id", template.ID,
			"date", dateStr,
			"title", template.Title,
		)
	}

	return nil
}
