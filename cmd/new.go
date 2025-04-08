package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/weslien/grit/pkg/grit"
	"gopkg.in/yaml.v3"
)

var newCmd = &cobra.Command{
	Use:   "new [type] [name]",
	Short: "Create a new package",
	Long:  "Initialize a new package template with specified type and name",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		typeName := args[0]
		pkgName := args[1]

		// Load root config
		config, err := loadRootConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Check if type exists
		typeConfig, exists := config.Types[typeName]
		if !exists {
			return fmt.Errorf("type '%s' does not exist", typeName)
		}

		// Create package directory
		pkgDir := filepath.Join(typeConfig.PackageDir, pkgName)
		if err := os.MkdirAll(pkgDir, 0755); err != nil {
			return fmt.Errorf("failed to create package directory: %w", err)
		}

		// Create standard package subdirectories
		subdirs := []string{
			filepath.Join(pkgDir, "src"),
			filepath.Join(pkgDir, ".prompt"),
			filepath.Join(pkgDir, ".mod"),
			filepath.Join(pkgDir, ".dev"),
			filepath.Join(pkgDir, ".ops"),
		}

		for _, dir := range subdirs {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", dir, err)
			}
		}

		// Create package config file
		pkgConfig := &grit.Config{
			Package: grit.Package{
				Name:    pkgName,
				Version: "0.1.0",
			},
			Targets: make(map[string]string),
		}

		// Copy targets from type config
		if typeConfig.Targets != nil {
			for k, v := range typeConfig.Targets {
				pkgConfig.Targets[k] = v
			}
		}

		// Save package config
		pkgConfigData, err := yaml.Marshal(pkgConfig)
		if err != nil {
			return fmt.Errorf("failed to marshal package config: %w", err)
		}

		if err := os.WriteFile(filepath.Join(pkgDir, "grit.yaml"), pkgConfigData, 0644); err != nil {
			return fmt.Errorf("failed to write package config: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Creating %s package: %s\n", typeName, pkgName)
		fmt.Fprintf(cmd.OutOrStdout(), "Package created at: %s\n", pkgDir)
		return nil
	},
}

var newTypeCmd = &cobra.Command{
	Use:   "type [name]",
	Short: "Create a new package type",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		typeName := args[0]

		// Update root grit.yaml
		config, err := loadRootConfig()
		if err != nil {
			log.Fatal(err)
		}

		// Add new type configuration
		config.Types[typeName] = grit.TypeConfig{
			PackageDir:  filepath.Join("packages", typeName),
			BuildDir:    filepath.Join("build", typeName),
			CoverageDir: filepath.Join("coverage", typeName),
			Targets: map[string]string{
				"build": "echo 'Implement build logic'",
				"test":  "echo 'Implement test logic'",
			},
		}

		// Write updated config
		if err := saveRootConfig(config); err != nil {
			log.Fatal(err)
		}

		// Create package directories
		dirs := []string{
			filepath.Join("packages", typeName),
			filepath.Join(".prompt", typeName),
			filepath.Join(".mod", typeName),
			filepath.Join(".dev", typeName),
			filepath.Join(".ops", typeName),
		}

		for _, dir := range dirs {
			if err := os.MkdirAll(dir, 0755); err != nil {
				log.Fatal(err)
			}
		}

		fmt.Printf("Created new type '%s'\n", typeName)
	},
}

func init() {
	newCmd.AddCommand(newTypeCmd)
	rootCmd.AddCommand(newCmd)
}

func loadRootConfig() (*grit.RootConfig, error) {
	data, err := os.ReadFile("grit.yaml")
	if err != nil {
		return &grit.RootConfig{Types: make(map[string]grit.TypeConfig)}, nil
	}
	var config grit.RootConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	if config.Types == nil {
		config.Types = make(map[string]grit.TypeConfig)
	}
	return &config, nil
}

func saveRootConfig(config *grit.RootConfig) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	return os.WriteFile("grit.yaml", data, 0644)
}
