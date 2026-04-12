package task

import "time"

type Status string

const (
	StatusNew        Status = "new"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
)

type Task struct {
	ID          int64
	Title       string
	Description string
	Status      Status
	ScheduledAt  *time.Time
	ParentTaskID *int64

	RecurrenceType          *RecurrenceType
	RecurrenceDailyInterval *int
	RecurrenceMonthlyDays   []int
	RecurrenceSpecificDates []time.Time
	RecurrenceDayParity *Parity

	CreatedAt time.Time
	UpdatedAt time.Time
}

// IsTemplate возвращает true, если задача является шаблоном периодичности.
func (t *Task) IsTemplate() bool {
	return t.RecurrenceType != nil && t.ParentTaskID == nil
}

// IsInstance возвращает true, если задача создана шаблоном.
func (t *Task) IsInstance() bool {
	return t.ParentTaskID != nil
}

// IsStandalone возвращает true, если задача разовая (нет правила и нет родителя).
func (t *Task) IsStandalone() bool {
	return t.RecurrenceType == nil && t.ParentTaskID == nil
}

func (s Status) Valid() bool {
	switch s {
	case StatusNew, StatusInProgress, StatusDone:
		return true
	default:
		return false
	}
}
