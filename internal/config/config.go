package config

import (
	"sync"
	"user-balance-service/pkg/logging"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Listen struct {
		BindIp string `yaml:"bind_ip" env-default:"127.0.0.1"`
		Port   string `yaml:"port" env-default:"8080"`
	}
}

var instance *Config

var once sync.Once

func GetConfig() *Config {
	once.Do(func() {
		logger := logging.NewLogger()
		logger.Info("Read application configuration")
		instance = &Config{}
		if err := cleanenv.ReadConfig("config.yml", instance); err != nil {
			help, _ := cleanenv.GetDescription(instance, nil)
			logger.Fatal(help)
		}
	})
	return instance
}
