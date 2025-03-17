package cmd

import (
	"fmt"
	"os"

	"github.com/AyCarlito/kube-visualization/pkg/logger"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func init() {
	rootCmd.PersistentFlags().StringVar(&configurationFile, "config", "config/config.json", "Path to configuration file.")
}

// CLI Flags
var configurationFile string

var rootCmd = &cobra.Command{
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

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
