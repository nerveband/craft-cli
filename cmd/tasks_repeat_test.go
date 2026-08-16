package cmd

import "testing"

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
			Type:         "weekly",
			Interval:     2,
			Weekdays:     []int{1, 3},
			EndDate:      "2026-12-31",
			Reminder:     "09:00",
			SkipWeekends: true,
			DynamicDays:  true,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.Type != "weekly" || cfg.Interval != 2 || len(cfg.Weekdays) != 2 ||
			cfg.EndDate != "2026-12-31" || cfg.Reminder != "09:00" ||
			!cfg.SkipWeekends || !cfg.DynamicDays {
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
}

func TestTasksCmdsRegisterRepeatFlags(t *testing.T) {
	for _, name := range []string{"repeat", "repeat-frequency", "repeat-interval",
		"repeat-weekdays", "repeat-end", "repeat-reminder",
		"repeat-skip-weekends", "repeat-dynamic-days"} {
		if tasksAddCmd.Flags().Lookup(name) == nil {
			t.Errorf("tasks add missing --%s", name)
		}
		if tasksUpdateCmd.Flags().Lookup(name) == nil {
			t.Errorf("tasks update missing --%s", name)
		}
	}
}
