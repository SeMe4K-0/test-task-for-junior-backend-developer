package schedule

import (
	"encoding/json"
	"time"
)

type RecurrenceType string

const (
	DailyType         RecurrenceType = "daily"
	MonthlyType       RecurrenceType = "monthly"
	EvenType          RecurrenceType = "even"
	OddType           RecurrenceType = "odd"
	CustomDatesType   RecurrenceType = "custom_dates"
	EndOfMonthType    RecurrenceType = "end_of_month"
)

type Schedule struct {
	ID               int64           `json:"id"`
	Title            string          `json:"title"`
	Description      string          `json:"description"`
	RecurrenceType   RecurrenceType  `json:"recurrence_type"`
	RecurrenceParams json.RawMessage `json:"recurrence_params"`
	IsActive         bool            `json:"is_active"`
	LastCreatedDate  *time.Time      `json:"last_created_date"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

func (r RecurrenceType) Valid() bool {
	switch r {
	case DailyType, MonthlyType, EvenType, OddType, CustomDatesType, EndOfMonthType:
		return true
	default:
		return false
	}
}
