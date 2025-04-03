package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/weslien/grit/pkg/grit"
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

		for _, pkg := range buildOrder {
			err := executeBuild(pkg, cacheDir, noCache)
			if err != nil {
				fmt.Printf("Error building %s: %v\n", pkg.Name, err)
				os.Exit(1)
			}
		}
	},
}

func init() {
	buildCmd.Flags().BoolVar(&noCache, "no-cache", false, "Bypass build cache")
	rootCmd.AddCommand(buildCmd)
}

func resolveDependencies(packages []grit.Package) ([]grit.Package, error) {
	// Build dependency graph
	graph := make(map[string][]string)
	nodeMap := make(map[string]grit.Package)
	inDegree := make(map[string]int)

	for _, pkg := range packages {
		nodeMap[pkg.Name] = pkg
		graph[pkg.Name] = []string{}
		for _, dep := range pkg.Dependencies {
			graph[pkg.Name] = append(graph[pkg.Name], dep.Name)
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

	var order []grit.Package
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

	if len(order) != len(packages) {
		return nil, fmt.Errorf("dependency cycle detected")
	}

	return order, nil
}

func executeBuild(pkg grit.Package, cacheDir string, noCache bool) error {
	cacheFile := filepath.Join(cacheDir, pkg.Name+".hash")

	if !noCache {
		if cachedHash, err := os.ReadFile(cacheFile); err == nil {
			if string(cachedHash) == pkg.Hash {
				fmt.Printf("Using cached build for %s\n", pkg.Name)
				return nil
			}
		}
	}

	fmt.Printf("Building package: %s\n", pkg.Name)
	// TODO: Actual build implementation
	// TODO: Calculate real package hash from source files

	if !noCache {
		os.WriteFile(cacheFile, []byte(pkg.Hash), 0644)
	}

	return nil
}
