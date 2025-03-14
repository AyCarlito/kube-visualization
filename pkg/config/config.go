package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Config represents a configuration file for the worker.
type Config struct {
	Namespace string                    `json:"namespace"`
	Resources []schema.GroupVersionKind `json:"resources"`
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
