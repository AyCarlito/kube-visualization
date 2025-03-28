package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/AyCarlito/kube-visualization/pkg/client"
	"github.com/AyCarlito/kube-visualization/pkg/config"
	"github.com/AyCarlito/kube-visualization/pkg/graph"
	"github.com/AyCarlito/kube-visualization/pkg/visualizer"
)

func init() {
	rootCmd.AddCommand(visualizeCmd)
}

// visualizeCmd is the command for visualising resources in a Kubernetes cluster.
var visualizeCmd = &cobra.Command{
	Use:   "visualize",
	Short: "List resources in a namespace and generate a heirarchical graph of them",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.NewConfig(configurationFile)
		if err != nil {
			panic(err)
		}

		client, err := client.NewClient(client.WithLabelSelector(labelSelector))
		if err != nil {
			panic(fmt.Errorf("failed to create new client: %v", err))
		}

		return visualizer.NewVisualizer(cmd.Context(), client, cfg, graph.NewGraph(assetsBasePath, outputFile), namespace, outputFile).Visualize()
	},
}
