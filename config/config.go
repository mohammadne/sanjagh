package config

import (
	"github.com/mohammadne/sanjagh/pkg/logger"
	webhookServer "github.com/mohammadne/sanjagh/webhook/server"
	webhookValidation "github.com/mohammadne/sanjagh/webhook/validation/config"
)

type Config struct {
	Logger  *logger.Config `koanf:"logger"`
	Webhook struct {
		Server     webhookServer.Config     `koanf:"server"`
		Validation webhookValidation.Config `koanf:"validation"`
	} `koanf:"webhook"`
}
