package server

import (
	"encoding/json"
	"fmt"

	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"go.uber.org/zap"

	"github.com/mohammadne/sanjagh/webhook/validation"
)

type Server struct {
	config     *Config
	logger     *zap.Logger
	validation validation.Validation

	managmentApp *fiber.App
	webhookApp   *fiber.App
}

func New(cfg *Config, lg *zap.Logger, validation validation.Validation) *Server {
	server := &Server{
		config:     cfg,
		logger:     lg,
		validation: validation,
	}

	// Managment Endpoints

	server.managmentApp = fiber.New(fiber.Config{JSONEncoder: json.Marshal, JSONDecoder: json.Unmarshal})
	server.managmentApp.Use(cors.New())

	healthz := server.managmentApp.Group("healthz")
	healthz.Get("/liveness", server.livenessHandler)
	healthz.Get("/readiness", server.readinessHandler)

	prometheus := fiberprometheus.New("sanjagh")
	prometheus.RegisterAt(server.managmentApp, "/metrics")
	server.managmentApp.Use(prometheus.Middleware)

	// Webhook Endpoints

	server.webhookApp = fiber.New(fiber.Config{JSONEncoder: json.Marshal, JSONDecoder: json.Unmarshal})
	server.webhookApp.Use(cors.New())

	server.webhookApp.Post("/validation", server.validationHandler)
	server.webhookApp.Post("/mutation", server.mutationHandler)
	server.webhookApp.Post("/conversion", server.conversionHandler)

	return server
}

func (server *Server) Serve(managmentPort, webhookPort int) {
	go func() {
		addr := fmt.Sprintf(":%d", managmentPort)
		err := server.managmentApp.ListenTLS(addr, server.config.TLSCert, server.config.TLSKey)
		server.logger.Fatal("error resolving managment server", zap.Error(err))
	}()

	go func() {
		addr := fmt.Sprintf(":%d", webhookPort)
		err := server.webhookApp.ListenTLS(addr, server.config.TLSCert, server.config.TLSKey)
		server.logger.Fatal("error resolving webhook server", zap.Error(err))
	}()
}
