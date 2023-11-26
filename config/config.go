package config

import (
	"github.com/mohammadne/sanjagh/pkg/logger"
)

type Config struct {
	Logger *logger.Config `koanf:"logger"`
}
