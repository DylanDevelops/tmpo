package cmd

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestValidateDate(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid date",
			input:   "12-25-2024",
			wantErr: false,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "invalid format",
			input:   "2024-12-25",
			wantErr: true,
		},
		{
			name:    "invalid date",
			input:   "13-32-2024",
			wantErr: true,
		},
		{
			name:    "far future date",
			input:   time.Now().Add(72 * time.Hour).Format("01-02-2006"),
			wantErr: true,
		},
		{
			name:    "past date",
			input:   "01-01-2020",
			wantErr: false,
		},
		{
			name:    "today",
			input:   time.Now().Format("01-02-2006"),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDate(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateTime(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		// 12-hour format
		{
			name:    "12-hour with AM",
			input:   "9:30 AM",
			wantErr: false,
		},
		{
			name:    "12-hour with PM",
			input:   "5:45 PM",
			wantErr: false,
		},
		{
			name:    "12-hour lowercase am",
			input:   "9:30 am",
			wantErr: false,
		},
		{
			name:    "12-hour with leading zero",
			input:   "09:30 AM",
			wantErr: false,
		},
		// 24-hour format
		{
			name:    "24-hour format",
			input:   "14:30",
			wantErr: false,
		},
		{
			name:    "24-hour midnight",
			input:   "00:00",
			wantErr: false,
		},
		{
			name:    "24-hour late night",
			input:   "23:59",
			wantErr: false,
		},
		// Invalid inputs
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "invalid format",
			input:   "9.30 AM",
			wantErr: true, // Wrong separator
		},
		{
			name:    "invalid hour",
			input:   "25:00",
			wantErr: true,
		},
		{
			name:    "invalid minute",
			input:   "9:60 AM",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTime(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateEndDateTime(t *testing.T) {
	tests := []struct {
		name      string
		startDate string
		startTime string
		endDate   string
		endTime   string
		wantErr   bool
	}{
		{
			name:      "valid range same day",
			startDate: "12-25-2024",
			startTime: "9:00 AM",
			endDate:   "12-25-2024",
			endTime:   "5:00 PM",
			wantErr:   false,
		},
		{
			name:      "valid range next day",
			startDate: "12-25-2024",
			startTime: "11:00 PM",
			endDate:   "12-26-2024",
			endTime:   "1:00 AM",
			wantErr:   false,
		},
		{
			name:      "end before start",
			startDate: "12-25-2024",
			startTime: "5:00 PM",
			endDate:   "12-25-2024",
			endTime:   "9:00 AM",
			wantErr:   true,
		},
		{
			name:      "same time",
			startDate: "12-25-2024",
			startTime: "9:00 AM",
			endDate:   "12-25-2024",
			endTime:   "9:00 AM",
			wantErr:   true,
		},
		{
			name:      "24-hour format",
			startDate: "12-25-2024",
			startTime: "09:00",
			endDate:   "12-25-2024",
			endTime:   "17:00",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateEndDateTime(tt.startDate, tt.startTime, tt.endDate, tt.endTime)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestParseDateTime(t *testing.T) {
	tests := []struct {
		name     string
		date     string
		timeStr  string
		wantErr  bool
		wantHour int
		wantMin  int
	}{
		{
			name:     "12-hour AM",
			date:     "12-25-2024",
			timeStr:  "9:30 AM",
			wantErr:  false,
			wantHour: 9,
			wantMin:  30,
		},
		{
			name:     "12-hour PM",
			date:     "12-25-2024",
			timeStr:  "5:45 PM",
			wantErr:  false,
			wantHour: 17,
			wantMin:  45,
		},
		{
			name:     "24-hour format",
			date:     "12-25-2024",
			timeStr:  "14:30",
			wantErr:  false,
			wantHour: 14,
			wantMin:  30,
		},
		{
			name:     "midnight",
			date:     "12-25-2024",
			timeStr:  "12:00 AM",
			wantErr:  false,
			wantHour: 0,
			wantMin:  0,
		},
		{
			name:     "noon",
			date:     "12-25-2024",
			timeStr:  "12:00 PM",
			wantErr:  false,
			wantHour: 12,
			wantMin:  0,
		},
		{
			name:     "lowercase am/pm",
			date:     "12-25-2024",
			timeStr:  "9:30 am",
			wantErr:  false,
			wantHour: 9,
			wantMin:  30,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseDateTime(tt.date, tt.timeStr)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantHour, result.Hour())
				assert.Equal(t, tt.wantMin, result.Minute())
			}
		})
	}
}

func TestNormalizeAMPM(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"9:30 am", "9:30 AM"},
		{"9:30 AM", "9:30 AM"},
		{"9:30 pm", "9:30 PM"},
		{"9:30 PM", "9:30 PM"},
		{"14:30", "14:30"},
		{"MIXED case Am", "MIXED CASE AM"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := normalizeAMPM(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
