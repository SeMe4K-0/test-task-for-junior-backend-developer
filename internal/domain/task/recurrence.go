package task

import "time"

type RecurrenceType string

const (
	None     RecurrenceType = "none"
	Daily    RecurrenceType = "daily"
	Monthly  RecurrenceType = "monthly"
	Parity   RecurrenceType = "parity"
	Specific RecurrenceType = "specific"
)

type Recurrence struct {
	Type          RecurrenceType `json:"rec_type" info:"Тип повторения"`
	Interval      int            `json:"interval,omitempty" info:"Интервал для задач с интервалом"`
	DayOfMonth    int            `json:"month_day,omitempty" info:"День_месяца"`
	SpecificDates []time.Time    `json:"spec_dates,omitempty" info:"Список специальных дней"`
	ParityType    bool           `json:"parity,omitempty"  info:"true-Четные, false-нечетные"`
	EndDate       time.Time  `json:"end_date,omitempty" info:"Дата последнего повторения"`
}


func (r RecurrenceType) Valid() bool {
	switch r {
	case None, Daily, Monthly, Parity, Specific:
		return true
	default:
		return false
	}
}