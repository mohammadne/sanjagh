package main

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/mohammadne/sanjagh/cmd"
)

func main() {
	const description = "Sanjagh operator"
	root := &cobra.Command{Short: description}

	root.AddCommand(
		cmd.Manager{}.Command(),
		cmd.Webhook{}.Command(),
	)

	if err := root.Execute(); err != nil {
		log.Fatal(err.Error(), "failed to execute root command")
	}
}
