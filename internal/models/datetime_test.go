package models

import (
	"encoding/json"
	"testing"
	"time"
)

// TestJSONTimeUnmarshal tests the custom JSONTime unmarshaling with various formats
func TestJSONTimeUnmarshal(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "Flutter format with milliseconds",
			input:   `"2025-12-16T21:26:00.000Z"`,
			wantErr: false,
		},
		{
			name:    "ISO8601 without milliseconds",
			input:   `"2025-12-16T21:26:00Z"`,
			wantErr: false,
		},
		{
			name:    "Date and time only",
			input:   `"2025-12-16T21:26:00"`,
			wantErr: false,
		},
		{
			name:    "Invalid format",
			input:   `"invalid-date"`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var jt JSONTime
			err := json.Unmarshal([]byte(tt.input), &jt)

			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && jt.Time().IsZero() {
				t.Errorf("UnmarshalJSON() resulted in zero time")
			}
		})
	}
}

// TestJSONTimeMarshal tests the custom JSONTime marshaling
func TestJSONTimeMarshal(t *testing.T) {
	testTime := time.Date(2025, 12, 16, 21, 26, 0, 0, time.UTC)
	jt := JSONTime(testTime)

	data, err := json.Marshal(jt)
	if err != nil {
		t.Fatalf("MarshalJSON() error = %v", err)
	}

	// The result should be a valid RFC3339 format
	var result string
	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	// Parse it back to ensure it's valid
	_, err = time.Parse(time.RFC3339Nano, result)
	if err != nil {
		t.Errorf("MarshalJSON() produced invalid RFC3339 format: %v", err)
	}

	t.Logf("Marshaled time: %s", result)
}

// TestCreateEventRequestUnmarshal tests the complete event creation flow
func TestCreateEventRequestUnmarshal(t *testing.T) {
	jsonData := `{
		"title": "Test Event",
		"description": "A test event",
		"start_date": "2025-12-16T21:26:00.000Z",
		"end_date": "2025-12-16T23:26:00.000Z",
		"location": "Test Location",
		"category": "Academic"
	}`

	var req CreateEventRequest
	err := json.Unmarshal([]byte(jsonData), &req)
	if err != nil {
		t.Fatalf("Failed to unmarshal CreateEventRequest: %v", err)
	}

	if req.Title != "Test Event" {
		t.Errorf("Title mismatch: got %s, want Test Event", req.Title)
	}

	startTime := req.StartDate.Time()
	endTime := req.EndDate.Time()

	if startTime.IsZero() {
		t.Error("StartDate is zero after unmarshaling")
	}

	if endTime.IsZero() {
		t.Error("EndDate is zero after unmarshaling")
	}

	if !endTime.After(startTime) {
		t.Error("EndDate is not after StartDate")
	}

	t.Logf("Successfully parsed event: %s from %s to %s", req.Title, startTime, endTime)
}
