package task

import "time"

type Status string

const (
	StatusNew        Status = "new"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
)

type Task struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      Status    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Recurrence *Recurrence `json:"recurrence,omitempty"`
	StartDateTime time.Time  `json:"start_date,omitempty"`
}

func (s Status) Valid() bool {
	switch s {
	case StatusNew, StatusInProgress, StatusDone:
		return true
	default:
		return false
	}
}

func (t *Task) IsActiveOn(target time.Time) bool {
	target = target.Truncate(24 * time.Hour)

	if t.Recurrence == nil || t.Recurrence.Type == None {
		if t.StartDateTime.IsZero() {
			return false
		}
		return t.StartDateTime.Truncate(24 * time.Hour).Equal(target)
	}

	if t.StartDateTime.IsZero() {
		return false
	}

	start := t.StartDateTime.Truncate(24 * time.Hour)

	if target.Before(start) {
		return false
	}

	r := t.Recurrence

	if r.EndDate != nil && target.After(r.EndDate.Truncate(24*time.Hour)) {
		return false
	}

	switch r.Type {

	case Daily:
		if r.IntervalDays <= 0 {
			return false
		}
		days := int(target.Sub(start).Hours() / 24)
		return days%r.IntervalDays == 0

	case Monthly:
		if r.DayOfMonth < 1 || r.DayOfMonth > 31 {
			return false
		}
		return target.Day() == r.DayOfMonth

	case Specific:
		for _, d := range r.SpecificDates {
			if d.IsZero() {
				continue
			}
			if d.Truncate(24 * time.Hour).Equal(target) {
				return true
			}
		}
		return false

	case Parity:
		isEven := target.Day()%2 == 0
		return isEven == r.ParityEven
	}

	return false
}