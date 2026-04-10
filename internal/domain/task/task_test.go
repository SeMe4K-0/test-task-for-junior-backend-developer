package task

import "testing"

func TestStatusValid(t *testing.T) {
	tests := []struct {
		name     string
		status   Status
		expected bool
	}{
		{"valid new", StatusNew, true},
		{"valid in_progress", StatusInProgress, true},
		{"valid done", StatusDone, true},
		{"invalid status", Status("invalid"), false},
		{"empty status", Status(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := tt.status.Valid(); result != tt.expected {
				t.Errorf("Status.Valid() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestRecurrenceTypeValid(t *testing.T) {
	tests := []struct {
		name     string
		recType  RecurrenceType
		expected bool
	}{
		{"daily", RecurrenceTypeDaily, true},
		{"monthly", RecurrenceTypeMonthly, true},
		{"specific_dates", RecurrenceTypeSpecificDates, true},
		{"parity", RecurrenceTypeParity, true},
		{"invalid", RecurrenceType("invalid"), false},
		{"empty", RecurrenceType(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := tt.recType.Valid(); result != tt.expected {
				t.Errorf("RecurrenceType.Valid() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestParityValid(t *testing.T) {
	tests := []struct {
		name     string
		parity   Parity
		expected bool
	}{
		{"even", ParityEven, true},
		{"odd", ParityOdd, true},
		{"invalid", Parity("invalid"), false},
		{"empty", Parity(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := tt.parity.Valid(); result != tt.expected {
				t.Errorf("Parity.Valid() = %v, want %v", result, tt.expected)
			}
		})
	}
}
