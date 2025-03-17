package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Resource represents a ranked GVR.
type Resource struct {
	schema.GroupVersionResource
	// Rank identifies where the GVR should be ranked in the heirarchical visualization.
	Rank *int `json:"rank"`
}

// UniqueRanks returns the number of unique ranks in a slice of Resource.
func UniqueRanks(resources []Resource) int {
	ranks := make(map[int]struct{})
	for _, resource := range resources {
		if resource.Rank != nil {
			ranks[*resource.Rank] = struct{}{}
		}
	}
	// Resources are to be placed in the order they are retrieved.
	if len(resources) > 0 && len(ranks) == 0 {
		return len(resources)
	}
	return len(ranks)
}

// Config represents a configuration file for the worker.
type Config struct {
	Resources []Resource `json:"resources"`
}

// NewConfig reads a configuration file from path and returns a new Config.
func NewConfig(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open configuration file: %v", err)
	}
	defer f.Close()

	fileBytes, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("failed to read configuration file: %v", err)
	}

	config := &Config{}
	err = json.Unmarshal(fileBytes, config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal configuration file: %v", err)
	}
	return config, nil
}
