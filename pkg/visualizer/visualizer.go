package visualizer

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/awalterschulze/gographviz"

	"github.com/AyCarlito/kube-visualization/pkg/client"
	"github.com/AyCarlito/kube-visualization/pkg/config"
	"github.com/AyCarlito/kube-visualization/pkg/logger"
)

// referenced is a Kubernetes object that either owns or is owned by another resource.
type referenced struct {
	ownerName string
	ownerKind string
	ownedKind string
}

// ownedToOwner maps an owned Kubenetes object name to the object that owns it.
// Ownership is determined based on the ownerReferences present in object.
var ownedToOwner = map[string]referenced{}

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
	ctx            context.Context
	client         *client.Client
	configuration  config.Config
	namespace      string
	outputFilePath string
}

// NewVisualizer returns a new *Visualizer.
func NewVisualizer(ctx context.Context, c *client.Client, cfg *config.Config, ns string, ofp string) *Visualizer {
	return &Visualizer{
		ctx:            ctx,
		client:         c,
		configuration:  *cfg,
		namespace:      ns,
		outputFilePath: ofp,
	}
}

// Visualize gathers namespaced resources in a Kubernetes cluster and generates a graphical representation of them.
func (v *Visualizer) Visualize() error {
	log := logger.LoggerFromContext(v.ctx)

	g := newSkeletonGraph("Visualization", v.namespace, config.SortedUniqueRanks(v.configuration.Resources))
	for _, resource := range v.configuration.Resources {
		log.Info("Gathering: " + resource.String())
		pomlList, err := v.client.List(v.ctx, resource.GroupVersionResource, v.namespace)
		if err != nil {
			return fmt.Errorf("failed to gather %s: %v", resource.Resource, err)
		}
		// Add a node for each object in the List. The node is added to the
		for _, poml := range pomlList.Items {
			g.AddNode(getSubgraphName(resource.Rank), getSanitizedObjectName(poml.Name, poml.Kind), map[string]string{
				"penwidth": "0",
				"label":    getNodeLabel(poml.Name),
				"image":    getImagePath(resource.Resource),
			})
			// If the object contains a controlling owner reference, track it.
			// We do this so an edge can be constructed to link the object node to the owner node.
			// Ideally, we would skip the tracking and just create the edge now. But the owner node may not exist at
			// this point.
			if len(poml.OwnerReferences) > 0 && poml.OwnerReferences[0].Controller != nil && *poml.OwnerReferences[0].Controller {
				ownedToOwner[poml.Name] = referenced{
					ownerName: poml.OwnerReferences[0].Name,
					ownerKind: poml.OwnerReferences[0].Kind,
					ownedKind: poml.Kind,
				}
			}
		}
	}

	log.Info("Graphing gathered resources")
	// Now create the edges for any ownership relationships that have been tracked.
	for ownedName, ref := range ownedToOwner {
		// It's possible that the ownership reference may be towards a resource that isn't part of this visualisation.
		// Check for the existence of the owner node first, and skip if it does not exist.
		if _, ok := g.Nodes.Lookup[getSanitizedObjectName(ref.ownerName, ref.ownerKind)]; !ok {
			continue
		}
		g.AddEdge(getSanitizedObjectName(ref.ownerName, ref.ownerKind), getSanitizedObjectName(ownedName, ref.ownedKind), true, map[string]string{
			"style": "dashed",
		})

	}

	// Write the string representation of the graph to file.
	log.Info("Writing to file: " + v.outputFilePath)
	err := writeGraphToDotFile(g, v.outputFilePath)
	if err != nil {
		return fmt.Errorf("failed to write graph to output dot file: %v", err)
	}

	return nil
}

// newSkeletonGraph returns a skeleton *gographviz.Graph for later population.
// The skeleton is composed of:
//   - Basic object metadata.
//   - A single subgraph representing the namespace to be visualised.
//   - A subgraph for each unique rank in the GVRs to be retrieved.
//   - An invisble node in each rank subgraph.
//   - Invisible edges connecting the invisble nodes across the rank subgraphs.
func newSkeletonGraph(name, namespace string, ranks []int) *gographviz.Graph {
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
	for _, i := range ranks {
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
	for i := 0; i < (len(ranks) - 1); i++ {
		g.AddEdge(getDummyNodeName(ranks[i]), getDummyNodeName(ranks[i+1]), true, map[string]string{"style": "invis"})
	}

	// Connect the namespace node to the first dummy node.
	g.AddEdge(getSanitizedObjectName(namespace, "namespace"), getDummyNodeName(ranks[0]), true, map[string]string{"style": "invis"})

	return g
}

// writeGraphToDotFile writes the string representation of a provided *gographviz.Graph to the provided output file path.
func writeGraphToDotFile(g *gographviz.Graph, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create dot file: %v", err)
	}
	defer file.Close()

	_, err = file.WriteString(g.String())
	if err != nil {
		return fmt.Errorf("failed to write to dot file: %v", err)
	}
	return nil
}
