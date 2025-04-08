package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/weslien/grit/pkg/grit"
	"gopkg.in/yaml.v3"
)

var noCache bool

var buildCmd = &cobra.Command{
	Use:   "build [type] [name]",
	Short: "Build packages and their dependencies",
	Long:  `Build packages respecting dependency order and utilizing build cache`,
	Run: func(cmd *cobra.Command, args []string) {
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Printf("Error getting current directory: %v\n", err)
			os.Exit(1)
		}

		pm := grit.NewPackageManager(cwd)
		packages, err := pm.LoadPackages()
		if err != nil {
			fmt.Printf("Error loading packages: %v\n", err)
			os.Exit(1)
		}

		buildOrder, err := resolveDependencies(packages)
		if err != nil {
			fmt.Printf("Error resolving dependencies: %v\n", err)
			os.Exit(1)
		}

		cacheDir := filepath.Join(cwd, ".grit", "cache")
		if !noCache {
			os.MkdirAll(cacheDir, 0755)
		}

		for _, cfg := range buildOrder {
			err := executeBuild(cfg, cacheDir, noCache)
			if err != nil {
				fmt.Printf("Error building %s: %v\n", cfg.Package.Name, err)
				os.Exit(1)
			}
		}
	},
}

func init() {
	buildCmd.Flags().BoolVar(&noCache, "no-cache", false, "Bypass build cache")
	rootCmd.AddCommand(buildCmd)
}

func resolveDependencies(packages []grit.Config) ([]grit.Config, error) {
	// Build dependency graph
	graph := make(map[string][]string)
	nodeMap := make(map[string]grit.Config)
	inDegree := make(map[string]int)

	// Initialize the graph with all packages
	for _, cfg := range packages {
		//fmt.Printf("Initializing node: %s\n", cfg.Package.Name)
		nodeMap[cfg.Package.Name] = cfg
		if _, exists := graph[cfg.Package.Name]; !exists {
			graph[cfg.Package.Name] = []string{}
		}
	}

	// Add dependencies to the graph
	for _, cfg := range packages {
		for _, dep := range cfg.Package.Dependencies {
			// Check if the dependency exists
			if _, exists := nodeMap[dep.Name]; !exists {
				// Skip missing dependencies or handle them differently
				fmt.Printf("Warning: Package %s depends on %s, but it doesn't exist\n",
					cfg.Package.Name, dep.Name)
				continue
			}

			graph[cfg.Package.Name] = append(graph[cfg.Package.Name], dep.Name)
			inDegree[dep.Name]++
		}
	}

	// Kahn's algorithm for topological sort
	var queue []string
	for name := range graph {
		if inDegree[name] == 0 {
			queue = append(queue, name)
		}
	}

	var order []grit.Config
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		order = append(order, nodeMap[node])

		for _, neighbor := range graph[node] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	// If we couldn't resolve all packages, there might be a cycle
	// But we'll return what we have instead of failing
	if len(order) != len(packages) {
		fmt.Println("Warning: Possible dependency cycle detected. Building packages in best-effort order.")

		// Add remaining packages in any order
		for name, cfg := range nodeMap {
			found := false
			for _, orderedCfg := range order {
				if orderedCfg.Package.Name == name {
					found = true
					break
				}
			}
			if !found {
				order = append(order, cfg)
			}
		}
	}

	return order, nil
}

func executeBuild(cfg grit.Config, cacheDir string, noCache bool) error {
	cacheFile := filepath.Join(cacheDir, cfg.Package.Name+".hash")

	if !noCache {
		if cachedHash, err := os.ReadFile(cacheFile); err == nil {
			if string(cachedHash) == cfg.Package.Hash {
				fmt.Printf("Using cached build for %s\n", cfg.Package.Name)
				return nil
			}
		}
	}

	fmt.Printf("Building package: %s\n", cfg.Package.Name)

	// Get the package directory from the stored path
	cfgDir := filepath.Dir(cfg.Package.Path)

	// Load the root config to get type information
	rootConfigPath := filepath.Join(filepath.Dir(cfgDir), "..", "..", "grit.yaml")
	rootConfigData, err := os.ReadFile(rootConfigPath)
	if err != nil {
		return fmt.Errorf("failed to read root config: %w", err)
	}

	var rootConfig grit.RootConfig
	if err := yaml.Unmarshal(rootConfigData, &rootConfig); err != nil {
		return fmt.Errorf("invalid root config: %w", err)
	}

	// Determine the package type from its path
	var cfgType string
	for typeName, typeConfig := range rootConfig.Types {
		if strings.Contains(cfgDir, typeConfig.PackageDir) {
			cfgType = typeName
			break
		}
	}

	if cfgType == "" {
		return fmt.Errorf("could not determine package type for %s", cfg.Package.Name)
	}

	// Load the package config
	var cfgConfig grit.Config
	cfgConfigData, err := os.ReadFile(cfg.Package.Path)
	if err != nil {
		return fmt.Errorf("failed to read package config: %w", err)
	}

	if err := yaml.Unmarshal(cfgConfigData, &cfgConfig); err != nil {
		return fmt.Errorf("invalid package config: %w", err)
	}

	// Get the build command - first try package config, then fall back to type config
	buildCmd, ok := cfgConfig.Targets["build"]
	if !ok || buildCmd == "" {
		// Fall back to type config
		typeConfig := rootConfig.Types[cfgType]
		buildCmd, ok = typeConfig.Targets["build"]
		if !ok || buildCmd == "" {
			return fmt.Errorf("no build command defined for package %s or type %s", cfg.Package.Name, cfgType)
		}
	}

	fmt.Printf("Executing build command for %s: %s\n", cfg.Package.Name, buildCmd)

	// Execute the build command
	cmd := exec.Command("sh", "-c", buildCmd)
	cmd.Dir = cfgDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("build command failed: %w", err)
	}

	// TODO: Calculate real package hash from source files
	if !noCache {
		os.WriteFile(cacheFile, []byte(cfg.Package.Hash), 0644)
	}

	return nil
}
