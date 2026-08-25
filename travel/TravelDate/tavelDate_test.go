package TravelDate

import (
	"testing"
	"time"
)

type recordingPool struct {
	maxOpen     int
	maxIdle     int
	maxLifetime time.Duration
	maxIdleTime time.Duration
}

func (p *recordingPool) SetMaxOpenConns(n int)              { p.maxOpen = n }
func (p *recordingPool) SetMaxIdleConns(n int)              { p.maxIdle = n }
func (p *recordingPool) SetConnMaxLifetime(d time.Duration) { p.maxLifetime = d }
func (p *recordingPool) SetConnMaxIdleTime(d time.Duration) { p.maxIdleTime = d }

func TestConfigurePoolAppliesAllLimits(t *testing.T) {
	pool := &recordingPool{}
	cfg := PoolConfig{
		MaxOpenConns:    50,
		MaxIdleConns:    10,
		ConnMaxLifetime: 30 * time.Minute,
		ConnMaxIdleTime: 10 * time.Minute,
	}

	ConfigurePool(pool, cfg)

	if pool.maxOpen != 50 || pool.maxIdle != 10 {
		t.Fatalf("connection counts = (%d, %d), want (50, 10)", pool.maxOpen, pool.maxIdle)
	}
	if pool.maxLifetime != 30*time.Minute || pool.maxIdleTime != 10*time.Minute {
		t.Fatalf("connection durations = (%s, %s), want (30m, 10m)", pool.maxLifetime, pool.maxIdleTime)
	}
}
