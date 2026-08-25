package loadtest

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"sort"
	"strconv"
	"sync"
	"time"
)

type Config struct {
	URL         string
	URLs        []string
	Headers     map[string]string
	Requests    int
	Concurrency int
	Timeout     time.Duration
}

type Sample struct {
	Latency    time.Duration
	StatusCode int
	Err        error
}

type Summary struct {
	Total      int
	Successful int
	Errors     int
	ErrorRate  float64
	QPS        float64
	SuccessQPS float64
	P50        time.Duration
	P95        time.Duration
	P99        time.Duration
	Max        time.Duration
}

func Run(ctx context.Context, cfg Config) ([]Sample, time.Duration, error) {
	urls := append([]string(nil), cfg.URLs...)
	if len(urls) == 0 && cfg.URL != "" {
		urls = []string{cfg.URL}
	}
	if len(urls) == 0 || cfg.Requests <= 0 || cfg.Concurrency <= 0 || cfg.Timeout <= 0 {
		return nil, 0, errors.New("URL, requests, concurrency, and timeout must be positive")
	}
	transport := &http.Transport{
		MaxIdleConns:        cfg.Concurrency,
		MaxIdleConnsPerHost: cfg.Concurrency,
		MaxConnsPerHost:     cfg.Concurrency,
	}
	client := &http.Client{Transport: transport, Timeout: cfg.Timeout}
	defer transport.CloseIdleConnections()

	samples := make([]Sample, cfg.Requests)
	jobs := make(chan int)
	var workers sync.WaitGroup
	workerCount := cfg.Concurrency
	if workerCount > cfg.Requests {
		workerCount = cfg.Requests
	}
	workers.Add(workerCount)
	for i := 0; i < workerCount; i++ {
		go func() {
			defer workers.Done()
			for index := range jobs {
				started := time.Now()
				req, err := http.NewRequestWithContext(ctx, http.MethodGet, urls[index%len(urls)], nil)
				if err != nil {
					samples[index] = Sample{Latency: time.Since(started), Err: err}
					continue
				}
				for name, value := range cfg.Headers {
					req.Header.Set(name, value)
				}
				resp, err := client.Do(req)
				sample := Sample{Latency: time.Since(started), Err: err}
				if resp != nil {
					sample.StatusCode = resp.StatusCode
					_, _ = io.Copy(io.Discard, resp.Body)
					_ = resp.Body.Close()
				}
				samples[index] = sample
			}
		}()
	}

	started := time.Now()
	for i := range samples {
		jobs <- i
	}
	close(jobs)
	workers.Wait()
	return samples, time.Since(started), nil
}

func Summarize(samples []Sample, elapsed time.Duration) Summary {
	summary := Summary{Total: len(samples)}
	if len(samples) == 0 {
		return summary
	}
	latencies := make([]time.Duration, len(samples))
	for i, sample := range samples {
		latencies[i] = sample.Latency
		if sample.Err != nil || sample.StatusCode < 200 || sample.StatusCode >= 300 {
			summary.Errors++
		}
	}
	sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })
	summary.P50 = percentile(latencies, 0.50)
	summary.P95 = percentile(latencies, 0.95)
	summary.P99 = percentile(latencies, 0.99)
	summary.Max = latencies[len(latencies)-1]
	summary.ErrorRate = float64(summary.Errors) / float64(summary.Total)
	summary.Successful = summary.Total - summary.Errors
	if elapsed > 0 {
		summary.QPS = float64(summary.Total) / elapsed.Seconds()
		summary.SuccessQPS = float64(summary.Successful) / elapsed.Seconds()
	}
	return summary
}

func FailureCounts(samples []Sample) map[string]int {
	counts := make(map[string]int)
	for _, sample := range samples {
		if sample.Err != nil {
			var networkError net.Error
			if errors.Is(sample.Err, context.DeadlineExceeded) || (errors.As(sample.Err, &networkError) && networkError.Timeout()) {
				counts["timeout"]++
			} else {
				counts["network"]++
			}
			continue
		}
		if sample.StatusCode < 200 || sample.StatusCode >= 300 {
			counts["http_"+strconv.Itoa(sample.StatusCode)]++
		}
	}
	return counts
}

func FailureExamples(samples []Sample, limit int) []string {
	if limit <= 0 {
		return nil
	}
	examples := make([]string, 0, limit)
	for _, sample := range samples {
		var example string
		if sample.Err != nil {
			example = sample.Err.Error()
		} else if sample.StatusCode < 200 || sample.StatusCode >= 300 {
			example = fmt.Sprintf("HTTP %d", sample.StatusCode)
		}
		if example != "" {
			examples = append(examples, example)
			if len(examples) == limit {
				break
			}
		}
	}
	return examples
}

func percentile(sorted []time.Duration, fraction float64) time.Duration {
	if len(sorted) == 0 {
		return 0
	}
	index := int(math.Ceil(fraction*float64(len(sorted)))) - 1
	if index < 0 {
		index = 0
	}
	return sorted[index]
}
