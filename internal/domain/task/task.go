package task

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

type Status string

const (
	StatusNew        Status = "new"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
)

// PeriodicityType определяет тип периодичности задачи
type PeriodicityType string

const (
	PeriodicityDaily        PeriodicityType = "daily"
	PeriodicityMonthly      PeriodicityType = "monthly"
	PeriodicitySpecificDate PeriodicityType = "specific_dates"
	PeriodicityEvenDays     PeriodicityType = "even_days"
	PeriodicityOddDays      PeriodicityType = "odd_days"
	PeriodicityNone         PeriodicityType = "none"
)

// PeriodicityConfig содержит параметры периодичности в зависимости от типа
type PeriodicityConfig struct {
	// Для ежедневных: интервал в днях (каждый n-й день)
	Interval int `json:"interval,omitempty"`
	// Для ежемесячных: день месяца (1-30)
	MonthDay int `json:"month_day,omitempty"`
	// Для конкретных дат: список дат в формате "2025-01-15"
	SpecificDates []string `json:"specific_dates,omitempty"`
	// Дата начала периодичности
	StartDate time.Time `json:"start_date"`
}

// Scan для поддержки сканирования из БД (PostgreSQL JSON)
func (pc *PeriodicityConfig) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, &pc)
}

// Value для поддержки записи в БД (PostgreSQL JSON)
func (pc *PeriodicityConfig) Value() (driver.Value, error) {
	return json.Marshal(pc)
}

type Task struct {
	ID                int64             `json:"id"`
	Title             string            `json:"title"`
	Description       string            `json:"description"`
	Status            Status            `json:"status"`
	PeriodicityType   PeriodicityType   `json:"periodicity_type"`
	PeriodicityConfig PeriodicityConfig `json:"periodicity_config"`
	CreatedAt         time.Time         `json:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at"`
}

func (s Status) Valid() bool {
	switch s {
	case StatusNew, StatusInProgress, StatusDone:
		return true
	default:
		return false
	}
}

func (pt PeriodicityType) Valid() bool {
	switch pt {
	case PeriodicityDaily, PeriodicityMonthly, PeriodicitySpecificDate,
		PeriodicityEvenDays, PeriodicityOddDays, PeriodicityNone:
		return true
	default:
		return false
	}
}
