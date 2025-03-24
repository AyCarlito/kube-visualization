package graph

import (
	"fmt"
	"os"
	"strconv"

	"github.com/awalterschulze/gographviz"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/AyCarlito/kube-visualization/pkg/config"
)

// Grapher creates gographviz graphs.
type Grapher struct {
	connections    []connection
	graph          *gographviz.Graph
	outputFilePath string
}

// NewGrapher returns a new *Grapher.
func NewGraph(o string) *Grapher {
	return &Grapher{outputFilePath: o}
}

// connection is a link between two Kubernetes objects.
type connection struct {
	sourceName      string
	sourceKind      string
	destinationName string
	destinationKind string
}

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
	return fmt.Sprintf("\"\\n\\n\\n\\n\\n\\n\\n\\n\\n%s\"", node)
}

// getSanitizedObjectName returns the sanitized name of an object in a gographviz.Graph.
// The provided name and kind are wrapped in double quotes.
func getSanitizedObjectName(name, kind string) string {
	return fmt.Sprintf("\"%s_%s\"", kind, name)
}

// Scaffold scaffolds the graph for later population.
// The scaffold is composed of:
//   - Basic object metadata.
//   - A single subgraph representing the namespace to be visualised.
//   - A subgraph for each unique rank in the GVRs to be retrieved.
//   - An invisble node in each rank subgraph.
//   - Invisible edges connecting the invisble nodes across the rank subgraphs.
func (g *Grapher) Scaffold(name, namespace string, ranks []int) {
	graph := gographviz.NewGraph()
	// In a directed graph, the arrows between nodes have a direction.
	// Direction indicates ownership, and reflects the owner references stored on the Kubernetes object.
	graph.SetDir(true)
	graph.SetName(name)
	// Setting the heirarchy here, Top to bottom.
	graph.AddAttr(name, "rankdir", "TB")

	// Highest level subgraph for the namespace.
	graph.AddSubGraph(name, getSanitizedObjectName(namespace, "namespace"), map[string]string{
		"style": "dotted",
	})

	graph.AddNode(getSanitizedObjectName(namespace, "namespace"), getSanitizedObjectName(namespace, "namespace"), map[string]string{
		"penwidth": "0",
		"height":   "0",
		"width":    "0",
		"margin":   "0",
		"label":    getNodeLabel(namespace),
		"image":    getImagePath("namespaces"),
	})

	// A subgraph within the namespace subgraph for each kind of resource.
	for _, i := range ranks {
		graph.AddSubGraph(getSanitizedObjectName(namespace, "namespace"), getSubgraphName(i), map[string]string{
			"rank":  "same",
			"style": "invis",
		})

		// A dummy node in subgraph.
		graph.AddNode(getSubgraphName(i), getDummyNodeName(i), map[string]string{
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
		graph.AddEdge(getDummyNodeName(ranks[i]), getDummyNodeName(ranks[i+1]), true, map[string]string{"style": "invis"})
	}

	// Connect the namespace node to the first dummy node.
	graph.AddEdge(getSanitizedObjectName(namespace, "namespace"), getDummyNodeName(ranks[0]), true, map[string]string{"style": "invis"})

	g.graph = graph
}

// Connect connects related nodes in the graph.
func (g *Grapher) Connect() {
	// Now create the edges for any connections that have been tracked.
	for _, connection := range g.connections {
		// It's possible that the connection may be towards a resource that isn't part of this visualisation.
		// Check for the existence of the source node first, and skip if it does not exist.
		if _, ok := g.graph.Nodes.Lookup[getSanitizedObjectName(connection.sourceName, connection.sourceKind)]; !ok {
			continue
		}
		g.graph.AddEdge(
			getSanitizedObjectName(connection.sourceName, connection.sourceKind),
			getSanitizedObjectName(connection.destinationName, connection.destinationKind),
			true,
			map[string]string{"style": "dashed"},
		)

	}
}

// Populate populates the graph.
func (g *Grapher) Populate(pomlList *metav1.PartialObjectMetadataList, resource config.Resource) {
	// Add a node for each object in the List to the subgraph corresponding to the resource's rank.
	for _, poml := range pomlList.Items {
		g.graph.AddNode(getSubgraphName(resource.Rank), getSanitizedObjectName(poml.Name, poml.Kind), map[string]string{
			"penwidth": "0",
			"label":    getNodeLabel(poml.Name),
			"image":    getImagePath(resource.Resource),
		})
		// If the object contains a controlling owner reference, track it.
		// We do this so an edge can be constructed to link the object node to the owner node.
		// Ideally, we would skip the tracking and just create the edge now. But the owner node may not exist at
		// this point.
		if len(poml.OwnerReferences) > 0 && poml.OwnerReferences[0].Controller != nil && *poml.OwnerReferences[0].Controller {
			g.connections = append(g.connections, connection{
				sourceName:      poml.OwnerReferences[0].Name,
				sourceKind:      poml.OwnerReferences[0].Kind,
				destinationName: poml.Name,
				destinationKind: poml.Kind,
			})
		}

	}
}

// WriteDotFile writes the string representation of the graph to file.
func (g *Grapher) WriteDotFile() error {
	file, err := os.Create(g.outputFilePath)
	if err != nil {
		return fmt.Errorf("failed to create dot file: %v", err)
	}
	defer file.Close()

	_, err = file.WriteString(g.graph.String())
	if err != nil {
		return fmt.Errorf("failed to write to dot file: %v", err)
	}
	return nil
}
