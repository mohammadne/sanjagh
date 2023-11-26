package cmd

import (
	"github.com/spf13/cobra"

	"github.com/mohammadne/sanjagh/config"
)

type WebhookCommand struct {
	config      *config.Config
	metricsPort int
	probePort   int
}

func NewWebhookCommand(cfg *config.Config, metricsPort int, probePort int) *cobra.Command {
	webhookCommand := WebhookCommand{
		config:      cfg,
		metricsPort: metricsPort,
		probePort:   probePort,
	}

	cmd := &cobra.Command{
		Use:   "webhook",
		Short: "run webhook server",
		Run: func(_ *cobra.Command, _ []string) {
			webhookCommand.main()
		},
	}

	return cmd
}

func (cmd *WebhookCommand) main() {}
