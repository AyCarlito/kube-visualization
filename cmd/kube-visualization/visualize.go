package main

import (
	"fmt"

	"github.com/AyCarlito/kube-visualization/pkg/client"
	"github.com/AyCarlito/kube-visualization/pkg/config"
	"github.com/AyCarlito/kube-visualization/pkg/visualizer"
	"github.com/spf13/cobra"
)

// visualizeCmd is the command for visualising resources in a Kubernetes cluster.
var visualizeCmd = &cobra.Command{
	Use:   "visualize",
	Short: "List resources in a namespace and generate a heirarchical graph of them",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.NewConfig(configurationFile)
		if err != nil {
			panic(err)
		}

		client, err := client.NewClient()
		if err != nil {
			panic(fmt.Errorf("failed to create new client: %v", err))
		}

		// TODO: Any from here should be logged properly.
		return visualizer.NewVisualizer(client, cfg).Visualize()
	},
}
