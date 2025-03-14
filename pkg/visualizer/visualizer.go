package visualizer

import (
	"fmt"

	"github.com/AyCarlito/kube-visualization/pkg/client"
	"github.com/AyCarlito/kube-visualization/pkg/config"
)

// Visualizer can list namespaced resources in a Kubernetes cluster and generate graphical representations of them.
type Visualizer struct {
	client        *client.Client
	configuration config.Config
}

// NewVisualizer returns a new *Visualizer.
func NewVisualizer(c *client.Client, cfg *config.Config) *Visualizer {
	return &Visualizer{
		client:        c,
		configuration: *cfg,
	}
}

// Visualize gathers namespaced resources in a Kubernetes cluster and generates a graphical representation of them.
func (v *Visualizer) Visualize() error {
	if err := v.gather(); err != nil {
		return fmt.Errorf("failed to gather resources: %v", err)
	}
	if err := v.graph(); err != nil {
		return fmt.Errorf("failed to graph resources: %v", err)
	}
	return nil
}

// gather iterates over a slice of GroupVersionKind.
func (v *Visualizer) gather() error {
	return nil
}

func (v *Visualizer) graph() error {
	return nil
}
