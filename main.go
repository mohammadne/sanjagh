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

	config := config.Load(true)

	root.AddCommand(
		cmd.NewManager(config),
		cmd.NewWebhook(config),
	)

	if err := root.Execute(); err != nil {
		log.Fatal(err.Error(), "failed to execute root command")
	}
}
