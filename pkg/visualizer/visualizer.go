package visualizer

import (
	"context"
	"fmt"
	"strconv"

	"github.com/awalterschulze/gographviz"

	"github.com/AyCarlito/kube-visualization/pkg/client"
	"github.com/AyCarlito/kube-visualization/pkg/config"
	"github.com/AyCarlito/kube-visualization/pkg/logger"
)

// getSubgraphName returns the name of a subgraph in a gographviz.Graph.
func getSubgraphName(i int) string {
	return fmt.Sprintf("rank_%s", fmt.Sprintf("%04s", strconv.Itoa(i)))
}

// getDummyNodeName returns the name of a dummy node for use in a gographviz.Graph.
func getDummyNodeName(i int) string {
	return fmt.Sprintf("node_%s", fmt.Sprintf("%04s", strconv.Itoa(i)))
}

// getImagePath returns the path to an image for a given resource.
func getImagePath(resource string) string {
	return fmt.Sprintf("\"./assets/%s.png\"", resource)
}

// getNodeLabel returns the label of a node in a gographviz.Graph.
// By default, the name of the node is used for the label which is then placed at the centre of the node.
// Here, we keep the name but prepend newlines so that it is displayed below the node.
func getNodeLabel(node string) string {
	return fmt.Sprintf("\"\\n\\n\\n\\n\\n\\n\\n%s\"", node)
}

// getSanitizedObjectName returns the sanitized name of an object in a gographviz.Graph.
// The provided name and kind are wrapped in double quotes.
func getSanitizedObjectName(name, kind string) string {
	return fmt.Sprintf("\"%s_%s\"", kind, name)
}

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

	g := newSkeletonGraph("Visualization", v.namespace, config.UniqueRanks(v.configuration.Resources))
	for i, resource := range v.configuration.Resources {
		log.Info("Gathering: " + resource.String())
		pomlList, err := v.client.List(v.ctx, resource.GroupVersionResource, v.namespace)
		if err != nil {
			return fmt.Errorf("failed to gather %s: %v", resource.Resource, err)
		}
		for _, poml := range pomlList.Items {
			g.AddNode(getSubgraphName(i), getSanitizedObjectName(poml.Name, poml.Kind), map[string]string{
				"penwidth": "0",
				"label":    getNodeLabel(poml.Name),
				"image":    getImagePath(resource.Resource),
			})

		}
	}

	fmt.Println(g.String())
	return nil
}

// newSkeletonGraph returns a skeleton *gographviz.Graph for later population.
// The skeleton is composed of:
//   - Basic object metadata.
//   - A single subgraph representing the namespace to be visualised.
//   - A subgraph for each unique rank in the GVRs to be retrieved.
//   - An invisble node in each rank subgraph.
//   - Invisible edges connecting the invisble nodes across the rank subgraphs.
func newSkeletonGraph(name, namespace string, numSubgraphs int) *gographviz.Graph {
	g := gographviz.NewGraph()
	// In a directed graph, the arrows between nodes have a direction.
	// Direction indicates ownership, and reflects the owner references stored on the Kubernetes object.
	g.SetDir(true)
	g.SetName(name)
	// Setting the heirarchy here, Top to bottom.
	g.AddAttr(name, "rankdir", "TB")

	// Highest level subgraph for the namespace.
	g.AddSubGraph(name, namespace, map[string]string{
		"style": "dotted",
	})

	g.AddNode(namespace, getSanitizedObjectName(namespace, "namespace"), map[string]string{
		"penwidth": "0",
		"height":   "0",
		"width":    "0",
		"margin":   "0",
		"label":    getNodeLabel(namespace),
		"image":    getImagePath("namespaces"),
	})

	// A subgraph within the namespace subgraph for each kind of resource.
	for i := range numSubgraphs {
		g.AddSubGraph(namespace, getSubgraphName(i), map[string]string{
			"rank":  "same",
			"style": "invis",
		})

		// A dummy node in subgraph.
		g.AddNode(getSubgraphName(i), getDummyNodeName(i), map[string]string{
			"style":  "invis",
			"height": "0",
			"width":  "0",
			"margin": "0",
		})

	}

	// Each dummy node is connected with an invisible edge.
	// Note the index here is offset by 1 as the final node cannot be the source node for a connection as there
	// is no destination node to connect it to!
	for i := 0; i < (numSubgraphs - 1); i++ {
		g.AddEdge(getDummyNodeName(i), getDummyNodeName(i+1), true, map[string]string{"style": "invis"})
	}

	// Connect the namespace node to the first dummy node.
	g.AddEdge(getSanitizedObjectName(namespace, "namespace"), getDummyNodeName(0), true, map[string]string{"style": "invis"})

	return g
}
