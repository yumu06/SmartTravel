package main

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"

	"travel/auth"
)

func TestNewEngineCanDisableAccessLog(t *testing.T) {
	previousMode := gin.Mode()
	gin.SetMode(gin.ReleaseMode)
	t.Cleanup(func() { gin.SetMode(previousMode) })
	var output bytes.Buffer
	previous := gin.DefaultWriter
	gin.DefaultWriter = &output
	t.Cleanup(func() { gin.DefaultWriter = previous })

	engine := newEngine(false)
	engine.GET("/health", func(ctx *gin.Context) { ctx.Status(http.StatusNoContent) })
	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	engine.ServeHTTP(httptest.NewRecorder(), request)

	if output.Len() != 0 {
		t.Fatalf("disabled access log wrote %q", output.String())
	}
}

func TestInitializeAuthWiresTokenManagerAndRedisStore(t *testing.T) {
	server := miniredis.RunT(t)
	viper.Reset()
	t.Cleanup(func() {
		viper.Reset()
		auth.SetDefaultService(nil)
	})
	viper.Set("redis.addr", server.Addr())
	viper.Set("redis.pool_size", 5)
	viper.Set("redis.dial_timeout", "100ms")
	viper.Set("redis.read_timeout", "100ms")
	viper.Set("redis.write_timeout", "100ms")
	viper.Set("jwt.access_secret", "access-secret-at-least-32-bytes-long")
	viper.Set("jwt.refresh_secret", "refresh-secret-at-least-32-bytes-long")
	viper.Set("jwt.access_ttl", "15m")
	viper.Set("jwt.refresh_ttl", "168h")
	viper.Set("jwt.issuer", "ongoing-trip")

	client, err := initializeAuth(context.Background())
	if err != nil {
		t.Fatalf("initializeAuth() error = %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })
	service, err := auth.DefaultService()
	if err != nil {
		t.Fatalf("DefaultService() error = %v", err)
	}
	if _, err := service.StartSession(context.Background(), 10001); err != nil {
		t.Fatalf("StartSession() after initialization error = %v", err)
	}
}

func TestInitializeAuthRejectsMissingJWTSecrets(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)

	if _, err := initializeAuth(context.Background()); err == nil {
		t.Fatal("initializeAuth() error = nil, want missing-secret error")
	}
}
