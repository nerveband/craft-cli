package models

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestRepeatConfigJSONRoundTrip(t *testing.T) {
	cfg := RepeatConfig{
		Type:         "weekly",
		Frequency:    "weekly",
		Interval:     2,
		Weekdays:     []int{1, 3, 5},
		EndDate:      "2026-12-31",
		SkipWeekends: true,
		DynamicDays:  true,
		Reminder:     "09:00",
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}
	for _, key := range []string{`"type":"weekly"`, `"frequency":"weekly"`, `"interval":2`,
		`"weekdays":[1,3,5]`, `"endDate":"2026-12-31"`, `"skipWeekends":true`,
		`"dynamicDays":true`, `"reminder":"09:00"`} {
		if !strings.Contains(string(data), key) {
			t.Errorf("marshaled RepeatConfig missing %s in %s", key, data)
		}
	}

	var decoded RepeatConfig
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	if !repeatConfigEqual(decoded, cfg) {
		t.Errorf("round trip mismatch: got %+v want %+v", decoded, cfg)
	}
}

func repeatConfigEqual(a, b RepeatConfig) bool {
	if a.Type != b.Type || a.Frequency != b.Frequency || a.Interval != b.Interval ||
		a.EndDate != b.EndDate || a.SkipWeekends != b.SkipWeekends ||
		a.DynamicDays != b.DynamicDays || a.Reminder != b.Reminder ||
		len(a.Weekdays) != len(b.Weekdays) {
		return false
	}
	for i := range a.Weekdays {
		if a.Weekdays[i] != b.Weekdays[i] {
			return false
		}
	}
	return true
}

func TestRepeatConfigOmitsEmptyFields(t *testing.T) {
	data, err := json.Marshal(RepeatConfig{Type: "daily"})
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}
	if string(data) != `{"type":"daily"}` {
		t.Errorf("expected omitempty on all optional fields, got %s", data)
	}
}
