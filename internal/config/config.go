package config

import (
	"github.com/mohammadne/sanjagh/pkg/logger"
)

type Config struct {
	MetricsPort    int  `koanf:"metrics_port"`
	ProbePort      int  `koanf:"probe_port"`
	LeaderElection bool `koanf:"leader_election"`

	Logger *logger.Config `koanf:"logger"`
}
