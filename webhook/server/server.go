package server

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	admissionv1 "k8s.io/api/admission/v1"

	"github.com/mohammadne/sanjagh/webhook/validation"
)

type Server struct {
	config     *Config
	logger     *zap.Logger
	validation validation.Validation
}

func New(cfg *Config, lg *zap.Logger, validation validation.Validation) (*Server, error) {
	return &Server{
		config:     cfg,
		validation: validation,
		logger:     lg.Named("webhook-server"),
	}, nil
}

func (s *Server) Run() error {
	router := &mux.Router{}
	router.HandleFunc(s.config.Path, s.handle)
	server := http.Server{Addr: s.config.ListenAddr, Handler: router}
	return server.ListenAndServeTLS(s.config.TLSCert, s.config.TLSKey)
}

func (s *Server) handle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	admissionReview, err := parseRequest(*r)
	if err != nil {
		s.handleError(err, "could not parse admission review request", []zapcore.Field{}, w)
		return
	}

	if err := s.validation.Validate(ctx, admissionReview); err != nil {
		fields := commonLoggerFields(admissionReview)
		s.handleError(err, "error validating resource", fields, w)
		return
	}

	if err := writeResponse(admissionReview, w); err != nil {
		fields := resultLoggerFields(admissionReview)
		s.handleError(err, "could not write response", fields, w)
		return
	}

	s.logger.Info("handled admission review")
}

func (s *Server) handleError(err error, message string, fields []zapcore.Field, w http.ResponseWriter) {
	s.logger.Error(message, append(fields, zap.Error(err))...)
	http.Error(w, fmt.Sprintf("%s: %v", message, err), http.StatusBadRequest)
}

func commonLoggerFields(ar *admissionv1.AdmissionReview) []zapcore.Field {
	return []zapcore.Field{
		zap.String("resource", ar.Request.Resource.String()),
		zap.String("namespace", ar.Request.Namespace),
		zap.String("name", ar.Request.Name),
		zap.Any("operation", ar.Request.Operation),
		zap.String("user", ar.Request.UserInfo.Username),
		zap.String("kind", ar.Request.Kind.String()),
	}
}

func resultLoggerFields(ar *admissionv1.AdmissionReview) []zapcore.Field {
	return []zapcore.Field{
		zap.Bool("allowed", ar.Response.Allowed),
		zap.String("reason", ar.Response.Result.Message),
	}
}
