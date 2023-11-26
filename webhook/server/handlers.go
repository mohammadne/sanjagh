package server

import (
	"context"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	admissionv1 "k8s.io/api/admission/v1"
)

func (server *Server) livenessHandler(c *fiber.Ctx) error {
	return c.SendStatus(http.StatusOK)
}

func (server *Server) readinessHandler(c *fiber.Ctx) error {
	return c.SendStatus(http.StatusOK)
}

func (server *Server) webhookHandler(c *fiber.Ctx, action func(context.Context, *admissionv1.AdmissionReview) error) error {
	request := admissionv1.AdmissionReview{}
	if err := c.BodyParser(&request); err != nil {
		server.logger.Error("Error parsing request body", zap.Any("request", request), zap.Error(err))
		return c.Status(http.StatusBadRequest).SendString("Error parsing request body")
	} else if request.Request == nil {
		server.logger.Error("admission review can't be used: Request field is nil", zap.Any("request", request), zap.Error(err))
		return c.Status(http.StatusBadRequest).SendString("AdmissionReview can't be used: Request field is nil")
	}

	if err := action(c.Context(), &request); err != nil {
		fields := []zapcore.Field{
			zap.String("resource", request.Request.Resource.String()),
			zap.String("namespace", request.Request.Namespace),
			zap.String("name", request.Request.Name),
			zap.Any("operation", request.Request.Operation),
			zap.String("user", request.Request.UserInfo.Username),
			zap.String("kind", request.Request.Kind.String()),
			zap.Error(err),
		}

		server.logger.Error("error validating resource", fields...)
		return c.Status(http.StatusBadRequest).SendString("error validating resource")
	}

	server.logger.Info("handled admission review")
	return c.Status(http.StatusOK).JSON(&request)
}

func (server *Server) validationHandler(c *fiber.Ctx) error {
	return server.webhookHandler(c, server.validation.Validate)
}

func (server *Server) mutationHandler(c *fiber.Ctx) error {
	// return server.webhookHandler(c, server.mutation.Mutate)
	return c.SendStatus(http.StatusNotImplemented)
}

func (server *Server) conversionHandler(c *fiber.Ctx) error {
	// return server.webhookHandler(c, server.conversion.Convert)
	return c.SendStatus(http.StatusNotImplemented)
}
