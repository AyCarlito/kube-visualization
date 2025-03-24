package visualizer

import (
	"context"
	"fmt"

	"github.com/AyCarlito/kube-visualization/pkg/client"
	"github.com/AyCarlito/kube-visualization/pkg/config"
	"github.com/AyCarlito/kube-visualization/pkg/graph"
	"github.com/AyCarlito/kube-visualization/pkg/logger"
)

// Visualizer can list namespaced resources in a Kubernetes cluster and generate graphical representations of them.
type Visualizer struct {
	ctx            context.Context
	client         *client.Client
	configuration  config.Config
	grapher        *graph.Grapher
	namespace      string
	outputFilePath string
}

// NewVisualizer returns a new *Visualizer.
func NewVisualizer(ctx context.Context, c *client.Client, cfg *config.Config, g *graph.Grapher, ns, ofp string) *Visualizer {
	return &Visualizer{
		ctx:            ctx,
		client:         c,
		configuration:  *cfg,
		grapher:        g,
		namespace:      ns,
		outputFilePath: ofp,
	}
}

// Visualize gathers namespaced resources in a Kubernetes cluster and generates a graphical representation of them.
func (v *Visualizer) Visualize() error {
	log := logger.LoggerFromContext(v.ctx)

	v.grapher.Scaffold("Visualization", v.namespace, config.SortedUniqueRanks(v.configuration.Resources))
	for _, resource := range v.configuration.Resources {
		log.Info("Gathering: " + resource.String())
		pomlList, err := v.client.List(v.ctx, resource.GroupVersionResource, v.namespace)
		if err != nil {
			return fmt.Errorf("failed to gather %s: %v", resource.Resource, err)
		}
		v.grapher.Populate(pomlList, resource)
	}

	log.Info("Connecting related resources")
	v.grapher.Connect()

	// Write the string representation of the graph to file.
	log.Info("Writing to file: " + v.outputFilePath)
	err := v.grapher.WriteDotFile()
	if err != nil {
		return fmt.Errorf("failed to write graph to output dot file: %v", err)
	}

	return nil
}
