package logger

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap/zapcore"
)

func TestHandlers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Logger Suite")
}

var _ = Describe("Logger Creation", func() {
	It("calling a method on the mock object has no error", func() {
		lg := NewNoop()
		lg.Error("It shouldn't panic bu throwing nil pointer derefrence")
	})

	It("load normal config", func() {
		cfg := &Config{
			Development: true,
			Encoding:    "console",
			Level:       "debug",
		}

		Expect(NewZap(cfg).Level()).Should(Equal(zapcore.DebugLevel))
	})

	It("invalid log level", func() {
		cfg := &Config{
			Development: true,
			Encoding:    "console",
			Level:       "invalid level",
		}

		Expect(NewZap(cfg).Level()).Should(Equal(zapcore.DebugLevel))
	})

	It("valid json encoding", func() {
		cfg := &Config{
			Development: false,
			Encoding:    "json",
			Level:       "info",
		}

		Expect(NewZap(cfg).Level()).Should(Equal(zapcore.InfoLevel))
	})
})
