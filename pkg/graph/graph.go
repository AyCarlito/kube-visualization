package graph

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/awalterschulze/gographviz"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/AyCarlito/kube-visualization/pkg/config"
)

const (
	ConfigMap             string = "ConfigMap"
	Endpoints             string = "Endpoints"
	Ingress               string = "Ingress"
	PersistentVolumeClaim string = "PersistentVolumeClaim"
	Pod                   string = "Pod"
	Secret                string = "Secret"
	Service               string = "Service"
)

// Grapher creates gographviz graphs.
type Grapher struct {
	connections    []connection
	graph          *gographviz.Graph
	assetsBasePath string
	outputFilePath string
}

// NewGrapher returns a new *Grapher.
func NewGraph(a, o string) *Grapher {
	return &Grapher{assetsBasePath: a, outputFilePath: o}
}

// connection is a link between two Kubernetes objects.
type connection struct {
	label           string
	sourceName      string
	sourceKind      string
	destinationName string
	destinationKind string
}

// sanitizedLabel returns the sanitized label of a connection.
// The label is wrapped in double quotes.
func (c *connection) sanitizedLabel() string {
	return fmt.Sprintf("\"%s\"", c.label)
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
func (g *Grapher) getImagePath(resource string) string {
	return fmt.Sprintf("\"%s\"", filepath.Join(g.assetsBasePath, resource+".png"))
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
		"image":    g.getImagePath("namespaces"),
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
		sourceNodeName := getSanitizedObjectName(connection.sourceName, connection.sourceKind)
		// It's possible that the connection may be towards a resource that isn't part of this visualisation.
		// Check for the existence of the source node first, and skip if it does not exist.
		if _, ok := g.graph.Nodes.Lookup[sourceNodeName]; !ok {
			continue
		}
		dstNodeName := getSanitizedObjectName(connection.destinationName, connection.destinationKind)
		// There may already be a connection between the source and destination node.
		// We only want to represent one.
		if _, ok := g.graph.Edges.SrcToDsts[sourceNodeName][dstNodeName]; ok {
			continue
		}
		attrs := map[string]string{"style": "dashed"}
		if connection.label != "" {
			attrs["label"] = connection.sanitizedLabel()
		}
		g.graph.AddEdge(sourceNodeName, dstNodeName, true, attrs)

	}
}

// Populate populates the graph.
func (g *Grapher) Populate(objects *unstructured.UnstructuredList, resource config.Resource) {
	// Add a node for each object in the List to the subgraph corresponding to the resource's rank.
	for _, object := range objects.Items {
		name := object.GetName()
		kind := object.GetKind()
		g.graph.AddNode(getSubgraphName(resource.Rank), getSanitizedObjectName(name, kind), map[string]string{
			"penwidth": "0",
			"label":    getNodeLabel(name),
			"image":    g.getImagePath(resource.Resource),
		})
		// If the object contains a controlling owner reference, track it.
		// We do this so an edge can be constructed to link the object node to the owner node.
		// Ideally, we would skip the tracking and just create the edge now. But the owner node may not exist at
		// this point.
		ownerReferences := object.GetOwnerReferences()
		if len(ownerReferences) > 0 && ownerReferences[0].Controller != nil && *ownerReferences[0].Controller {
			g.connections = append(g.connections, connection{
				sourceName:      ownerReferences[0].Name,
				sourceKind:      ownerReferences[0].Kind,
				destinationName: name,
				destinationKind: kind,
			})
		}

		// Services are connected to Endpoints by name.
		if kind == Service {
			service := &corev1.Service{}
			runtime.DefaultUnstructuredConverter.FromUnstructured(object.UnstructuredContent(), service)
			// Consider two ports:
			//     - &ServicePort{Name:api,Protocol:TCP,Port:8080,TargetPort:{1 0 api},NodePort:0,AppProtocol:nil,}
			//     - &ServicePort{Name:metrics,Protocol:TCP,Port:3001,TargetPort:{0 3001 },NodePort:0,AppProtocol:nil,}
			// Generate a label for the connection:
			//     8080/TCP/api\n3001/TCP/metrics
			var connectionLabel string
			for _, port := range service.Spec.Ports {
				connectionLabel += fmt.Sprintf("%d/%s/%s\\n", port.Port, port.Protocol, port.Name)
			}
			g.connections = append(g.connections, connection{
				label:           connectionLabel,
				sourceName:      name,
				sourceKind:      kind,
				destinationName: name,
				destinationKind: Endpoints,
			})
		}

		// Endpoints are connected to the pods referenced in its subsets.
		if kind == Endpoints {
			endpoints := &corev1.Endpoints{}
			runtime.DefaultUnstructuredConverter.FromUnstructured(object.UnstructuredContent(), endpoints)
			for _, subset := range endpoints.Subsets {
				for _, address := range subset.Addresses {
					if address.TargetRef == nil || address.TargetRef.Kind != Pod {
						continue
					}
					// Consider two ports:
					//     - &EndpointPort{Name:api,Port:8080,Protocol:TCP,AppProtocol:nil,}
					//     - &EndpointPort{Name:metrics,Port:3001,Protocol:TCP,AppProtocol:nil,}
					// Generate a label for the connection:
					//     8080/TCP/api\n3001/TCP/metrics
					var connectionLabel string
					for _, port := range subset.Ports {
						connectionLabel += fmt.Sprintf("%d/%s/%s\\n", port.Port, port.Protocol, port.Name)
					}
					g.connections = append(g.connections, connection{
						label:           connectionLabel,
						sourceName:      name,
						sourceKind:      kind,
						destinationName: address.TargetRef.Name,
						destinationKind: Pod,
					})
				}
			}
		}

		// Ingresses are connected to the service defined in the IngressBackend.
		if kind == Ingress {
			ingress := &networkingv1.Ingress{}
			runtime.DefaultUnstructuredConverter.FromUnstructured(object.UnstructuredContent(), ingress)
			for _, rule := range ingress.Spec.Rules {
				if rule.IngressRuleValue.HTTP == nil {
					continue
				}
				for _, path := range rule.IngressRuleValue.HTTP.Paths {
					var serviceName string
					if path.Backend.Resource != nil && path.Backend.Resource.Kind == Service {
						serviceName = path.Backend.Resource.Name
					} else if path.Backend.Service != nil {
						serviceName = path.Backend.Service.Name
					} else {
						continue
					}
					g.connections = append(g.connections, connection{
						label:           path.Path,
						sourceName:      name,
						sourceKind:      kind,
						destinationName: serviceName,
						destinationKind: Service,
					})
				}
			}
		}

		// Pods are connected to ConfigMaps, Secrets and PersistentVolumeClaims through their volumes.
		if kind == Pod {
			pod := &corev1.Pod{}
			runtime.DefaultUnstructuredConverter.FromUnstructured(object.UnstructuredContent(), pod)
			for _, volume := range pod.Spec.Volumes {
				var volumeName, volumeKind string
				if volume.ConfigMap != nil {
					volumeName = volume.ConfigMap.Name
					volumeKind = ConfigMap
				} else if volume.Secret != nil {
					volumeName = volume.Secret.SecretName
					volumeKind = Secret
				} else if volume.PersistentVolumeClaim != nil {
					volumeName = volume.PersistentVolumeClaim.ClaimName
					volumeKind = PersistentVolumeClaim
				} else {
					continue
				}
				g.connections = append(g.connections, connection{
					sourceName:      volumeName,
					sourceKind:      volumeKind,
					destinationName: name,
					destinationKind: kind,
				})
			}
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
