package logger

import (
	"log"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewNoop() *zap.Logger {
	return zap.NewNop()
}

func NewZap(cfg *Config) *zap.Logger {
	return zap.New(
		zapcore.NewCore(Encoder(cfg), WriteSyncer(cfg), LoggerLevel(cfg)),
		Options(cfg)...,
	)
}

func Encoder(cfg *Config) zapcore.Encoder {
	var encoderConfig zapcore.EncoderConfig
	if cfg.Development {
		encoderConfig = zap.NewDevelopmentEncoderConfig()
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		encoderConfig = zap.NewProductionEncoderConfig()
		encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	}

	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	var encoder zapcore.Encoder
	if cfg.Encoding == "console" {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	return encoder
}

func WriteSyncer(cfg *Config) zapcore.WriteSyncer {
	return zapcore.Lock(os.Stdout)
}

func LoggerLevel(cfg *Config) zap.AtomicLevel {
	var level zapcore.Level

	if err := level.Set(cfg.Level); err != nil {
		log.Printf("using debug level for zap due to an error in user's config value %s", cfg.Level)
		return zap.NewAtomicLevelAt(zapcore.DebugLevel)
	}

	return zap.NewAtomicLevelAt(level)
}

func Options(cfg *Config) []zap.Option {
	return []zap.Option{
		zap.AddStacktrace(zap.ErrorLevel),
		zap.AddCaller(),
	}
}
