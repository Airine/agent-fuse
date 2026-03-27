package manifest

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Manifest represents the parsed contents of an agentfs.yaml file.
type Manifest struct {
	Version     string `yaml:"version"`
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

// Parse reads agentfs.yaml from the given path and returns a Manifest.
// Returns error if: file not found, invalid YAML, or required Name field is missing.
func Parse(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var m Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("invalid YAML: %w", err)
	}

	if m.Name == "" {
		return nil, fmt.Errorf("missing required field: name")
	}

	if m.Version == "" {
		m.Version = "1.0"
	}

	return &m, nil
}
