package cmd

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/weslien/grit/pkg/grit"
	"gopkg.in/yaml.v3"
)

//go:embed tpl/grit.yaml
var gritYamlTemplate string

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new Grit workspace",
	Long:  `Create a new Grit workspace with default configuration files`,
	RunE: func(cmd *cobra.Command, args []string) error {
		rootDir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}

		gritDir := filepath.Join(rootDir, ".grit")

		if err := os.MkdirAll(gritDir, 0755); err != nil {
			return fmt.Errorf("failed to create .grit directory: %w", err)
		}

		configFile := filepath.Join("grit.yaml")
		var existingConfig grit.Config

		if _, err := os.Stat(configFile); err == nil {
			data, err := os.ReadFile(configFile)
			if err != nil {
				return fmt.Errorf("failed to read existing grit.yaml: %w", err)
			}
			if err := yaml.Unmarshal(data, &existingConfig); err != nil {
				return fmt.Errorf("failed to parse existing grit.yaml: %w", err)
			}
		}

		templateConfig := grit.TypeConfig{}
		yaml.Unmarshal([]byte(gritYamlTemplate), &templateConfig)

		mergedConfig := existingConfig
		if existingConfig.Types == nil {
			existingConfig.Types = make(map[string]grit.TypeConfig)
		}

		// Merge type configuration
		if _, exists := existingConfig.Types["lib"]; !exists {
			existingConfig.Types["lib"] = grit.TypeConfig{
				PackageDir:  "packages/lib",
				BuildDir:    "build/lib",
				CoverageDir: "coverage/lib",
				Targets:     templateConfig.Targets,
			}
		}

		mergedData, err := yaml.Marshal(mergedConfig)
		if err != nil {
			return fmt.Errorf("failed to marshal merged config: %w", err)
		}

		if err := os.WriteFile(configFile, mergedData, 0644); err != nil {
			return fmt.Errorf("failed to update grit.yaml: %w", err)
		}

		fmt.Printf("Initialized Grit workspace in %s\n", rootDir)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
