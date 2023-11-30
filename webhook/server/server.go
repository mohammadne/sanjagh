package server

import (
	"encoding/json"
	"fmt"

	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/mohammadne/sanjagh/webhook/validation"
)

type Server struct {
	config     *Config
	logger     *zap.Logger
	validation validation.Validation

	managementApp *fiber.App // the metrics and probe App
	masterApp     *fiber.App // the webhook App
}

func New(cfg *Config, lg *zap.Logger, validation validation.Validation) *Server {
	server := &Server{
		config:     cfg,
		logger:     lg,
		validation: validation,
	}

	fiberConfig := fiber.Config{
		JSONEncoder:           json.Marshal,
		JSONDecoder:           json.Unmarshal,
		DisableStartupMessage: true,
	}

	server.managementApp = fiber.New(fiberConfig)
	server.masterApp = fiber.New(fiberConfig)

	// Management Endpoints

	healthz := server.managementApp.Group("healthz")
	healthz.Get("/liveness", server.livenessHandler)
	healthz.Get("/readiness", server.readinessHandler)

	prometheus := fiberprometheus.New("sanjagh")
	prometheus.RegisterAt(server.managementApp, "/metrics")
	server.managementApp.Use(prometheus.Middleware)

	// Master Endpoints

	server.masterApp.Post("/validation", server.validationHandler)
	server.masterApp.Post("/mutation", server.mutationHandler)
	server.masterApp.Post("/conversion", server.conversionHandler)

	return server
}

func (server *Server) Serve(managementPort, webhookPort int) {
	go func() {
		addr := fmt.Sprintf(":%d", managementPort)
		server.logger.Info("Management server listens on", zap.String("address", addr))
		err := server.managementApp.Listen(addr)
		server.logger.Fatal("Error resolving management server", zap.Error(err))
	}()

	go func() {
		addr := fmt.Sprintf(":%d", webhookPort)
		server.logger.Info("Master (webhook) server listens on", zap.String("address", addr))
		err := server.masterApp.ListenTLS(addr, server.config.TLS.Certificate, server.config.TLS.PrivateKey)
		server.logger.Fatal("Error resolving webhook server", zap.Error(err))
	}()
}
