package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"
	"travel/benchmark/seed"
)

func main() {
	user := requireEnv("BENCHMARK_MYSQL_USER")
	password := requireEnv("BENCHMARK_MYSQL_PASSWORD")
	host := envOrDefault("BENCHMARK_MYSQL_HOST", "127.0.0.1")
	port := envOrDefault("BENCHMARK_MYSQL_PORT", "3306")
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/?charset=utf8mb4&parseTime=True&loc=%s", user, password, host, port, url.QueryEscape("Asia/Shanghai"))

	dataset, err := seed.Build(seed.DefaultScale())
	if err != nil {
		fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	if err := seed.ResetAndSeed(ctx, dsn, seed.BenchmarkDatabase, dataset); err != nil {
		fatal(err)
	}

	limit := 100
	if len(dataset.Posts) < limit {
		limit = len(dataset.Posts)
	}
	ids := make([]string, limit)
	for i := range ids {
		ids[i] = dataset.Posts[i].ID.String()
	}
	if err := os.WriteFile("benchmark/post_ids.txt", []byte(strings.Join(ids, "\n")+"\n"), 0o600); err != nil {
		fatal(fmt.Errorf("write benchmark post IDs: %w", err))
	}
	fmt.Printf("benchmark database ready: users=%d posts=%d comments=%d likes=%d favorites=%d foots=%d\n",
		len(dataset.Users), len(dataset.Posts), len(dataset.Comments), len(dataset.Likes), len(dataset.Favorites), len(dataset.Foots))
}

func requireEnv(name string) string {
	value := os.Getenv(name)
	if value == "" {
		fatal(fmt.Errorf("%s is required", name))
	}
	return value
}

func envOrDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
