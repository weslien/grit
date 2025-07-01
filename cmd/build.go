package cmd

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"github.com/weslien/grit/pkg/grit"
	"github.com/weslien/grit/pkg/output"
	"gopkg.in/yaml.v3"
)

var noCache bool
var dirtyFlag bool // Add this variable declaration

var buildCmd = &cobra.Command{
	Use:   "build [type] [name]",
	Short: "Build packages and their dependencies",
	Long:  `Build packages respecting dependency order and utilizing build cache`,
	Run: func(cmd *cobra.Command, args []string) {
		formatter := output.New()

		cwd, err := os.Getwd()
		if err != nil {
			formatter.Error(fmt.Sprintf("Error getting current directory: %v", err))
			os.Exit(1)
		}

		formatter.Header("GRIT Build")
		formatter.Section("Loading Packages")

		pm := grit.NewPackageManager(cwd)
		packages, err := pm.LoadPackages()
		if err != nil {
			formatter.Error(fmt.Sprintf("Error loading packages: %v", err))
			os.Exit(1)
		}
		formatter.Success(fmt.Sprintf("Loaded %d packages", len(packages)))

		// Define cacheDir here, before it's used in the dirty flag check
		cacheDir := filepath.Join(cwd, ".grit", "cache")
		if !noCache {
			os.MkdirAll(cacheDir, 0755)
		}

		// Add this block to filter packages if --dirty flag is set
		// In the dirtyFlag check section, after loading packages

		if dirtyFlag {
			formatter.Info("Filtering packages with no changes")
			var dirtyPackages []grit.Config

			// First, build a reverse dependency map
			reverseDeps := make(map[string][]string)
			for _, cfg := range packages {
				for _, depName := range cfg.Package.Dependencies {
					reverseDeps[depName] = append(reverseDeps[depName], cfg.Package.Name)
				}
			}

			// Track directly dirty packages first
			directlyDirty := make(map[string]bool)

			for _, cfg := range packages {
				if cfg.Package.Name == "" {
					continue // Skip root config
				}

				cfgDir := filepath.Dir(cfg.Package.Path)
				newHash, err := calculatePackageHash(cfgDir)
				if err != nil {
					formatter.Warning(fmt.Sprintf("Could not calculate hash for %s: %v", cfg.Package.Name, err))
					directlyDirty[cfg.Package.Name] = true
					continue
				}

				cacheFile := filepath.Join(cacheDir, cfg.Package.Name+".hash")

				if cachedHash, err := os.ReadFile(cacheFile); err != nil {
					directlyDirty[cfg.Package.Name] = true
				} else if string(cachedHash) != newHash {
					directlyDirty[cfg.Package.Name] = true
				}
			}

			// Now propagate dirtiness to dependent packages
			allDirty := make(map[string]bool)
			for pkgName := range directlyDirty {
				allDirty[pkgName] = true
				propagateDirtiness(pkgName, reverseDeps, allDirty, formatter)
			}

			// Build the final list of dirty packages
			for _, cfg := range packages {
				if allDirty[cfg.Package.Name] {
					dirtyPackages = append(dirtyPackages, cfg)
				}
			}

			formatter.Success(fmt.Sprintf("Found %d packages with changes", len(dirtyPackages)))
			if len(directlyDirty) < len(allDirty) {
				formatter.Detail(fmt.Sprintf("%d packages are directly changed, %d are affected by dependencies",
					len(directlyDirty), len(allDirty)-len(directlyDirty)))
			}

			packages = dirtyPackages

			if len(packages) == 0 {
				formatter.Success("No packages to build")
				return
			}
		}

		formatter.Section("Resolving Dependencies")
		buildOrder, err := resolveDependencies(packages, formatter)
		if err != nil {
			formatter.Error(fmt.Sprintf("Error resolving dependencies: %v", err))
			os.Exit(1)
		}
		formatter.Success("Dependencies resolved successfully")

		// In the buildCmd.Run function, add more detailed logging
		formatter.Section("Building Packages")
		packageNames := getPackageNames(buildOrder)
		formatter.Detail(fmt.Sprintf("Build order: %s", strings.Join(packageNames, " → ")))

		// Group packages by their dependency level
		buildLevels := groupPackagesByLevel(buildOrder, formatter)
		formatter.Detail(fmt.Sprintf("Build will execute in %d parallel stages", len(buildLevels)))

		// Create overall progress bar
		totalPackages := len(packageNames)
		if totalPackages > 0 {
			progress := formatter.Progress(totalPackages, "Building packages")
			
			successCount := 0
			failedPackages := []string{}
			startTime := time.Now()
			
			for level, levelPackages := range buildLevels {
				levelStart := time.Now()
				formatter.Info(fmt.Sprintf("Stage %d/%d: Building %d packages in parallel", 
					level+1, len(buildLevels), len(levelPackages)))
				
				// Create channels for this level
				var wg sync.WaitGroup
				type buildResult struct {
					packageName string
					success     bool
					duration    time.Duration
					err         error
				}
				resultChan := make(chan buildResult, len(levelPackages))
				
				// Launch goroutines for each package at this level
				for _, cfg := range levelPackages {
					if cfg.Package.Name == "" {
						continue // Skip root config
					}
					
					wg.Add(1)
					go func(cfg grit.Config) {
						defer wg.Done()
						buildStart := time.Now()
						err := executeBuild(cfg, cacheDir, noCache, formatter, cwd)
						buildDuration := time.Since(buildStart)
						
						resultChan <- buildResult{
							packageName: cfg.Package.Name,
							success:     err == nil,
							duration:    buildDuration,
							err:         err,
						}
					}(cfg)
				}
				
				// Wait for all builds at this level to complete
				wg.Wait()
				close(resultChan)
				
				// Process results
				levelFailures := 0
				for result := range resultChan {
					progress.Add(1)
					if result.success {
						successCount++
						formatter.Detail(fmt.Sprintf("✓ %s built in %v", result.packageName, result.duration))
					} else {
						levelFailures++
						failedPackages = append(failedPackages, result.packageName)
						formatter.Detail(fmt.Sprintf("✗ %s failed: %v", result.packageName, result.err))
					}
				}
				
				levelDuration := time.Since(levelStart)
				if levelFailures > 0 {
					formatter.Warning(fmt.Sprintf("Stage %d completed with %d failures (%v)", 
						level+1, levelFailures, levelDuration))
					break // Stop on first stage failure
				} else {
					formatter.Success(fmt.Sprintf("Stage %d completed successfully (%v)", 
						level+1, levelDuration))
				}
			}
			
			progress.Close()
			totalDuration := time.Since(startTime)
			
			// Enhanced summary
			formatter.Summary(successCount, totalPackages, totalDuration)
			
			if len(failedPackages) > 0 {
				formatter.NewLine()
				formatter.Error("Failed packages:")
				for _, pkg := range failedPackages {
					formatter.Detail(fmt.Sprintf("• %s", pkg))
				}
				os.Exit(1)
			}
		} else {
			formatter.Info("No packages to build")
		}
	},
}

func init() {
	buildCmd.Flags().BoolVar(&noCache, "no-cache", false, "Bypass build cache")
	buildCmd.Flags().BoolVar(&dirtyFlag, "dirty", false, "Only build packages with changes") // Add this flag
	rootCmd.AddCommand(buildCmd)
}

func resolveDependencies(packages []grit.Config, formatter *output.Formatter) ([]grit.Config, error) {
	// Build dependency graph
	graph := make(map[string][]string)
	nodeMap := make(map[string]grit.Config)
	inDegree := make(map[string]int)

	// Initialize the graph with all packages
	for _, cfg := range packages {
		nodeMap[cfg.Package.Name] = cfg
		if _, exists := graph[cfg.Package.Name]; !exists {
			graph[cfg.Package.Name] = []string{}
		}
	}

	// Add dependencies to the graph
	for _, cfg := range packages {
		for _, depName := range cfg.Package.Dependencies {
			// Check if the dependency exists
			if _, exists := nodeMap[depName]; !exists {
				// Skip missing dependencies or handle them differently
				formatter.Warning(fmt.Sprintf("Package %s depends on %s, but it doesn't exist",
					cfg.Package.Name, depName))
				continue
			}

			graph[cfg.Package.Name] = append(graph[cfg.Package.Name], depName)
			inDegree[depName]++
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
	if len(order) != len(packages) {
		formatter.Warning("Possible dependency cycle detected. Building packages in best-effort order.")

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

	// Reverse the order to get bottom-up (dependencies first)
	reversed := make([]grit.Config, len(order))
	for i, cfg := range order {
		reversed[len(order)-1-i] = cfg
	}

	return reversed, nil
}

func executeBuild(cfg grit.Config, cacheDir string, noCache bool, formatter *output.Formatter, cwd string) error {
	// Skip if this is the root config file
	if cfg.Package.Name == "" {
		return nil
	}

	// Get the package directory from the stored path
	cfgDir := filepath.Dir(cfg.Package.Path)

	// Calculate a hash based on the package files
	// If we're using --dirty, we might have already calculated this hash
	var newHash string
	if dirtyFlag && !noCache {
		// Try to get the hash from the dirty check
		cacheFile := filepath.Join(cacheDir, cfg.Package.Name+".hash")
		if cachedHash, err := os.ReadFile(cacheFile); err == nil {
			// We have a cached hash, but we know it's dirty, so use it
			newHash = string(cachedHash)
		} else {
			// Calculate the hash
			var err error
			newHash, err = calculatePackageHash(cfgDir)
			if err != nil {
				return fmt.Errorf("failed to calculate package hash: %w", err)
			}
		}
	} else {
		// Calculate the hash normally
		var err error
		newHash, err = calculatePackageHash(cfgDir)
		if err != nil {
			return fmt.Errorf("failed to calculate package hash: %w", err)
		}
	}

	cacheFile := filepath.Join(cacheDir, cfg.Package.Name+".hash")

	if !noCache {
		if cachedHash, err := os.ReadFile(cacheFile); err == nil {
			if string(cachedHash) == newHash {
				formatter.Detail(fmt.Sprintf("Using cached build for %s", cfg.Package.Name))
				return nil
			}
			formatter.Warning(fmt.Sprintf("Cache invalidated for %s (files changed)", cfg.Package.Name))
		}
	}



	// In the executeBuild function, fix the root config path
	// Load the root config to get type information
	rootConfigPath := filepath.Join(cwd, "grit.yaml")
	formatter.Detail(fmt.Sprintf("Looking for root config at: %s", rootConfigPath))
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

	// In the executeBuild function, add timeout and better error handling for the command execution
	formatter.Detail(fmt.Sprintf("Executing build command: %s", buildCmd))

	// Execute the build command with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sh", "-c", buildCmd)
	cmd.Dir = cfgDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("build command timed out after 2 minutes")
		}
		return fmt.Errorf("build command failed: %w", err)
	}

	formatter.Success(fmt.Sprintf("Built %s successfully", cfg.Package.Name))

	// Save the new hash to the cache
	if !noCache {
		os.WriteFile(cacheFile, []byte(newHash), 0644)
	}

	return nil
}

// Add this new function to calculate a hash based on directory contents
func calculatePackageHash(pkgDir string) (string, error) {
	var fileInfos []string

	// Walk through the package directory
	err := filepath.Walk(pkgDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files we can't access
		}

		// Skip directories and hidden files/directories
		if info.IsDir() {
			if strings.HasPrefix(filepath.Base(path), ".") && path != pkgDir {
				return filepath.SkipDir // Skip hidden directories
			}
			return nil
		}

		// Skip hidden files
		if strings.HasPrefix(filepath.Base(path), ".") {
			return nil
		}

		// Add file info to our list (path, size, mod time)
		relPath, _ := filepath.Rel(pkgDir, path)
		fileInfo := fmt.Sprintf("%s:%d:%d",
			relPath,
			info.Size(),
			info.ModTime().UnixNano())
		fileInfos = append(fileInfos, fileInfo)
		return nil
	})

	if err != nil {
		return "", err
	}

	// Sort the file infos for consistent hashing
	sort.Strings(fileInfos)

	// Join all file infos and hash them
	allInfos := strings.Join(fileInfos, "|")
	hasher := sha256.New()
	hasher.Write([]byte(allInfos))

	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}

// Helper function to get package names for logging
func getPackageNames(configs []grit.Config) []string {
	names := make([]string, 0, len(configs))
	for _, cfg := range configs {
		if cfg.Package.Name != "" {
			names = append(names, cfg.Package.Name)
		}
	}
	return names
}

// Helper function to recursively propagate dirtiness to dependent packages
func propagateDirtiness(pkgName string, reverseDeps map[string][]string, allDirty map[string]bool, formatter *output.Formatter) {
	for _, depender := range reverseDeps[pkgName] {
		if !allDirty[depender] {
			formatter.Detail(fmt.Sprintf("Package %s is dirty because it depends on %s", depender, pkgName))
			allDirty[depender] = true
			// Recursively mark packages that depend on this one
			propagateDirtiness(depender, reverseDeps, allDirty, formatter)
		}
	}
}

// Helper function to group packages by their dependency level for parallel building
func groupPackagesByLevel(buildOrder []grit.Config, formatter *output.Formatter) [][]grit.Config {
    // Create a map of package name to its dependencies
    dependsOn := make(map[string]map[string]bool)
    for _, cfg := range buildOrder {
        if cfg.Package.Name == "" {
            continue
        }
        
        dependsOn[cfg.Package.Name] = make(map[string]bool)
        for _, dep := range cfg.Package.Dependencies {
            dependsOn[cfg.Package.Name][dep] = true
        }
    }
    
    // Create a map of package name to its dependents
    dependedOnBy := make(map[string]map[string]bool)
    for pkgName, deps := range dependsOn {
        for dep := range deps {
            if dependedOnBy[dep] == nil {
                dependedOnBy[dep] = make(map[string]bool)
            }
            dependedOnBy[dep][pkgName] = true
        }
    }
    
    // Group packages by levels
    var levels [][]grit.Config
    remaining := make(map[string]grit.Config)
    
    // Initialize remaining packages
    for _, cfg := range buildOrder {
        if cfg.Package.Name != "" {
            remaining[cfg.Package.Name] = cfg
        }
    }
    
    // Continue until all packages are assigned to levels
    for len(remaining) > 0 {
        var currentLevel []grit.Config
        
        // Find packages with no remaining dependencies
        for pkgName, cfg := range remaining {
            canBuild := true
            for dep := range dependsOn[pkgName] {
                if _, exists := remaining[dep]; exists {
                    canBuild = false
                    break
                }
            }
            
            if canBuild {
                currentLevel = append(currentLevel, cfg)
            }
        }
        
        // In the groupPackagesByLevel function, there's an unused variable in the cycle detection section
        if len(currentLevel) == 0 && len(remaining) > 0 {
        formatter.Warning("Possible dependency cycle detected. Breaking cycle to continue build.")
        for _, cfg := range remaining {
        currentLevel = append(currentLevel, cfg)
        break
        }
        }
        
        // Sort the current level by dependency count (packages with more dependents first)
        sort.Slice(currentLevel, func(i, j int) bool {
            nameI := currentLevel[i].Package.Name
            nameJ := currentLevel[j].Package.Name
            return len(dependedOnBy[nameI]) > len(dependedOnBy[nameJ])
        })
        
        // Remove the packages from remaining
        for _, cfg := range currentLevel {
            delete(remaining, cfg.Package.Name)
        }
        
        levels = append(levels, currentLevel)
    }
    
    return levels
}
