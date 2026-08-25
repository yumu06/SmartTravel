package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"
	"travel/auth"
	"travel/cache"
	traveljwt "travel/pkg/jwt"
)

type tokenConfig struct {
	accessSecret  []byte
	refreshSecret []byte
	issuer        string
	userID        uint64
}

func issueBenchmarkToken(ctx context.Context, cfg tokenConfig, store auth.SessionStore) (string, error) {
	manager, err := traveljwt.NewManager(traveljwt.Config{
		AccessSecret: cfg.accessSecret, RefreshSecret: cfg.refreshSecret,
		AccessTTL: 15 * time.Minute, RefreshTTL: 7 * 24 * time.Hour, Issuer: cfg.issuer,
	})
	if err != nil {
		return "", err
	}
	pair, err := auth.NewService(manager, store).StartSession(ctx, cfg.userID)
	if err != nil {
		return "", err
	}
	return pair.AccessToken, nil
}

func main() {
	redisDB, err := strconv.Atoi(envOrDefault("BENCHMARK_REDIS_DB", "0"))
	if err != nil {
		fatal(fmt.Errorf("parse BENCHMARK_REDIS_DB: %w", err))
	}
	client := cache.NewRedisClient(cache.RedisConfig{
		Addr:     envOrDefault("BENCHMARK_REDIS_ADDR", "127.0.0.1:6379"),
		Password: os.Getenv("BENCHMARK_REDIS_PASSWORD"), DB: redisDB,
		PoolSize: 2, MinIdleConns: 1, DialTimeout: 2 * time.Second,
		ReadTimeout: 2 * time.Second, WriteTimeout: 2 * time.Second,
	})
	defer client.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := cache.Ping(ctx, client); err != nil {
		fatal(fmt.Errorf("connect benchmark Redis: %w", err))
	}
	token, err := issueBenchmarkToken(ctx, tokenConfig{
		accessSecret:  []byte(requireEnv("BENCHMARK_JWT_ACCESS_SECRET")),
		refreshSecret: []byte(requireEnv("BENCHMARK_JWT_REFRESH_SECRET")),
		issuer:        "ongoing-trip-benchmark", userID: 1,
	}, auth.NewRedisSessionStore(client))
	if err != nil {
		fatal(err)
	}
	fmt.Println(token)
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
