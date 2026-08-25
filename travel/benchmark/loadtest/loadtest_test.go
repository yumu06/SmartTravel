package loadtest

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestSummarizeCalculatesLiteralPercentilesAndErrors(t *testing.T) {
	samples := []Sample{
		{Latency: time.Millisecond, StatusCode: 200},
		{Latency: 2 * time.Millisecond, StatusCode: 200},
		{Latency: 3 * time.Millisecond, StatusCode: 500},
		{Latency: 4 * time.Millisecond, StatusCode: 200},
		{Latency: 5 * time.Millisecond, Err: context.DeadlineExceeded},
	}

	summary := Summarize(samples, time.Second)

	if summary.Total != 5 || summary.Errors != 2 {
		t.Fatalf("counts = (%d total, %d errors), want (5, 2)", summary.Total, summary.Errors)
	}
	if summary.P50 != 3*time.Millisecond || summary.P95 != 5*time.Millisecond || summary.P99 != 5*time.Millisecond {
		t.Fatalf("percentiles = (%s, %s, %s), want (3ms, 5ms, 5ms)", summary.P50, summary.P95, summary.P99)
	}
	if summary.QPS != 5 {
		t.Fatalf("QPS = %v, want 5", summary.QPS)
	}
	if summary.Successful != 3 || summary.SuccessQPS != 3 {
		t.Fatalf("success = (%d, %v QPS), want (3, 3 QPS)", summary.Successful, summary.SuccessQPS)
	}
}

func TestFailureCountsSeparatesTimeoutNetworkAndHTTPStatus(t *testing.T) {
	samples := []Sample{
		{StatusCode: 200},
		{StatusCode: 500},
		{StatusCode: 429},
		{Err: context.DeadlineExceeded},
		{Err: errors.New("connection reset")},
	}

	counts := FailureCounts(samples)

	if counts["http_500"] != 1 || counts["http_429"] != 1 || counts["timeout"] != 1 || counts["network"] != 1 {
		t.Fatalf("FailureCounts() = %#v, want one failure in each category", counts)
	}
}

func TestFailureExamplesReturnsConcreteErrorsUpToLimit(t *testing.T) {
	samples := []Sample{
		{StatusCode: 200},
		{StatusCode: 503},
		{Err: errors.New("connection reset")},
		{StatusCode: 500},
	}

	examples := FailureExamples(samples, 2)

	if len(examples) != 2 || examples[0] != "HTTP 503" || examples[1] != "connection reset" {
		t.Fatalf("FailureExamples() = %#v, want [HTTP 503, connection reset]", examples)
	}
}

func TestRunExecutesEveryRequestAndCountsHTTPFailures(t *testing.T) {
	var calls atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if calls.Add(1)%2 == 0 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	samples, elapsed, err := Run(context.Background(), Config{
		URL:         server.URL,
		Requests:    10,
		Concurrency: 3,
		Timeout:     time.Second,
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	summary := Summarize(samples, elapsed)
	if calls.Load() != 10 || summary.Total != 10 || summary.Errors != 5 {
		t.Fatalf("result = (%d calls, %d total, %d errors), want (10, 10, 5)", calls.Load(), summary.Total, summary.Errors)
	}
}

func TestRunRotatesURLsAndSendsConfiguredHeader(t *testing.T) {
	var first atomic.Int64
	var second atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if request.Header.Get("Authorization") != "Bearer benchmark-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		switch request.URL.Path {
		case "/first":
			first.Add(1)
		case "/second":
			second.Add(1)
		default:
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	samples, elapsed, err := Run(context.Background(), Config{
		URLs:        []string{server.URL + "/first", server.URL + "/second"},
		Headers:     map[string]string{"Authorization": "Bearer benchmark-token"},
		Requests:    4,
		Concurrency: 1,
		Timeout:     time.Second,
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	summary := Summarize(samples, elapsed)
	if first.Load() != 2 || second.Load() != 2 || summary.Errors != 0 {
		t.Fatalf("result = (%d first, %d second, %d errors), want (2, 2, 0)", first.Load(), second.Load(), summary.Errors)
	}
}
