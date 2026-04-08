package task

import (
	"encoding/json"
	"fmt"
	"time"
)

const dateLayout = "2006-01-02"

type Date struct {
	time.Time
}

func ParseDate(raw string) (Date, error) {
	parsed, err := time.Parse(dateLayout, raw)
	if err != nil {
		return Date{}, fmt.Errorf("invalid date %q: %w", raw, err)
	}

	return NewDate(parsed), nil
}

func NewDate(t time.Time) Date {
	return NewDateFromParts(t.UTC().Year(), t.UTC().Month(), t.UTC().Day())
}

func NewDateFromParts(year int, month time.Month, day int) Date {
	return Date{
		Time: time.Date(year, month, day, 0, 0, 0, 0, time.UTC),
	}
}

func NewDateInLocation(t time.Time, loc *time.Location) Date {
	local := t.In(loc)
	return NewDateFromParts(local.Year(), local.Month(), local.Day())
}

func (d Date) String() string {
	return d.Time.Format(dateLayout)
}

func (d Date) AddDays(days int) Date {
	return NewDate(d.Time.AddDate(0, 0, days))
}

func (d Date) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

func (d *Date) UnmarshalJSON(data []byte) error {
	var raw string
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	parsed, err := ParseDate(raw)
	if err != nil {
		return err
	}

	*d = parsed

	return nil
}
