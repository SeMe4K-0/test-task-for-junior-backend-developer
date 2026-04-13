package task

import "time"

type Status string

const (
	StatusNew        Status = "new"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
	StatusCanceled   Status = "canceled"
)

type Task struct {
	ID           int64      `json:"id"`
	RuleID       *int       `json:"rule_id,omitempty"`
	Title        string     `json:"title"`
	Description  string     `json:"description"`
	DueDate      *time.Time `json:"due_date"`
	Status       Status     `json:"status"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	ScheduledFor *time.Time `json:"scheduled_for"`
}

func (s Status) Valid() bool {
	switch s {
	case StatusNew, StatusInProgress, StatusDone, StatusCanceled:
		return true
	default:
		return false
	}
}
