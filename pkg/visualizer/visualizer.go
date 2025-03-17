package visualizer

import (
	"context"
	"fmt"

	"github.com/AyCarlito/kube-visualization/pkg/client"
	"github.com/AyCarlito/kube-visualization/pkg/config"
	"github.com/AyCarlito/kube-visualization/pkg/logger"
)

// Visualizer can list namespaced resources in a Kubernetes cluster and generate graphical representations of them.
type Visualizer struct {
	ctx           context.Context
	client        *client.Client
	configuration config.Config
	namespace     string
}

// NewVisualizer returns a new *Visualizer.
func NewVisualizer(ctx context.Context, c *client.Client, cfg *config.Config, ns string) *Visualizer {
	return &Visualizer{
		ctx:           ctx,
		client:        c,
		configuration: *cfg,
		namespace:     ns,
	}
}

// Visualize gathers namespaced resources in a Kubernetes cluster and generates a graphical representation of them.
func (v *Visualizer) Visualize() error {
	log := logger.LoggerFromContext(v.ctx)
	log.Info("Gathering resources")

	for _, resource := range v.configuration.Resources {
		log.Info("Gathering: " + resource.String())
		_, err := v.client.List(v.ctx, resource, v.namespace)
		if err != nil {
			return fmt.Errorf("failed to gather %s: %v", resource.Resource, err)
		}
	}

	return nil
}
