package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func TestInitConfigFromEnvironmentOverridesSecrets(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)

	dir := t.TempDir()
	configDir := filepath.Join(dir, "config")
	if err := os.Mkdir(configDir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := []byte("server:\n  port: 1016\njwt:\n  access_secret: file-value\n")
	if err := os.WriteFile(filepath.Join(configDir, "application.yml"), content, 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("JWT_ACCESS_SECRET", "environment-value")
	t.Setenv("MYSQL_PASSWORD", "mysql-environment-value")
	t.Setenv("REDIS_PASSWORD", "redis-environment-value")
	t.Setenv("WX_APPSECRET", "wx-environment-value")
	t.Setenv("TENCENT_WEBSERVICE_API", "tencent-environment-value")

	if err := InitConfigFrom(dir); err != nil {
		t.Fatalf("InitConfigFrom() error = %v", err)
	}
	if got := viper.GetString("jwt.access_secret"); got != "environment-value" {
		t.Fatalf("jwt.access_secret = %q, want environment override", got)
	}
	checks := map[string]string{
		"mysql.password":         "mysql-environment-value",
		"redis.password":         "redis-environment-value",
		"wx.appSecret":           "wx-environment-value",
		"tencent.webservice_api": "tencent-environment-value",
	}
	for key, want := range checks {
		if got := viper.GetString(key); got != want {
			t.Fatalf("%s = %q, want %q", key, got, want)
		}
	}
}

func TestInitConfigFromMissingFileReturnsError(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)

	if err := InitConfigFrom(t.TempDir()); err == nil {
		t.Fatal("InitConfigFrom() error = nil, want missing-config error")
	}
}
