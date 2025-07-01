package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/weslien/grit/pkg/grit"
	"github.com/weslien/grit/pkg/output"
	"gopkg.in/yaml.v3"
)

var (
	outputFormat string
	outputFile   string
	showTypes    bool
)

var graphCmd = &cobra.Command{
	Use:   "graph",
	Short: "Visualize package dependencies",
	Long: `Generate and display dependency graphs in various formats.
	
Supports text tree format for quick viewing and DOT format for use with Graphviz.

Examples:
  grit graph                    # Show dependency tree in terminal
  grit graph --format dot       # Output DOT format for Graphviz
  grit graph --output deps.dot  # Save DOT format to file
  grit graph --types            # Include package types in output`,
	Run: func(cmd *cobra.Command, args []string) {
		formatter := output.New()

		cwd, err := os.Getwd()
		if err != nil {
			formatter.Error(fmt.Sprintf("Error getting current directory: %v", err))
			os.Exit(1)
		}

		formatter.Header("Dependency Graph")
		formatter.Section("Loading Packages")

		pm := grit.NewPackageManager(cwd)
		packages, err := pm.LoadPackages()
		if err != nil {
			formatter.Error(fmt.Sprintf("Error loading packages: %v", err))
			os.Exit(1)
		}
		formatter.Success(fmt.Sprintf("Loaded %d packages", len(packages)))

		// Build dependency map
		depMap := make(map[string][]string)
		packageTypes := make(map[string]string)
		packageVersions := make(map[string]string)

		// Load root config for type information
		rootConfig, err := loadRootConfigForGraph(cwd)
		if err != nil {
			formatter.Warning("Could not load root config, package types will not be shown")
		}

		for _, cfg := range packages {
			if cfg.Package.Name == "" {
				continue // Skip root config
			}

			depMap[cfg.Package.Name] = cfg.Package.Dependencies
			packageVersions[cfg.Package.Name] = cfg.Package.Version

			// Determine package type
			if rootConfig != nil {
				pkgType := getPackageType(cfg.Package.Path, rootConfig, cwd)
				packageTypes[cfg.Package.Name] = pkgType
			}
		}

		if len(depMap) == 0 {
			formatter.Info("No packages found")
			return
		}

		switch outputFormat {
		case "dot":
			err := generateDotGraph(depMap, packageTypes, packageVersions, outputFile, formatter)
			if err != nil {
				formatter.Error(fmt.Sprintf("Error generating DOT graph: %v", err))
				os.Exit(1)
			}
		case "tree", "":
			generateTreeGraph(depMap, packageTypes, packageVersions, formatter)
		default:
			formatter.Error(fmt.Sprintf("Unknown output format: %s", outputFormat))
			os.Exit(1)
		}
	},
}

func init() {
	graphCmd.Flags().StringVarP(&outputFormat, "format", "f", "tree", "Output format (tree, dot)")
	graphCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file (default: stdout)")
	graphCmd.Flags().BoolVar(&showTypes, "types", false, "Show package types in output")
	rootCmd.AddCommand(graphCmd)
}

func generateTreeGraph(depMap map[string][]string, packageTypes map[string]string, packageVersions map[string]string, formatter *output.Formatter) {
	formatter.Section("Package Dependencies")

	// Sort packages for consistent output
	var packages []string
	for pkg := range depMap {
		packages = append(packages, pkg)
	}
	sort.Strings(packages)

	// Find root packages (packages with no dependents)
	dependents := make(map[string][]string)
	for pkg, deps := range depMap {
		for _, dep := range deps {
			dependents[dep] = append(dependents[dep], pkg)
		}
	}

	var roots []string
	for _, pkg := range packages {
		if len(dependents[pkg]) == 0 {
			roots = append(roots, pkg)
		}
	}

	if len(roots) == 0 {
		// No clear roots, show all packages
		roots = packages
	}

	// Display tree for each root
	for i, root := range roots {
		if i > 0 {
			formatter.NewLine()
		}
		displayPackageTree(root, depMap, packageTypes, packageVersions, formatter, "", make(map[string]bool))
	}

	// Show statistics
	formatter.NewLine()
	formatter.Section("Statistics")
	
	totalDeps := 0
	for _, deps := range depMap {
		totalDeps += len(deps)
	}
	
	formatter.Detail(fmt.Sprintf("Total packages: %d", len(packages)))
	formatter.Detail(fmt.Sprintf("Total dependencies: %d", totalDeps))
	if len(packages) > 0 {
		formatter.Detail(fmt.Sprintf("Average dependencies per package: %.1f", float64(totalDeps)/float64(len(packages))))
	}
	
	// Find packages with most dependencies
	type packageDepCount struct {
		name  string
		count int
	}
	
	var depCounts []packageDepCount
	for pkg, deps := range depMap {
		if len(deps) > 0 {
			depCounts = append(depCounts, packageDepCount{pkg, len(deps)})
		}
	}
	
	sort.Slice(depCounts, func(i, j int) bool {
		return depCounts[i].count > depCounts[j].count
	})
	
	if len(depCounts) > 0 {
		formatter.Detail("Packages with most dependencies:")
		for i, dc := range depCounts {
			if i >= 3 { // Show top 3
				break
			}
			formatter.Detail(fmt.Sprintf("  • %s (%d dependencies)", dc.name, dc.count))
		}
	}
}

func displayPackageTree(pkg string, depMap map[string][]string, packageTypes map[string]string, packageVersions map[string]string, formatter *output.Formatter, prefix string, visited map[string]bool) {
	if visited[pkg] {
		fmt.Printf("%s├─ %s (circular reference)\n", prefix, pkg)
		return
	}

	visited[pkg] = true
	defer func() { visited[pkg] = false }()

	// Format package name with optional type and version
	pkgDisplay := pkg
	if showTypes {
		if pkgType, ok := packageTypes[pkg]; ok && pkgType != "" {
			pkgDisplay += fmt.Sprintf(" (%s)", pkgType)
		}
		if version, ok := packageVersions[pkg]; ok && version != "" {
			pkgDisplay += fmt.Sprintf(" v%s", version)
		}
	}

	fmt.Printf("%s├─ %s\n", prefix, pkgDisplay)

	deps := depMap[pkg]
	for i, dep := range deps {
		isLast := i == len(deps)-1
		var newPrefix string
		if isLast {
			newPrefix = prefix + "   "
		} else {
			newPrefix = prefix + "│  "
		}
		
		displayPackageTree(dep, depMap, packageTypes, packageVersions, formatter, newPrefix, visited)
	}
}

func generateDotGraph(depMap map[string][]string, packageTypes map[string]string, packageVersions map[string]string, outputFile string, formatter *output.Formatter) error {
	var output strings.Builder
	
	output.WriteString("digraph dependencies {\n")
	output.WriteString("  rankdir=TB;\n")
	output.WriteString("  node [shape=box, style=rounded];\n")
	output.WriteString("  edge [color=gray];\n\n")

	// Add nodes with styling based on package type
	typeColors := map[string]string{
		"app":     "lightblue",
		"lib":     "lightgreen", 
		"service": "lightyellow",
		"tool":    "lightcoral",
	}

	for pkg := range depMap {
		label := pkg
		if showTypes {
			if version, ok := packageVersions[pkg]; ok && version != "" {
				label += "\\nv" + version
			}
			if pkgType, ok := packageTypes[pkg]; ok && pkgType != "" {
				label += "\\n(" + pkgType + ")"
			}
		}

		color := "lightgray"
		if pkgType, ok := packageTypes[pkg]; ok {
			if typeColor, exists := typeColors[pkgType]; exists {
				color = typeColor
			}
		}

		output.WriteString(fmt.Sprintf("  \"%s\" [label=\"%s\", fillcolor=\"%s\", style=\"filled,rounded\"];\n", 
			pkg, label, color))
	}

	output.WriteString("\n")

	// Add edges
	for pkg, deps := range depMap {
		for _, dep := range deps {
			output.WriteString(fmt.Sprintf("  \"%s\" -> \"%s\";\n", pkg, dep))
		}
	}

	output.WriteString("}\n")

	// Output to file or stdout
	if outputFile != "" {
		err := os.WriteFile(outputFile, []byte(output.String()), 0644)
		if err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		formatter.Success(fmt.Sprintf("DOT graph written to %s", outputFile))
		formatter.Detail("To generate an image, run: dot -Tpng " + outputFile + " -o deps.png")
		formatter.Detail("Or view interactively with: xdot " + outputFile)
	} else {
		fmt.Print(output.String())
	}

	return nil
}

func loadRootConfigForGraph(cwd string) (*grit.RootConfig, error) {
	rootConfigPath := filepath.Join(cwd, "grit.yaml")
	data, err := os.ReadFile(rootConfigPath)
	if err != nil {
		return nil, err
	}

	var rootConfig grit.RootConfig
	err = yaml.Unmarshal(data, &rootConfig)
	if err != nil {
		return nil, err
	}

	return &rootConfig, nil
}

func getPackageType(packagePath string, rootConfig *grit.RootConfig, cwd string) string {
	relPath, err := filepath.Rel(cwd, filepath.Dir(packagePath))
	if err != nil {
		return ""
	}

	for typeName, typeConfig := range rootConfig.Types {
		if strings.Contains(relPath, typeConfig.PackageDir) {
			return typeName
		}
	}

	return ""
}