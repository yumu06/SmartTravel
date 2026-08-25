package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"time"
	"travel/TravelDate"
	"travel/auth"
	"travel/cache"
	"travel/config"
	traveljwt "travel/pkg/jwt"
	"travel/pkg/snowflake"
	"travel/router"
)

func main() {
	if err := config.InitConfig(); err != nil {
		fmt.Printf("init config failed: %v\n", err)
		return
	}
	authContext, cancelAuth := context.WithTimeout(context.Background(), 2*time.Second)
	redisClient, err := initializeAuth(authContext)
	cancelAuth()
	if err != nil {
		fmt.Printf("init authentication failed: %v\n", err)
		return
	}
	defer redisClient.Close()
	if err := TravelDate.InitDB(); err != nil {
		fmt.Printf("init database failed: %v\n", err)
		return
	}

	// 雪花算法生成分布式ID
	if err := snowflake.Init(1); err != nil {
		fmt.Printf("init snowflake failed, err:%v\n", err)
		return
	}

	r := newEngine(viper.GetBool("server.access_log"))
	r = router.NewRouter(r)
	port := viper.GetString("server.port")
	if port != "" {
		panic(r.Run(":" + port))
	} else {
		panic(r.Run())
	}
}

func newEngine(accessLog bool) *gin.Engine {
	engine := gin.New()
	if accessLog {
		engine.Use(gin.Logger())
	}
	engine.Use(gin.Recovery())
	return engine
}

func initializeAuth(ctx context.Context) (*redis.Client, error) {
	manager, err := traveljwt.NewManager(traveljwt.Config{
		AccessSecret:  []byte(viper.GetString("jwt.access_secret")),
		RefreshSecret: []byte(viper.GetString("jwt.refresh_secret")),
		AccessTTL:     viper.GetDuration("jwt.access_ttl"),
		RefreshTTL:    viper.GetDuration("jwt.refresh_ttl"),
		Issuer:        viper.GetString("jwt.issuer"),
	})
	if err != nil {
		return nil, err
	}
	redisClient := cache.NewRedisClient(cache.RedisConfig{
		Addr:         viper.GetString("redis.addr"),
		Password:     viper.GetString("redis.password"),
		DB:           viper.GetInt("redis.db"),
		PoolSize:     viper.GetInt("redis.pool_size"),
		MinIdleConns: viper.GetInt("redis.min_idle_conns"),
		DialTimeout:  viper.GetDuration("redis.dial_timeout"),
		ReadTimeout:  viper.GetDuration("redis.read_timeout"),
		WriteTimeout: viper.GetDuration("redis.write_timeout"),
	})
	if err := cache.Ping(ctx, redisClient); err != nil {
		_ = redisClient.Close()
		return nil, fmt.Errorf("ping Redis: %w", err)
	}
	traveljwt.SetDefaultManager(manager)
	auth.SetDefaultService(auth.NewService(manager, auth.NewRedisSessionStore(redisClient)))
	return redisClient, nil
}
