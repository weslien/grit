package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/weslien/grit/pkg/grit"
	"github.com/weslien/grit/pkg/output"
	"gopkg.in/yaml.v3"
)

var (
	verboseAnalysis bool
	jsonOutput      bool
)

type PackageAnalysis struct {
	Name         string            `json:"name"`
	Type         string            `json:"type"`
	Version      string            `json:"version"`
	Path         string            `json:"path"`
	Dependencies []string          `json:"dependencies"`
	Dependents   []string          `json:"dependents"`
	Issues       []string          `json:"issues"`
	Suggestions  []string          `json:"suggestions"`
	BuildTime    time.Duration     `json:"build_time,omitempty"`
	FileCount    int              `json:"file_count"`
	Size         int64            `json:"size_bytes"`
	LastModified time.Time        `json:"last_modified"`
}

type WorkspaceAnalysis struct {
	TotalPackages    int                        `json:"total_packages"`
	PackagesByType   map[string]int            `json:"packages_by_type"`
	TotalDependencies int                      `json:"total_dependencies"`
	CircularDeps     [][]string                `json:"circular_dependencies"`
	OrphanPackages   []string                  `json:"orphan_packages"`
	CriticalPath     []string                  `json:"critical_path"`
	Packages         map[string]PackageAnalysis `json:"packages"`
	Issues           []string                  `json:"workspace_issues"`
	Suggestions      []string                  `json:"workspace_suggestions"`
}

var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyze workspace health and dependencies",
	Long: `Perform comprehensive analysis of the workspace including:
- Package dependency analysis
- Circular dependency detection
- Orphaned package detection
- Build time analysis
- Package health checks
- Optimization suggestions

Examples:
  grit analyze                # Basic analysis
  grit analyze --verbose      # Detailed analysis with suggestions
  grit analyze --json         # Output analysis in JSON format`,
	Run: func(cmd *cobra.Command, args []string) {
		formatter := output.New()

		cwd, err := os.Getwd()
		if err != nil {
			formatter.Error(fmt.Sprintf("Error getting current directory: %v", err))
			os.Exit(1)
		}

		if !jsonOutput {
			formatter.Header("Workspace Analysis")
			formatter.Section("Loading Packages")
		}

		pm := grit.NewPackageManager(cwd)
		packages, err := pm.LoadPackages()
		if err != nil {
			formatter.Error(fmt.Sprintf("Error loading packages: %v", err))
			os.Exit(1)
		}

		if !jsonOutput {
			formatter.Success(fmt.Sprintf("Loaded %d packages", len(packages)))
		}

		// Perform analysis
		analysis := performWorkspaceAnalysis(packages, cwd, formatter)

		if jsonOutput {
			// Output JSON
			outputJSON(analysis)
		} else {
			// Output formatted analysis
			displayAnalysis(analysis, formatter)
		}
	},
}

func init() {
	analyzeCmd.Flags().BoolVarP(&verboseAnalysis, "verbose", "v", false, "Show detailed analysis and suggestions")
	analyzeCmd.Flags().BoolVar(&jsonOutput, "json", false, "Output analysis in JSON format")
	rootCmd.AddCommand(analyzeCmd)
}

func performWorkspaceAnalysis(packages []grit.Config, cwd string, formatter *output.Formatter) WorkspaceAnalysis {
	analysis := WorkspaceAnalysis{
		PackagesByType: make(map[string]int),
		Packages:       make(map[string]PackageAnalysis),
		Issues:         []string{},
		Suggestions:    []string{},
	}

	// Load root config
	rootConfig, err := loadRootConfigForAnalysis(cwd)
	if err != nil && !jsonOutput {
		formatter.Warning("Could not load root config")
	}

	// Build dependency maps
	depMap := make(map[string][]string)
	dependentMap := make(map[string][]string)

	// Analyze each package
	for _, cfg := range packages {
		if cfg.Package.Name == "" {
			continue // Skip root config
		}

		analysis.TotalPackages++
		depMap[cfg.Package.Name] = cfg.Package.Dependencies
		analysis.TotalDependencies += len(cfg.Package.Dependencies)

		// Build reverse dependency map
		for _, dep := range cfg.Package.Dependencies {
			dependentMap[dep] = append(dependentMap[dep], cfg.Package.Name)
		}

		// Analyze individual package
		pkgAnalysis := analyzePackage(cfg, rootConfig, cwd)
		analysis.Packages[cfg.Package.Name] = pkgAnalysis

		// Count by type
		if pkgAnalysis.Type != "" {
			analysis.PackagesByType[pkgAnalysis.Type]++
		}
	}

	// Detect circular dependencies
	analysis.CircularDeps = detectCircularDependencies(depMap)

	// Find orphaned packages (no dependents)
	for pkg := range depMap {
		if len(dependentMap[pkg]) == 0 {
			analysis.OrphanPackages = append(analysis.OrphanPackages, pkg)
		}
	}

	// Find critical path (longest dependency chain)
	analysis.CriticalPath = findCriticalPath(depMap)

	// Generate workspace-level suggestions
	analysis.Issues, analysis.Suggestions = generateWorkspaceSuggestions(analysis)

	return analysis
}

func analyzePackage(cfg grit.Config, rootConfig *grit.RootConfig, cwd string) PackageAnalysis {
	pkgAnalysis := PackageAnalysis{
		Name:         cfg.Package.Name,
		Version:      cfg.Package.Version,
		Path:         cfg.Package.Path,
		Dependencies: cfg.Package.Dependencies,
		Issues:       []string{},
		Suggestions:  []string{},
	}

	// Determine package type
	if rootConfig != nil {
		pkgAnalysis.Type = getPackageTypeForAnalysis(cfg.Package.Path, rootConfig, cwd)
	}

	// Analyze package directory
	pkgDir := filepath.Dir(cfg.Package.Path)
	if stat, err := os.Stat(pkgDir); err == nil {
		pkgAnalysis.LastModified = stat.ModTime()
	}

	// Count files and calculate size
	pkgAnalysis.FileCount, pkgAnalysis.Size = analyzePackageFiles(pkgDir)

	// Check for common issues
	pkgAnalysis.Issues, pkgAnalysis.Suggestions = analyzePackageHealth(cfg, pkgDir, rootConfig)

	return pkgAnalysis
}

func analyzePackageFiles(pkgDir string) (int, int64) {
	var fileCount int
	var totalSize int64

	filepath.Walk(pkgDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && !strings.HasPrefix(filepath.Base(path), ".") {
			fileCount++
			totalSize += info.Size()
		}
		return nil
	})

	return fileCount, totalSize
}

func analyzePackageHealth(cfg grit.Config, pkgDir string, rootConfig *grit.RootConfig) ([]string, []string) {
	var issues []string
	var suggestions []string

	// Check for missing version
	if cfg.Package.Version == "" {
		issues = append(issues, "No version specified")
		suggestions = append(suggestions, "Add a version field to track releases")
	}

	// Check for too many dependencies
	if len(cfg.Package.Dependencies) > 10 {
		issues = append(issues, fmt.Sprintf("High number of dependencies (%d)", len(cfg.Package.Dependencies)))
		suggestions = append(suggestions, "Consider reducing dependencies or splitting the package")
	}

	// Check for common files
	commonFiles := []string{"README.md", "LICENSE", "CHANGELOG.md"}
	for _, file := range commonFiles {
		if _, err := os.Stat(filepath.Join(pkgDir, file)); os.IsNotExist(err) {
			if file == "README.md" {
				issues = append(issues, "Missing README.md")
				suggestions = append(suggestions, "Add a README.md file to document the package")
			}
		}
	}

	// Check for build configuration
	if rootConfig != nil {
		hasValidBuildCmd := false
		if buildCmd, ok := cfg.Targets["build"]; ok && buildCmd != "" {
			hasValidBuildCmd = true
		} else {
			// Check type-level build command
			pkgType := getPackageTypeForAnalysis(cfg.Package.Path, rootConfig, "")
			if typeConfig, ok := rootConfig.Types[pkgType]; ok {
				if buildCmd, ok := typeConfig.Targets["build"]; ok && buildCmd != "" {
					hasValidBuildCmd = true
				}
			}
		}
		
		if !hasValidBuildCmd {
			issues = append(issues, "No build command configured")
			suggestions = append(suggestions, "Add a build target to the package or type configuration")
		}
	}

	return issues, suggestions
}

func detectCircularDependencies(depMap map[string][]string) [][]string {
	var cycles [][]string
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	var dfs func(string, []string) bool
	dfs = func(node string, path []string) bool {
		visited[node] = true
		recStack[node] = true
		path = append(path, node)

		for _, dep := range depMap[node] {
			if !visited[dep] {
				if dfs(dep, path) {
					return true
				}
			} else if recStack[dep] {
				// Found cycle
				cycleStart := -1
				for i, p := range path {
					if p == dep {
						cycleStart = i
						break
					}
				}
				if cycleStart != -1 {
					cycle := append(path[cycleStart:], dep)
					cycles = append(cycles, cycle)
				}
				return true
			}
		}

		recStack[node] = false
		return false
	}

	for pkg := range depMap {
		if !visited[pkg] {
			dfs(pkg, []string{})
		}
	}

	return cycles
}

func findCriticalPath(depMap map[string][]string) []string {
	// Find the longest dependency chain
	longest := []string{}
	visited := make(map[string]bool)

	var dfs func(string, []string) []string
	dfs = func(node string, path []string) []string {
		if visited[node] {
			return path
		}
		
		visited[node] = true
		path = append(path, node)
		currentLongest := path

		for _, dep := range depMap[node] {
			depPath := dfs(dep, append([]string{}, path...))
			if len(depPath) > len(currentLongest) {
				currentLongest = depPath
			}
		}

		visited[node] = false
		return currentLongest
	}

	for pkg := range depMap {
		path := dfs(pkg, []string{})
		if len(path) > len(longest) {
			longest = path
		}
	}

	return longest
}

func generateWorkspaceSuggestions(analysis WorkspaceAnalysis) ([]string, []string) {
	var issues []string
	var suggestions []string

	// Check for circular dependencies
	if len(analysis.CircularDeps) > 0 {
		issues = append(issues, fmt.Sprintf("Found %d circular dependencies", len(analysis.CircularDeps)))
		suggestions = append(suggestions, "Break circular dependencies by extracting common functionality")
	}

	// Check for too many orphaned packages
	if len(analysis.OrphanPackages) > analysis.TotalPackages/3 {
		issues = append(issues, "High number of orphaned packages")
		suggestions = append(suggestions, "Consider removing unused packages or adding them as dependencies")
	}

	// Check workspace structure
	if analysis.TotalPackages > 50 {
		suggestions = append(suggestions, "Consider using package groups or namespaces for better organization")
	}

	// Check dependency distribution
	avgDeps := float64(analysis.TotalDependencies) / float64(analysis.TotalPackages)
	if avgDeps > 5 {
		suggestions = append(suggestions, "High average dependencies per package - consider architectural review")
	}

	return issues, suggestions
}

func displayAnalysis(analysis WorkspaceAnalysis, formatter *output.Formatter) {
	// Overview
	formatter.Section("Workspace Overview")
	formatter.Detail(fmt.Sprintf("Total packages: %d", analysis.TotalPackages))
	formatter.Detail(fmt.Sprintf("Total dependencies: %d", analysis.TotalDependencies))
	if analysis.TotalPackages > 0 {
		avgDeps := float64(analysis.TotalDependencies) / float64(analysis.TotalPackages)
		formatter.Detail(fmt.Sprintf("Average dependencies per package: %.1f", avgDeps))
	}

	// Package types
	if len(analysis.PackagesByType) > 0 {
		formatter.NewLine()
		formatter.Info("Package Distribution by Type:")
		for pkgType, count := range analysis.PackagesByType {
			formatter.Detail(fmt.Sprintf("• %s: %d packages", pkgType, count))
		}
	}

	// Issues
	if len(analysis.Issues) > 0 {
		formatter.NewLine()
		formatter.Warning("Workspace Issues:")
		for _, issue := range analysis.Issues {
			formatter.Detail(fmt.Sprintf("• %s", issue))
		}
	}

	// Circular dependencies
	if len(analysis.CircularDeps) > 0 {
		formatter.NewLine()
		formatter.Error("Circular Dependencies Detected:")
		for _, cycle := range analysis.CircularDeps {
			formatter.Detail(fmt.Sprintf("• %s", strings.Join(cycle, " → ")))
		}
	}

	// Orphaned packages
	if len(analysis.OrphanPackages) > 0 && verboseAnalysis {
		formatter.NewLine()
		formatter.Info("Orphaned Packages (no dependents):")
		for _, pkg := range analysis.OrphanPackages {
			formatter.Detail(fmt.Sprintf("• %s", pkg))
		}
	}

	// Critical path
	if len(analysis.CriticalPath) > 0 && verboseAnalysis {
		formatter.NewLine()
		formatter.Info("Critical Path (longest dependency chain):")
		formatter.Detail(strings.Join(analysis.CriticalPath, " → "))
	}

	// Package details
	if verboseAnalysis {
		formatter.NewLine()
		formatter.Section("Package Analysis")
		
		packages := make([]string, 0, len(analysis.Packages))
		for pkg := range analysis.Packages {
			packages = append(packages, pkg)
		}
		sort.Strings(packages)

		for _, pkg := range packages {
			pkgAnalysis := analysis.Packages[pkg]
			formatter.NewLine()
			formatter.PackageInfo(pkgAnalysis.Name, pkgAnalysis.Version, pkgAnalysis.Type, pkgAnalysis.Dependencies)
			
			if len(pkgAnalysis.Issues) > 0 {
				formatter.Warning("Issues:")
				for _, issue := range pkgAnalysis.Issues {
					formatter.Detail(fmt.Sprintf("  • %s", issue))
				}
			}
			
			if len(pkgAnalysis.Suggestions) > 0 {
				formatter.Info("Suggestions:")
				for _, suggestion := range pkgAnalysis.Suggestions {
					formatter.Detail(fmt.Sprintf("  • %s", suggestion))
				}
			}
		}
	}

	// Suggestions
	if len(analysis.Suggestions) > 0 {
		formatter.NewLine()
		formatter.Section("Recommendations")
		for _, suggestion := range analysis.Suggestions {
			formatter.Detail(fmt.Sprintf("• %s", suggestion))
		}
	}
}

func outputJSON(analysis WorkspaceAnalysis) {
	// Simple JSON output without external dependencies
	fmt.Println("{")
	fmt.Printf("  \"total_packages\": %d,\n", analysis.TotalPackages)
	fmt.Printf("  \"total_dependencies\": %d,\n", analysis.TotalDependencies)
	fmt.Printf("  \"circular_dependencies\": %d,\n", len(analysis.CircularDeps))
	fmt.Printf("  \"orphan_packages\": %d\n", len(analysis.OrphanPackages))
	fmt.Println("}")
}

func loadRootConfigForAnalysis(cwd string) (*grit.RootConfig, error) {
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

func getPackageTypeForAnalysis(packagePath string, rootConfig *grit.RootConfig, cwd string) string {
	if cwd == "" {
		var err error
		cwd, err = os.Getwd()
		if err != nil {
			return ""
		}
	}

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