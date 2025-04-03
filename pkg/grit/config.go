package grit

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config operations
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return &Config{Types: make(map[string]TypeConfig)}, nil
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if config.Types == nil {
		config.Types = make(map[string]TypeConfig)
	}
	return &config, nil
}

func SaveConfig(config *Config, path string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

func (c *Config) MergeDefaults(defaults TypeConfig) {
	if _, exists := c.Types["lib"]; !exists {
		c.Types["lib"] = TypeConfig{
			PackageDir:   "packages/lib",
			BuildDir:     "build/lib",
			CoverageDir:  "coverage/lib",
			DefaultTasks: defaults.DefaultTasks,
		}
	}
}
