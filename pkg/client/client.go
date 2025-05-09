package client

import (
	"context"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// requestTimeout defines the timeout before a context is cancelled when performing Kubernetes API operations.
const requestTimeout = 5 * time.Second

// OptFunc is a function that mutates a clientOpts
type OptFunc func(*clientOpts)

// clientOpts are the configuration options for the Client.
type clientOpts struct {
	labelSelector  string
	kubeConfigPath string
}

// defaultOpts return the default configuration options for a Client
func defaultOpts() clientOpts {
	return clientOpts{
		labelSelector:  "",
		kubeConfigPath: "",
	}
}

// WithLabelSelector returns an optFunc to mutate the labelSelector configuration option of the Client.
func WithLabelSelector(ls string) OptFunc {
	return func(o *clientOpts) {
		o.labelSelector = ls
	}
}

// WithKubeConfigPath returns an optFunc to mutate the kubeConfigPath configuration option of the Client.
func WithKubeConfigPath(p string) OptFunc {
	return func(o *clientOpts) {
		o.kubeConfigPath = p
	}
}

// Client interacts with resources on a Kubernetes cluster.
type Client struct {
	client *dynamic.DynamicClient
	opts   clientOpts
}

// NewClient returns a new *Client.
// An in-cluster REST configuration is fetched. If this fails, a local one is used in its place.
func NewClient(opts ...OptFunc) (*Client, error) {
	o := defaultOpts()
	for _, fn := range opts {
		fn(&o)
	}

	restConfiguration, err := rest.InClusterConfig()
	if err != nil {
		kubeConfigLoadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
		kubeConfigLoadingRules.ExplicitPath = o.kubeConfigPath
		restConfiguration, err = clientcmd.BuildConfigFromFlags("", kubeConfigLoadingRules.GetDefaultFilename())
		if err != nil {
			return nil, fmt.Errorf("failed to get REST configuration: %v", err)
		}
	}

	dynamicClient, err := dynamic.NewForConfig(restConfiguration)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %v", err)
	}

	return &Client{client: dynamicClient, opts: o}, nil
}

// List returns a list of objects in a namespace for a given GVK.
// The full object definition is not returned, only the metadata.
func (c *Client) List(ctx context.Context, gvr schema.GroupVersionResource, namespace string) (*unstructured.UnstructuredList, error) {
	// Timebox the API call.
	timeoutCtx, cxl := context.WithTimeout(ctx, requestTimeout)
	defer cxl()

	// List the objects.
	unstructuredList, err := c.client.Resource(gvr).Namespace(namespace).List(timeoutCtx, metav1.ListOptions{LabelSelector: c.opts.labelSelector})
	if err != nil {
		return nil, fmt.Errorf("failed to list objects: %v", err)
	}

	return unstructuredList, nil
}
