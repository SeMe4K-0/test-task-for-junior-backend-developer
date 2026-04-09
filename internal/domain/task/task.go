package task

import "time"
//статус я не меняю

type Status string

const (
	StatusNew        Status = "new"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
)

//добавляем dto для периодов

type PeriodType string

const (
	Daily   	PeriodType = "daily"
	Monthly 	PeriodType = "monthly"
	Specific 	PeriodType = "specific"
	EvenOdd		PeriodType = "even_odd"
)

type EvenOddType string

const (
	Even	EvenOddType = "even"
	Odd		EvenOddType = "odd"
)

//создаем конструктор 

type PeriodConf struct {
	Type          PeriodType 	  `json:"type"`
	IntervalDays  int             `json:"interval_days,omitempty"`
	MonthDay      int             `json:"month_day,omitempty"`      //как сказано в задании от 1 до 30, без учета 31 чисел и 29 февраля
	SpecificDates []time.Time     `json:"specific_dates,omitempty"` //
	EvenOddType   EvenOddType     `json:"even_odd_type,omitempty"`
}

//добавляем в дто тэскс наши периоды

type Task struct {
	ID          int64     	`json:"id"`
	Title       string    	`json:"title"`
	Description string    	`json:"description"`
	Status      Status    	`json:"status"`
	CreatedAt   time.Time 	`json:"created_at"`
	UpdatedAt   time.Time 	`json:"updated_at"`
	PeriodConf  *PeriodConf	`json:"period_conf"`
}


func (s Status) Valid() bool {
	switch s {
	case StatusNew, StatusInProgress, StatusDone:
		return true
	default:
		return false
	}
}

