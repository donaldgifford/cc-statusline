package segments_test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/donaldgifford/cc-statusline/internal/model"
	"github.com/donaldgifford/cc-statusline/internal/render/segments"
	"github.com/donaldgifford/cc-statusline/internal/usageapi"
)

func mockFetcher(resp *usageapi.UsageResponse, err error) func() (*usageapi.UsageResponse, error) {
	return func() (*usageapi.UsageResponse, error) {
		return resp, err
	}
}

func intPtrUsage(v int) *int { return &v }

// --- FiveHour tests ---

func TestFiveHour_Name(t *testing.T) {
	t.Parallel()
	assertEqual(t, "five_hour", segments.FiveHour{}.Name())
}

func TestFiveHour_Source(t *testing.T) {
	t.Parallel()
	assertEqual(t, "experimental:usage_api", segments.FiveHour{}.Source())
}

func TestFiveHour_NilFetcher(t *testing.T) {
	t.Parallel()
	seg := segments.FiveHour{}
	_, err := seg.Render(&model.StatusData{}, th())
	if err == nil {
		t.Fatal("expected error for nil fetcher")
	}
}

func TestFiveHour_FetcherError(t *testing.T) {
	t.Parallel()
	seg := segments.FiveHour{
		UsageFetcher: mockFetcher(nil, fmt.Errorf("api down")),
	}
	_, err := seg.Render(&model.StatusData{}, th())
	if err == nil {
		t.Fatal("expected error from fetcher")
	}
}

func TestFiveHour_NilWindow(t *testing.T) {
	t.Parallel()
	seg := segments.FiveHour{
		UsageFetcher: mockFetcher(&usageapi.UsageResponse{}, nil),
	}
	got, err := seg.Render(&model.StatusData{}, th())
	assertNoError(t, err)
	assertEqual(t, "", got)
}

func TestFiveHour_WithData(t *testing.T) {
	t.Parallel()
	resetTime := time.Now().Add(3*time.Hour + 12*time.Minute).UTC().Format(time.RFC3339)
	seg := segments.FiveHour{
		UsageFetcher: mockFetcher(&usageapi.UsageResponse{
			FiveHour: &usageapi.UsageWindow{
				Utilization: json.Number("0.36"),
				ResetsAt:    &resetTime,
			},
		}, nil),
	}
	got, err := seg.Render(&model.StatusData{}, th())
	assertNoError(t, err)
	if got == "" {
		t.Fatal("expected non-empty output")
	}
	// Should contain percentage and time.
	if !containsStr(got, "36%") {
		t.Errorf("expected percentage in output, got %q", got)
	}
	if !containsStr(got, "5h:") {
		t.Errorf("expected '5h:' prefix in output, got %q", got)
	}
}

func TestFiveHour_ZeroUtilization(t *testing.T) {
	t.Parallel()
	resetTime := time.Now().Add(5 * time.Hour).UTC().Format(time.RFC3339)
	seg := segments.FiveHour{
		UsageFetcher: mockFetcher(&usageapi.UsageResponse{
			FiveHour: &usageapi.UsageWindow{
				Utilization: json.Number("0"),
				ResetsAt:    &resetTime,
			},
		}, nil),
	}
	got, err := seg.Render(&model.StatusData{}, th())
	assertNoError(t, err)
	if !containsStr(got, "0%") {
		t.Errorf("expected 0%% in output, got %q", got)
	}
}

func TestFiveHour_FullUtilization(t *testing.T) {
	t.Parallel()
	resetTime := time.Now().Add(1 * time.Hour).UTC().Format(time.RFC3339)
	seg := segments.FiveHour{
		UsageFetcher: mockFetcher(&usageapi.UsageResponse{
			FiveHour: &usageapi.UsageWindow{
				Utilization: json.Number("1.0"),
				ResetsAt:    &resetTime,
			},
		}, nil),
	}
	got, err := seg.Render(&model.StatusData{}, th())
	assertNoError(t, err)
	if !containsStr(got, "100%") {
		t.Errorf("expected 100%% in output, got %q", got)
	}
}

func TestFiveHour_UtilizationAsInt(t *testing.T) {
	t.Parallel()
	resetTime := time.Now().Add(2 * time.Hour).UTC().Format(time.RFC3339)
	seg := segments.FiveHour{
		UsageFetcher: mockFetcher(&usageapi.UsageResponse{
			FiveHour: &usageapi.UsageWindow{
				Utilization: json.Number("1"),
				ResetsAt:    &resetTime,
			},
		}, nil),
	}
	got, err := seg.Render(&model.StatusData{}, th())
	assertNoError(t, err)
	if !containsStr(got, "100%") {
		t.Errorf("expected 100%% in output, got %q", got)
	}
}

func TestFiveHour_NilResetsAt(t *testing.T) {
	t.Parallel()
	seg := segments.FiveHour{
		UsageFetcher: mockFetcher(&usageapi.UsageResponse{
			FiveHour: &usageapi.UsageWindow{
				Utilization: json.Number("0.5"),
				ResetsAt:    nil,
			},
		}, nil),
	}
	got, err := seg.Render(&model.StatusData{}, th())
	assertNoError(t, err)
	if !containsStr(got, "--") {
		t.Errorf("expected '--' for nil reset time, got %q", got)
	}
}

func TestFiveHour_PastResetTime(t *testing.T) {
	t.Parallel()
	pastTime := time.Now().Add(-1 * time.Hour).UTC().Format(time.RFC3339)
	seg := segments.FiveHour{
		UsageFetcher: mockFetcher(&usageapi.UsageResponse{
			FiveHour: &usageapi.UsageWindow{
				Utilization: json.Number("0.8"),
				ResetsAt:    &pastTime,
			},
		}, nil),
	}
	got, err := seg.Render(&model.StatusData{}, th())
	assertNoError(t, err)
	if !containsStr(got, "0m left") {
		t.Errorf("expected '0m left' for past reset time, got %q", got)
	}
}

// --- WeeklyLimits tests ---

func TestWeeklyLimits_Name(t *testing.T) {
	t.Parallel()
	assertEqual(t, "weekly_limits", segments.WeeklyLimits{}.Name())
}

func TestWeeklyLimits_Source(t *testing.T) {
	t.Parallel()
	assertEqual(t, "experimental:usage_api", segments.WeeklyLimits{}.Source())
}

func TestWeeklyLimits_NilFetcher(t *testing.T) {
	t.Parallel()
	seg := segments.WeeklyLimits{}
	_, err := seg.Render(&model.StatusData{}, th())
	if err == nil {
		t.Fatal("expected error for nil fetcher")
	}
}

func TestWeeklyLimits_FetcherError(t *testing.T) {
	t.Parallel()
	seg := segments.WeeklyLimits{
		UsageFetcher: mockFetcher(nil, fmt.Errorf("api error")),
	}
	_, err := seg.Render(&model.StatusData{}, th())
	if err == nil {
		t.Fatal("expected error from fetcher")
	}
}

func TestWeeklyLimits_NoWindows(t *testing.T) {
	t.Parallel()
	seg := segments.WeeklyLimits{
		UsageFetcher: mockFetcher(&usageapi.UsageResponse{}, nil),
	}
	got, err := seg.Render(&model.StatusData{}, th())
	assertNoError(t, err)
	assertEqual(t, "", got)
}

func TestWeeklyLimits_SonnetOnly(t *testing.T) {
	t.Parallel()
	resetTime := time.Date(2025, 6, 14, 14, 0, 0, 0, time.UTC).Format(time.RFC3339)
	seg := segments.WeeklyLimits{
		UsageFetcher: mockFetcher(&usageapi.UsageResponse{
			SevenDaySonnet: &usageapi.UsageWindow{
				Utilization: json.Number("0.45"),
				ResetsAt:    &resetTime,
			},
		}, nil),
	}
	got, err := seg.Render(&model.StatusData{}, th())
	assertNoError(t, err)
	if !containsStr(got, "wk:") {
		t.Errorf("expected 'wk:' prefix, got %q", got)
	}
	if !containsStr(got, "sonnet 45%") {
		t.Errorf("expected sonnet percentage, got %q", got)
	}
	if !containsStr(got, "resets") {
		t.Errorf("expected reset time, got %q", got)
	}
}

func TestWeeklyLimits_BothWindows(t *testing.T) {
	t.Parallel()
	sonnetReset := time.Date(2025, 6, 14, 14, 0, 0, 0, time.UTC).Format(time.RFC3339)
	allReset := time.Date(2025, 6, 12, 21, 0, 0, 0, time.UTC).Format(time.RFC3339)
	seg := segments.WeeklyLimits{
		UsageFetcher: mockFetcher(&usageapi.UsageResponse{
			SevenDaySonnet: &usageapi.UsageWindow{
				Utilization: json.Number("0.45"),
				ResetsAt:    &sonnetReset,
			},
			SevenDay: &usageapi.UsageWindow{
				Utilization: json.Number("0.62"),
				ResetsAt:    &allReset,
			},
		}, nil),
	}
	got, err := seg.Render(&model.StatusData{}, th())
	assertNoError(t, err)
	if !containsStr(got, "sonnet 45%") {
		t.Errorf("expected sonnet percentage, got %q", got)
	}
	if !containsStr(got, "all 62%") {
		t.Errorf("expected all percentage, got %q", got)
	}
	if !containsStr(got, "/") {
		t.Errorf("expected separator between windows, got %q", got)
	}
}

func TestWeeklyLimits_NilResetsAt(t *testing.T) {
	t.Parallel()
	seg := segments.WeeklyLimits{
		UsageFetcher: mockFetcher(&usageapi.UsageResponse{
			SevenDay: &usageapi.UsageWindow{
				Utilization: json.Number("0.30"),
				ResetsAt:    nil,
			},
		}, nil),
	}
	got, err := seg.Render(&model.StatusData{}, th())
	assertNoError(t, err)
	if !containsStr(got, "--") {
		t.Errorf("expected '--' for nil reset time, got %q", got)
	}
}

func TestWeeklyLimits_Over100Percent(t *testing.T) {
	t.Parallel()
	resetTime := time.Now().Add(24 * time.Hour).UTC().Format(time.RFC3339)
	seg := segments.WeeklyLimits{
		UsageFetcher: mockFetcher(&usageapi.UsageResponse{
			SevenDay: &usageapi.UsageWindow{
				Utilization: json.Number("1.5"),
				ResetsAt:    &resetTime,
			},
		}, nil),
	}
	got, err := seg.Render(&model.StatusData{}, th())
	assertNoError(t, err)
	if !containsStr(got, "150%") {
		t.Errorf("expected 150%% for over-100 utilization, got %q", got)
	}
}

// --- ExtraUsage tests ---

func TestExtraUsage_Name(t *testing.T) {
	t.Parallel()
	assertEqual(t, "extra_usage", segments.ExtraUsage{}.Name())
}

func TestExtraUsage_Source(t *testing.T) {
	t.Parallel()
	assertEqual(t, "experimental:usage_api", segments.ExtraUsage{}.Source())
}

func TestExtraUsage_NilFetcher(t *testing.T) {
	t.Parallel()
	seg := segments.ExtraUsage{}
	_, err := seg.Render(&model.StatusData{}, th())
	if err == nil {
		t.Fatal("expected error for nil fetcher")
	}
}

func TestExtraUsage_FetcherError(t *testing.T) {
	t.Parallel()
	seg := segments.ExtraUsage{
		UsageFetcher: mockFetcher(nil, fmt.Errorf("api error")),
	}
	_, err := seg.Render(&model.StatusData{}, th())
	if err == nil {
		t.Fatal("expected error from fetcher")
	}
}

func TestExtraUsage_NilExtraUsage(t *testing.T) {
	t.Parallel()
	seg := segments.ExtraUsage{
		UsageFetcher: mockFetcher(&usageapi.UsageResponse{}, nil),
	}
	got, err := seg.Render(&model.StatusData{}, th())
	assertNoError(t, err)
	assertEqual(t, "", got)
}

func TestExtraUsage_Disabled(t *testing.T) {
	t.Parallel()
	seg := segments.ExtraUsage{
		UsageFetcher: mockFetcher(&usageapi.UsageResponse{
			ExtraUsage: &usageapi.ExtraUsage{
				IsEnabled:    false,
				MonthlyLimit: intPtrUsage(5000),
				UsedCredits:  intPtrUsage(1250),
			},
		}, nil),
	}
	got, err := seg.Render(&model.StatusData{}, th())
	assertNoError(t, err)
	assertEqual(t, "", got)
}

func TestExtraUsage_Enabled(t *testing.T) {
	t.Parallel()
	seg := segments.ExtraUsage{
		UsageFetcher: mockFetcher(&usageapi.UsageResponse{
			ExtraUsage: &usageapi.ExtraUsage{
				IsEnabled:    true,
				MonthlyLimit: intPtrUsage(5000),
				UsedCredits:  intPtrUsage(1250),
			},
		}, nil),
	}
	got, err := seg.Render(&model.StatusData{}, th())
	assertNoError(t, err)
	assertEqual(t, "extra: $12.50 / $50.00", got)
}

func TestExtraUsage_ZeroUsed(t *testing.T) {
	t.Parallel()
	seg := segments.ExtraUsage{
		UsageFetcher: mockFetcher(&usageapi.UsageResponse{
			ExtraUsage: &usageapi.ExtraUsage{
				IsEnabled:    true,
				MonthlyLimit: intPtrUsage(10000),
				UsedCredits:  intPtrUsage(0),
			},
		}, nil),
	}
	got, err := seg.Render(&model.StatusData{}, th())
	assertNoError(t, err)
	assertEqual(t, "extra: $0.00 / $100.00", got)
}

func TestExtraUsage_NilCredits(t *testing.T) {
	t.Parallel()
	seg := segments.ExtraUsage{
		UsageFetcher: mockFetcher(&usageapi.UsageResponse{
			ExtraUsage: &usageapi.ExtraUsage{
				IsEnabled:    true,
				MonthlyLimit: intPtrUsage(5000),
				UsedCredits:  nil,
			},
		}, nil),
	}
	got, err := seg.Render(&model.StatusData{}, th())
	assertNoError(t, err)
	assertEqual(t, "extra: $0.00 / $50.00", got)
}

func TestExtraUsage_NilLimit(t *testing.T) {
	t.Parallel()
	seg := segments.ExtraUsage{
		UsageFetcher: mockFetcher(&usageapi.UsageResponse{
			ExtraUsage: &usageapi.ExtraUsage{
				IsEnabled:    true,
				MonthlyLimit: nil,
				UsedCredits:  intPtrUsage(1000),
			},
		}, nil),
	}
	got, err := seg.Render(&model.StatusData{}, th())
	assertNoError(t, err)
	assertEqual(t, "extra: $10.00 / $0.00", got)
}

// containsStr checks if s contains substr.
func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && searchStr(s, substr)
}

func searchStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
