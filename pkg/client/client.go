package client

import (
	"fmt"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Client interacts with resources on a Kubernetes cluster.
type Client struct {
	client *dynamic.DynamicClient
}

// NewClient returns a new *Client.
// An in-cluster REST configuration is fetched. If this fails, a local one is used in its place.
func NewClient() (*Client, error) {
	restConfiguration, err := rest.InClusterConfig()
	if err != nil {
		kubeconfig := clientcmd.NewDefaultClientConfigLoadingRules().GetDefaultFilename()
		restConfiguration, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("failed to get REST configuration: %v", err)
		}
	}

	dynamicClient, err := dynamic.NewForConfig(restConfiguration)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %v", err)
	}

	return &Client{client: dynamicClient}, nil
}
