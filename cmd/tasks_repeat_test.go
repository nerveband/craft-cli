package cmd

import (
	"encoding/json"
	"testing"
)

func TestBuildRepeatConfig(t *testing.T) {
	t.Run("returns nil when nothing set", func(t *testing.T) {
		cfg, err := buildRepeatConfig(repeatFlagValues{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg != nil {
			t.Fatalf("expected nil config, got %+v", cfg)
		}
	})

	t.Run("builds full config", func(t *testing.T) {
		cfg, err := buildRepeatConfig(repeatFlagValues{
			Type:        "weekly",
			Interval:    2,
			Weekdays:    []int{1, 3},
			Reminder:    "09:00",
			DynamicDays: true,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.Type != "flexible" || cfg.Frequency != "weekly" || cfg.Interval != 2 || cfg.Weekly == nil || cfg.Reminder == nil {
			t.Errorf("unexpected config: %+v", cfg)
		}
	})

	t.Run("rejects invalid type", func(t *testing.T) {
		if _, err := buildRepeatConfig(repeatFlagValues{Type: "hourly"}); err == nil {
			t.Fatal("expected error for invalid repeat type")
		}
	})

	t.Run("rejects modifiers without a rule", func(t *testing.T) {
		if _, err := buildRepeatConfig(repeatFlagValues{Interval: 2}); err == nil {
			t.Fatal("expected error when --repeat-interval set without --repeat/--repeat-frequency")
		}
	})

	t.Run("rejects out-of-range weekday", func(t *testing.T) {
		if _, err := buildRepeatConfig(repeatFlagValues{Type: "weekly", Weekdays: []int{7}}); err == nil {
			t.Fatal("expected error for weekday 7")
		}
	})

	t.Run("rejects bad end date", func(t *testing.T) {
		if _, err := buildRepeatConfig(repeatFlagValues{Type: "daily", EndDate: "31-12-2026"}); err == nil {
			t.Fatal("expected error for non-ISO end date")
		}
	})

	t.Run("builds monthly and yearly rules from the contract fields", func(t *testing.T) {
		cases := []struct {
			name string
			in   repeatFlagValues
			want string
		}{
			{"monthly days", repeatFlagValues{Type: "monthly", StartDate: "2026-12-01", MonthDays: []int{1, 15}, SkipWeekends: true},
				`{"type":"fixed","frequency":"monthly","startDate":"2026-12-01","monthly":{"days":[1,15],"skipWeekends":true}}`},
			{"monthly last day", repeatFlagValues{Type: "monthly", DayOfMonth: "last"},
				`{"type":"fixed","frequency":"monthly","monthly":{"dynamicDays":[{"selector":"last","type":"day"}]}}`},
			{"yearly", repeatFlagValues{Type: "yearly", Months: []int{3, 9}, MonthDays: []int{4}},
				`{"type":"fixed","frequency":"yearly","yearly":{"days":[4],"months":[3,9]}}`},
		}
		for _, tc := range cases {
			cfg, err := buildRepeatConfig(tc.in)
			if err != nil {
				t.Fatalf("%s: unexpected error: %v", tc.name, err)
			}
			got, err := json.Marshal(cfg)
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != tc.want {
				t.Errorf("%s:\n got %s\nwant %s", tc.name, got, tc.want)
			}
		}
	})

	t.Run("rejects values the contract does not allow", func(t *testing.T) {
		for name, in := range map[string]repeatFlagValues{
			"bad start date":                   {Type: "daily", StartDate: "01/12/2026"},
			"month day out of range":           {Type: "monthly", MonthDays: []int{32}},
			"month out of range":               {Type: "yearly", Months: []int{13}},
			"months on non-yearly rule":        {Type: "monthly", Months: []int{3}},
			"month days on weekly rule":        {Type: "weekly", MonthDays: []int{1}},
			"unknown day selector":             {Type: "monthly", DayOfMonth: "middle"},
			"yearly options without months":    {Type: "yearly", DayOfMonth: "first"},
			"yearly month days without months": {Type: "yearly", MonthDays: []int{1}},
		} {
			if _, err := buildRepeatConfig(in); err == nil {
				t.Errorf("%s: expected error", name)
			}
		}
	})
}

func TestTasksCmdsRegisterRepeatFlags(t *testing.T) {
	for _, name := range []string{"repeat", "repeat-frequency", "repeat-interval",
		"repeat-weekdays", "repeat-end", "repeat-reminder",
		"repeat-skip-weekends", "repeat-dynamic-days", "repeat-start",
		"repeat-months", "repeat-month-days", "repeat-day"} {
		if tasksAddCmd.Flags().Lookup(name) == nil {
			t.Errorf("tasks add missing --%s", name)
		}
		if tasksUpdateCmd.Flags().Lookup(name) == nil {
			t.Errorf("tasks update missing --%s", name)
		}
	}
}
