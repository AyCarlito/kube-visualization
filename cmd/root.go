package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/AyCarlito/kube-visualization/pkg/logger"
)

func init() {
	rootCmd.PersistentFlags().StringVar(&assetsBasePath, "assets", "assets/", "Path to assets directory.")
	rootCmd.PersistentFlags().StringVar(&configurationFile, "config", "config/config.json", "Path to configuration file.")
	rootCmd.PersistentFlags().StringVar(&namespace, "namespace", "default", "Namespace of resources.")
	rootCmd.PersistentFlags().StringVar(&outputFile, "output", "assets/output.dot", "Path to output file.")
	rootCmd.PersistentFlags().StringVar(&labelSelector, "label-selector", "", "Filter resources by label. Comma separated key-value pairs.")
	rootCmd.PersistentFlags().StringVar(&kubeConfigPath, "kubeconfig", "", "Path to a kubeconfig file.")
}

// CLI Flags
var (
	assetsBasePath    string
	configurationFile string
	outputFile        string
	namespace         string
	labelSelector     string
	kubeConfigPath    string
)

var rootCmd = &cobra.Command{
	Use:           "kube-visualization",
	Short:         "Allows resources in a given namespace in a Kubernetes cluster to be visualized.",
	SilenceErrors: true,
	SilenceUsage:  true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Build logger.
		log, err := logger.NewZapConfig().Build()
		if err != nil {
			panic(fmt.Errorf("failed to build zap logger: %v", err))
		}
		cmd.SetContext(logger.ContextWithLogger(cmd.Context(), log))
		cmd.Parent().SetContext(logger.ContextWithLogger(cmd.Parent().Context(), log))
		return nil
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		// By default, cobra prints the error and usage string on every error.
		// We only desire this behaviour in the case where command line parsing fails e.g. unknown command or flag.
		// Cobra does not provide a mechanism for achieving this fine grain control, so we implement our own.
		if strings.Contains(err.Error(), "command") || strings.Contains(err.Error(), "flag") {
			// Parsing errors are printed along with the usage string.
			fmt.Println(err.Error())
			fmt.Println(rootCmd.UsageString())
		} else {
			// Other errors logged, no usage string displayed.
			log := logger.LoggerFromContext(rootCmd.Context())
			log.Error(err.Error())
		}
		os.Exit(1)
	}
}
