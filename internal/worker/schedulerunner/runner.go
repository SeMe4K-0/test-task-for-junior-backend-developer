package schedulerunner

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	scheduledomain "example.com/taskservice/internal/domain/schedule"
	taskdomain "example.com/taskservice/internal/domain/task"
)

type TaskRepository interface {
	Create(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error)
}

type ScheduleRepository interface {
	ListActive(ctx context.Context) ([]scheduledomain.Schedule, error)
	UpdateLastCreatedDate(ctx context.Context, id int64, date time.Time) error
}

type Runner struct {
	tasks     TaskRepository
	schedules ScheduleRepository
	logger    *slog.Logger
	interval  time.Duration
	now       func() time.Time
}

func New(tasks TaskRepository, schedules ScheduleRepository, logger *slog.Logger, interval time.Duration) *Runner {
	return &Runner{
		tasks:     tasks,
		schedules: schedules,
		logger:    logger,
		interval:  interval,
		now:       func() time.Time { return time.Now().UTC() },
	}
}

func (r *Runner) Start(ctx context.Context) {
	r.logger.Info("schedule runner started", "interval", r.interval.String())

	r.tick(ctx)

	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			r.logger.Info("schedule runner stopped")
			return
		case <-ticker.C:
			r.tick(ctx)
		}
	}
}

func (r *Runner) tick(ctx context.Context) {
	schedules, err := r.schedules.ListActive(ctx)
	if err != nil {
		r.logger.Error("list active schedules", "error", err)
		return
	}

	today := startOfDay(r.now())

	for i := range schedules {
		s := &schedules[i]
		if err := r.processSchedule(ctx, s, today); err != nil {
			r.logger.Error("process schedule", "schedule_id", s.ID, "error", err)
		}
	}
}

func (r *Runner) processSchedule(ctx context.Context, s *scheduledomain.Schedule, today time.Time) error {
	if s.LastCreatedDate != nil && sameDate(*s.LastCreatedDate, today) {
		return nil
	}

	should, err := shouldCreateToday(s, today)
	if err != nil {
		return err
	}

	if !should {
		return nil
	}

	task := &taskdomain.Task{
		Title:       s.Title,
		Description: s.Description,
		Status:      taskdomain.StatusNew,
		ScheduleID:  &s.ID,
		CreatedAt:   r.now(),
		UpdatedAt:   r.now(),
	}

	if _, err := r.tasks.Create(ctx, task); err != nil {
		return err
	}

	if err := r.schedules.UpdateLastCreatedDate(ctx, s.ID, today); err != nil {
		return err
	}

	r.logger.Info("task created from schedule", "schedule_id", s.ID, "title", s.Title)

	return nil
}

func shouldCreateToday(s *scheduledomain.Schedule, today time.Time) (bool, error) {
	switch s.RecurrenceType {
	case scheduledomain.DailyType:
		var p struct {
			Interval int `json:"interval"`
		}
		if err := json.Unmarshal(s.RecurrenceParams, &p); err != nil {
			return false, err
		}
		if s.LastCreatedDate == nil {
			return true, nil
		}
		diff := int(today.Sub(startOfDay(*s.LastCreatedDate)).Hours() / 24)
		return diff >= p.Interval, nil

	case scheduledomain.MonthlyType:
		var p struct {
			Day int `json:"day"`
		}
		if err := json.Unmarshal(s.RecurrenceParams, &p); err != nil {
			return false, err
		}
		return today.Day() == p.Day, nil

	case scheduledomain.EvenType:
		return today.Day()%2 == 0, nil

	case scheduledomain.OddType:
		return today.Day()%2 != 0, nil

	case scheduledomain.CustomDatesType:
		var p struct {
			Dates []string `json:"dates"`
		}
		if err := json.Unmarshal(s.RecurrenceParams, &p); err != nil {
			return false, err
		}
		todayStr := today.Format("2006-01-02")
		for _, d := range p.Dates {
			if d == todayStr {
				return true, nil
			}
		}
		return false, nil

	case scheduledomain.EndOfMonthType:
		return today.Day() == lastDayOfMonth(today), nil
	}

	return false, nil
}

func startOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func sameDate(a, b time.Time) bool {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}

func lastDayOfMonth(t time.Time) int {
	firstOfNext := time.Date(t.Year(), t.Month()+1, 1, 0, 0, 0, 0, t.Location())
	return firstOfNext.AddDate(0, 0, -1).Day()
}
