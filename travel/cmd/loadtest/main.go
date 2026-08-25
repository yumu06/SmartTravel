package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"travel/benchmark/loadtest"
)

func main() {
	url := flag.String("url", "http://127.0.0.1:1016/travel/notice/list?pageNum=1&pageSize=20", "HTTP URL to test")
	urlsFile := flag.String("urls-file", "", "file containing one absolute URL per line")
	requests := flag.Int("n", 1000, "total request count")
	concurrency := flag.Int("c", 10, "concurrent workers")
	timeout := flag.Duration("timeout", 2*time.Second, "per-request timeout")
	flag.Parse()
	urls := []string{}
	if *urlsFile != "" {
		content, err := os.ReadFile(*urlsFile)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		for _, line := range strings.Split(string(content), "\n") {
			line = strings.TrimSpace(line)
			if line != "" {
				urls = append(urls, line)
			}
		}
	}
	headers := map[string]string{}
	if token := os.Getenv("BENCHMARK_ACCESS_TOKEN"); token != "" {
		headers["Authorization"] = "Bearer " + token
	}

	samples, elapsed, err := loadtest.Run(context.Background(), loadtest.Config{
		URL:         *url,
		URLs:        urls,
		Headers:     headers,
		Requests:    *requests,
		Concurrency: *concurrency,
		Timeout:     *timeout,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	summary := loadtest.Summarize(samples, elapsed)
	fmt.Printf("requests=%d successful=%d concurrency=%d elapsed=%s attempted_qps=%.2f success_qps=%.2f errors=%d error_rate=%.4f p50=%s p95=%s p99=%s max=%s\n",
		summary.Total,
		summary.Successful,
		*concurrency,
		elapsed.Round(time.Millisecond),
		summary.QPS,
		summary.SuccessQPS,
		summary.Errors,
		summary.ErrorRate,
		summary.P50,
		summary.P95,
		summary.P99,
		summary.Max,
	)
	failures := loadtest.FailureCounts(samples)
	keys := make([]string, 0, len(failures))
	for key := range failures {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		fmt.Printf("failure[%s]=%d\n", key, failures[key])
	}
	for _, example := range loadtest.FailureExamples(samples, 3) {
		fmt.Printf("failure_example=%s\n", example)
	}
}
