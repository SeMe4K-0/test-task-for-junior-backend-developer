package recurrence

import (
	"testing"
	"time"
)

func date(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func intPtr(v int) *int { return &v }

func TestGenerateDaily_EveryDay(t *testing.T) {
	rule := &RecurrenceRule{
		Type:       TypeDaily,
		EveryNDays: intPtr(1),
		StartDate:  date(2026, 5, 1),
		EndDate:    date(2026, 5, 5),
	}
	dates, err := GenerateDates(rule)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(dates) != 5 {
		t.Fatalf("expected 5 dates, got %d", len(dates))
	}
	for i, d := range dates {
		expected := date(2026, 5, 1+i)
		if !d.Equal(expected) {
			t.Errorf("dates[%d] = %v, want %v", i, d, expected)
		}
	}
}

func TestGenerateDaily_EveryThirdDay(t *testing.T) {
	rule := &RecurrenceRule{
		Type:       TypeDaily,
		EveryNDays: intPtr(3),
		StartDate:  date(2026, 5, 1),
		EndDate:    date(2026, 5, 10),
	}
	dates, err := GenerateDates(rule)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := []time.Time{
		date(2026, 5, 1),
		date(2026, 5, 4),
		date(2026, 5, 7),
		date(2026, 5, 10),
	}
	if len(dates) != len(expected) {
		t.Fatalf("expected %d dates, got %d", len(expected), len(dates))
	}
	for i, d := range dates {
		if !d.Equal(expected[i]) {
			t.Errorf("dates[%d] = %v, want %v", i, d, expected[i])
		}
	}
}

func TestGenerateMonthly_Day15(t *testing.T) {
	rule := &RecurrenceRule{
		Type:       TypeMonthly,
		DayOfMonth: intPtr(15),
		StartDate:  date(2026, 1, 1),
		EndDate:    date(2026, 6, 30),
	}
	dates, err := GenerateDates(rule)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(dates) != 6 {
		t.Fatalf("expected 6 dates, got %d", len(dates))
	}
	for i, d := range dates {
		if d.Day() != 15 {
			t.Errorf("dates[%d] day = %d, want 15", i, d.Day())
		}
	}
}

func TestGenerateMonthly_Day31_ClampsToLastDay(t *testing.T) {
	rule := &RecurrenceRule{
		Type:       TypeMonthly,
		DayOfMonth: intPtr(31),
		StartDate:  date(2026, 1, 1),
		EndDate:    date(2026, 4, 30),
	}
	dates, err := GenerateDates(rule)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(dates) != 4 {
		t.Fatalf("expected 4 dates, got %d", len(dates))
	}
	expectedDays := []int{31, 28, 31, 30}
	for i, d := range dates {
		if d.Day() != expectedDays[i] {
			t.Errorf("dates[%d] = %v (day %d), want day %d", i, d, d.Day(), expectedDays[i])
		}
	}
}

func TestGenerateMonthly_Day29_LeapYear(t *testing.T) {
	rule := &RecurrenceRule{
		Type:       TypeMonthly,
		DayOfMonth: intPtr(29),
		StartDate:  date(2024, 2, 1),
		EndDate:    date(2024, 3, 31),
	}
	dates, err := GenerateDates(rule)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(dates) != 2 {
		t.Fatalf("expected 2 dates, got %d", len(dates))
	}
	if dates[0].Day() != 29 {
		t.Errorf("Feb date day = %d, want 29 (leap year)", dates[0].Day())
	}
	if dates[1].Day() != 29 {
		t.Errorf("Mar date day = %d, want 29", dates[1].Day())
	}
}

func TestGenerateSpecificDates(t *testing.T) {
	rule := &RecurrenceRule{
		Type: TypeSpecificDates,
		SpecificDates: []time.Time{
			date(2026, 4, 15),
			date(2026, 5, 10),
			date(2026, 5, 20),
			date(2026, 7, 1),
		},
		StartDate: date(2026, 5, 1),
		EndDate:   date(2026, 6, 30),
	}
	dates, err := GenerateDates(rule)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(dates) != 2 {
		t.Fatalf("expected 2 dates (filtered), got %d", len(dates))
	}
}

func TestGenerateEvenDays(t *testing.T) {
	rule := &RecurrenceRule{
		Type:      TypeEvenOdd,
		EvenOdd:   EvenDays,
		StartDate: date(2026, 5, 1),
		EndDate:   date(2026, 5, 10),
	}
	dates, err := GenerateDates(rule)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := []int{2, 4, 6, 8, 10}
	if len(dates) != len(expected) {
		t.Fatalf("expected %d dates, got %d", len(expected), len(dates))
	}
	for i, d := range dates {
		if d.Day() != expected[i] {
			t.Errorf("dates[%d] day = %d, want %d", i, d.Day(), expected[i])
		}
	}
}

func TestGenerateOddDays(t *testing.T) {
	rule := &RecurrenceRule{
		Type:      TypeEvenOdd,
		EvenOdd:   OddDays,
		StartDate: date(2026, 5, 1),
		EndDate:   date(2026, 5, 10),
	}
	dates, err := GenerateDates(rule)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := []int{1, 3, 5, 7, 9}
	if len(dates) != len(expected) {
		t.Fatalf("expected %d dates, got %d", len(expected), len(dates))
	}
	for i, d := range dates {
		if d.Day() != expected[i] {
			t.Errorf("dates[%d] day = %d, want %d", i, d.Day(), expected[i])
		}
	}
}

func TestGenerateDates_StartAfterEnd_Error(t *testing.T) {
	rule := &RecurrenceRule{
		Type:       TypeDaily,
		EveryNDays: intPtr(1),
		StartDate:  date(2026, 6, 1),
		EndDate:    date(2026, 5, 1),
	}
	_, err := GenerateDates(rule)
	if err == nil {
		t.Fatal("expected error for start_date > end_date")
	}
}

func TestGenerateDates_EmptyResult_Error(t *testing.T) {
	rule := &RecurrenceRule{
		Type:          TypeSpecificDates,
		SpecificDates: []time.Time{date(2026, 1, 1)},
		StartDate:     date(2026, 5, 1),
		EndDate:       date(2026, 5, 31),
	}
	_, err := GenerateDates(rule)
	if err == nil {
		t.Fatal("expected error for empty date list")
	}
}

func TestGenerateMonthly_StartMidMonth(t *testing.T) {
	rule := &RecurrenceRule{
		Type:       TypeMonthly,
		DayOfMonth: intPtr(5),
		StartDate:  date(2026, 3, 10),
		EndDate:    date(2026, 6, 30),
	}
	dates, err := GenerateDates(rule)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(dates) != 3 {
		t.Fatalf("expected 3 dates (Apr, May, Jun), got %d", len(dates))
	}
	if dates[0] != date(2026, 4, 5) {
		t.Errorf("first date = %v, want 2026-04-05", dates[0])
	}
}
