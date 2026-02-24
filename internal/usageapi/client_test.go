package usageapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestClient_FetchSuccess(t *testing.T) {
	t.Parallel()

	resp := UsageResponse{
		FiveHour: &UsageWindow{
			Utilization: json.Number("0.36"),
			ResetsAt:    strPtr("2025-06-01T15:00:00Z"),
		},
		SevenDay: &UsageWindow{
			Utilization: json.Number("0.62"),
			ResetsAt:    strPtr("2025-06-05T21:00:00Z"),
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("Authorization = %q, want %q", r.Header.Get("Authorization"), "Bearer test-token")
		}
		if r.Header.Get("anthropic-beta") != betaHeader {
			t.Errorf("anthropic-beta = %q, want %q", r.Header.Get("anthropic-beta"), betaHeader)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	client := newTestClient("test-token", srv.URL)
	got, err := client.Fetch()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.FiveHour == nil {
		t.Fatal("FiveHour is nil")
	}
	if f := got.FiveHour.UtilizationFloat(); f != 0.36 {
		t.Errorf("FiveHour.Utilization = %f, want 0.36", f)
	}
	if got.FiveHour.ResetsAt == nil || *got.FiveHour.ResetsAt != "2025-06-01T15:00:00Z" {
		t.Errorf("FiveHour.ResetsAt unexpected")
	}
}

func TestClient_Fetch401(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	client := newTestClient("bad-token", srv.URL)
	_, err := client.Fetch()
	if err == nil {
		t.Fatal("expected error for 401, got nil")
	}
}

func TestClient_Fetch5xxRetry(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		n := calls.Add(1)
		if n == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(UsageResponse{FiveHour: &UsageWindow{Utilization: json.Number("0.5")}})
	}))
	defer srv.Close()

	client := newTestClient("test-token", srv.URL)
	got, err := client.Fetch()
	if err != nil {
		t.Fatalf("unexpected error after retry: %v", err)
	}
	if got.FiveHour == nil {
		t.Fatal("FiveHour is nil after retry")
	}
	if calls.Load() != 2 {
		t.Errorf("expected 2 requests (original + retry), got %d", calls.Load())
	}
}

func TestClient_FetchMalformedJSON(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("not json"))
	}))
	defer srv.Close()

	client := newTestClient("test-token", srv.URL)
	_, err := client.Fetch()
	if err == nil {
		t.Fatal("expected error for malformed JSON, got nil")
	}
}

func TestClient_UtilizationAsInt(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// utilization as integer.
		w.Write([]byte(`{"five_hour":{"utilization":1,"resets_at":"2025-06-01T15:00:00Z"}}`))
	}))
	defer srv.Close()

	client := newTestClient("test-token", srv.URL)
	got, err := client.Fetch()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.FiveHour.UtilizationFloat() != 1.0 {
		t.Errorf("UtilizationFloat() = %f, want 1.0", got.FiveHour.UtilizationFloat())
	}
}

func TestClient_NullResetsAt(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"five_hour":{"utilization":0.5,"resets_at":null}}`))
	}))
	defer srv.Close()

	client := newTestClient("test-token", srv.URL)
	got, err := client.Fetch()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.FiveHour.ResetsAt != nil {
		t.Errorf("ResetsAt = %v, want nil", got.FiveHour.ResetsAt)
	}
}

func TestClient_ExtraUsage(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"extra_usage":{"is_enabled":true,"monthly_limit":5000,"used_credits":1250}}`))
	}))
	defer srv.Close()

	client := newTestClient("test-token", srv.URL)
	got, err := client.Fetch()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ExtraUsage == nil {
		t.Fatal("ExtraUsage is nil")
	}
	if !got.ExtraUsage.IsEnabled {
		t.Error("ExtraUsage.IsEnabled = false, want true")
	}
	if got.ExtraUsage.MonthlyLimit == nil || *got.ExtraUsage.MonthlyLimit != 5000 {
		t.Error("ExtraUsage.MonthlyLimit != 5000")
	}
	if got.ExtraUsage.UsedCredits == nil || *got.ExtraUsage.UsedCredits != 1250 {
		t.Error("ExtraUsage.UsedCredits != 1250")
	}
}

func TestUsageWindow_UtilizationFloat_Nil(t *testing.T) {
	t.Parallel()

	var w *UsageWindow
	if f := w.UtilizationFloat(); f != 0 {
		t.Errorf("UtilizationFloat() = %f, want 0 for nil", f)
	}
}

func newTestClient(token, baseURL string) *Client {
	c := NewClient(token)
	// Override the URL for testing by replacing the constant.
	// We use a custom transport that rewrites the URL.
	c.httpClient.Transport = &urlRewriter{baseURL: baseURL}
	return c
}

// urlRewriter rewrites requests to point to the test server.
type urlRewriter struct {
	baseURL string
}

func (u *urlRewriter) RoundTrip(req *http.Request) (*http.Response, error) {
	req.URL.Scheme = "http"
	req.URL.Host = u.baseURL[len("http://"):]
	return http.DefaultTransport.RoundTrip(req)
}

func strPtr(s string) *string { return &s }
