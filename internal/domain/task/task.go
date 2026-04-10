package task

import (
	"fmt"
	"time"
)

type Status string

const (
	StatusNew        Status = "new"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
)

type Task struct {
	ID            int64      `json:"id"`
	Title         string     `json:"title"`
	Description   string     `json:"description"`
	Status        Status     `json:"status"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	StartDateTime time.Time  `json:"start_date,omitempty" info:"Дата первого выполнения задачи"`
	Rec           Recurrence `json:"recurrence"`
}

func (s Status) Valid() bool {
	switch s {
	case StatusNew, StatusInProgress, StatusDone:
		return true
	default:
		return false
	}
}

func (t *Task) IsTaskActiveOn(target time.Time) bool {
	fmt.Println(t)
	taskStartDate := t.StartDateTime.Truncate(24 * time.Hour)
	if t.Rec.Type == None {
		return taskStartDate.Equal(target)
	}

	if taskStartDate.Equal(target) {
		if t.Rec.Type == None {
			return true
		}
	} else {
		if target.Before(taskStartDate) {
			return false
		}

		if !t.Rec.EndDate.IsZero() && target.After(t.Rec.EndDate) {
			return false
		}
	}

	switch t.Rec.Type {
	case Daily:
		daysDiff := int(target.Sub(taskStartDate).Hours() / 24)
		return daysDiff%t.Rec.Interval == 0

	case Monthly:
		return target.Day() == t.Rec.DayOfMonth

	case Specific:
		for _, d := range t.Rec.SpecificDates {
			if d.Year() == target.Year() && d.YearDay() == target.YearDay() {
				return true
			}
		}

	case Parity:

		isEven := target.Day()%2 == 0
		if t.Rec.ParityType {
			return isEven
		}

		return !isEven
	}

	return false
}
