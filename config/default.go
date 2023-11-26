package config

import (
	"github.com/mohammadne/sanjagh/pkg/logger"
)

func defaultConfig() *Config {
	return &Config{
		Logger: &logger.Config{
			Development: true,
			Level:       "info",
			Encoding:    "console",
		},
	}
}
