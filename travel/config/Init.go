package config

import (
	"fmt"
	"github.com/spf13/viper"
	"os"
	"strings"
)

// @title     InitConfig
// @description     初始化配置文件
// @auth      Snactop            2023-12-5 18:01
// @return    void	没有回参
func InitConfig() error {
	workDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}
	return InitConfigFrom(workDir)
}

func InitConfigFrom(workDir string) error {
	viper.SetConfigName("application")
	viper.SetConfigType("yml")
	viper.AddConfigPath(workDir + "/config")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("read config: %w", err)
	}
	return nil
}
