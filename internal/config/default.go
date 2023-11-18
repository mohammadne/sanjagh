package config

import (
	"github.com/mohammadne/sanjagh/pkg/logger"
)

func defaultConfig() *Config {
	return &Config{
		MetricsPort:    8080,
		ProbePort:      8081,
		LeaderElection: false,

		Logger: &logger.Config{
			Development: true,
			Level:       "info",
			Encoding:    "console",
		},
	}
}
