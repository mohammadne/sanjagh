package cmd

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"sigs.k8s.io/controller-runtime/pkg/cache"

	"github.com/mohammadne/sanjagh/config"
	"github.com/mohammadne/sanjagh/pkg/k8s"
	"github.com/mohammadne/sanjagh/pkg/logger"
	"github.com/mohammadne/sanjagh/webhook/server"
	"github.com/mohammadne/sanjagh/webhook/validation"
)

type Webhook struct {
	config        *config.Config
	managmentPort int
	masterPort    int
	kubeconfig    string
}

func NewWebhook(cfg *config.Config) *cobra.Command {
	webhook := Webhook{config: cfg}

	cmd := &cobra.Command{
		Use:   "webhook",
		Short: "run webhook server",
		Run: func(_ *cobra.Command, _ []string) {
			webhook.main()
		},
	}

	cmd.Flags().IntVar(&webhook.managmentPort, "managment-bind-address", 8080, "The port the metric and probe endpoints binds to")
	cmd.Flags().IntVar(&webhook.masterPort, "master-bind-address", 8081, "The port the webhook server listens on")
	cmd.Flags().StringVar(&webhook.kubeconfig, "kubeconfig", "", "The kubeconfig file path")

	return cmd
}

func (cmd *Webhook) main() {
	logger := logger.NewZap(cmd.config.Logger)

	kubeConfig, err := k8s.KubeConfig(cmd.kubeconfig)
	if err != nil {
		logger.Fatal("Unable to create kubernetes configuration", zap.Error(err))
	}

	client, err := k8s.NewCachedClient(kubeConfig, indexer)
	if err != nil {
		logger.Fatal("Couldn't create cached client", zap.Error(err))
	}

	validation := validation.NewValidation(&cmd.config.Webhook.Validation, client)

	trap := make(chan os.Signal, 1)
	signal.Notify(trap, syscall.SIGINT, syscall.SIGTERM)

	server.New(&cmd.config.Webhook.Server, logger, validation).
		Serve(cmd.managmentPort, cmd.masterPort)

	// Keep this at the bottom of the main function
	field := zap.String("signal trap", (<-trap).String())
	logger.Info("exiting by receiving a unix signal", field)
}

// indexer adds indexers for given cached client
func indexer(cache cache.Cache) {}
