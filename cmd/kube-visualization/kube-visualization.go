package main

import (
	"fmt"
	"os"

	"github.com/AyCarlito/kube-visualization/pkg/logger"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// CLI Flags
var configurationFile string

func main() {
	root := &cobra.Command{
		Use:   "kube-visualization",
		Short: "Allows resources in a given namespace in a Kubernetes cluster to be visualised",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Build zap logger using default production configuration.
			log, err := zap.NewProductionConfig().Build()
			if err != nil {
				panic(fmt.Errorf("failed to build zap logger: %v", err))
			}
			cmd.SetContext(logger.ContextWithLogger(cmd.Context(), log))
			return nil
		},
	}

	root.AddCommand(visualizeCmd)
	root.PersistentFlags().StringVar(&configurationFile, "config", "config/config.json", "Path to configuration file.")
	err := root.ExecuteContext(root.Context())
	if err != nil {
		os.Exit(1)
	}
}
