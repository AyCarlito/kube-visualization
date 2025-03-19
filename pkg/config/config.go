package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Resource represents a ranked GVR.
type Resource struct {
	schema.GroupVersionResource
	// Rank identifies where the GVR should be ranked in the heirarchical visualization.
	Rank int `json:"rank"`
}

// uniqueRanks returns the unique ranks of the Resources.
func uniqueRanks(resources []Resource) []int {
	uniqueRanks := []int{}
	existingRanks := make(map[int]struct{})
	for _, resource := range resources {
		existingRanks[resource.Rank] = struct{}{}
		uniqueRanks = append(uniqueRanks, resource.Rank)
	}
	return uniqueRanks
}

// SortedUniqueRanks returns the sorted unique ranks of the Resources.
func SortedUniqueRanks(resources []Resource) []int {
	u := uniqueRanks(resources)
	sort.Slice(u, func(i, j int) bool {
		return u[i] < u[j]
	})
	return u
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
