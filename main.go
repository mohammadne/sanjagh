package main

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/mohammadne/sanjagh/cmd"
	"github.com/mohammadne/sanjagh/config"
)

func main() {
	const description = "Sanjagh operator"
	root := &cobra.Command{Short: description}

	var (
		config      = config.Load(true)
		metricsPort int
		probePort   int
	)

	root.Flags().IntVar(&metricsPort, "metrics-bind-address", 8080, "The port the metric endpoint binds to")
	root.Flags().IntVar(&probePort, "health-probe-bind-address", 8081, "The port the probe endpoint binds to")

	root.AddCommand(
		cmd.NewManagerCommand(config, metricsPort, probePort),
		cmd.NewWebhookCommand(config, metricsPort, probePort),
	)

	if err := root.Execute(); err != nil {
		log.Fatal(err.Error(), "failed to execute root command")
	}
}
