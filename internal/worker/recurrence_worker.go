package worker

import (
	"context"
	"log/slog"
	"time"

	"example.com/taskservice/internal/config"
	taskusecase "example.com/taskservice/internal/usecase/task"
)

// RecurrenceWorker генерирует задачи для всех активных правил периодичности
type RecurrenceWorker struct {
	usecase taskusecase.Usecase
	logger  *slog.Logger
	cfg     config.WorkerCfg
}

func NewRecurrenceWorker(usecase taskusecase.Usecase, logger *slog.Logger, cfg config.WorkerCfg) *RecurrenceWorker {
	return &RecurrenceWorker{
		usecase: usecase,
		logger:  logger,
		cfg:     cfg,
	}
}

// генерирует задачи на диапозон сегодня + LookaheadDays
func (w *RecurrenceWorker) Run(ctx context.Context) error {
	now := time.Now().UTC()
	from := now.Truncate(24 * time.Hour)
	to := from.AddDate(0, 0, w.cfg.LookaheadDays-1)

	w.logger.Info("generating recurrenced tasks",
		"from", from.Format("2006-01-02"),
		"to", to.Format("2006-01-02"),
		"days", w.cfg.LookaheadDays,
	)

	start := time.Now()
	if err := w.usecase.CreateRecurrencedTasks(ctx, from, to); err != nil {
		w.logger.Error("failed to generate recurrenced tasks",
			"from", from.Format("2006-01-02"),
			"to", to.Format("2006-01-02"),
			"error", err,
			"elapsed", time.Since(start),
		)
		return err
	}

	w.logger.Info("recurrenced tasks generated successfully",
		"from", from.Format("2006-01-02"),
		"to", to.Format("2006-01-02"),
		"elapsed", time.Since(start),
	)
	return nil
}
