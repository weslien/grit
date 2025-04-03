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
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Implement package creation logic
		fmt.Printf("Creating %s package: %s\n", args[0], args[1])
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
			DefaultTasks: map[string]string{
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

func loadRootConfig() (*grit.Config, error) {
	data, err := os.ReadFile("grit.yaml")
	if err != nil {
		return &grit.Config{Types: make(map[string]grit.TypeConfig)}, nil
	}
	var config grit.Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	if config.Types == nil {
		config.Types = make(map[string]grit.TypeConfig)
	}
	return &config, nil
}

func saveRootConfig(config *grit.Config) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	return os.WriteFile("grit.yaml", data, 0644)
}
