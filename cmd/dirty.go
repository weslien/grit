package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/weslien/grit/pkg/grit"
	"github.com/weslien/grit/pkg/output"
)

var dirtyCmd = &cobra.Command{
	Use:   "dirty",
	Short: "List packages with changes",
	Long:  `List all packages that have changes compared to their cached state`,
	Run: func(cmd *cobra.Command, args []string) {
		formatter := output.New()
		
		cwd, err := os.Getwd()
		if err != nil {
			formatter.Error(fmt.Sprintf("Error getting current directory: %v", err))
			os.Exit(1)
		}

		formatter.Header("GRIT Dirty Packages")
		formatter.Section("Loading Packages")
		
		pm := grit.NewPackageManager(cwd)
		packages, err := pm.LoadPackages()
		if err != nil {
			formatter.Error(fmt.Sprintf("Error loading packages: %v", err))
			os.Exit(1)
		}
		formatter.Success(fmt.Sprintf("Loaded %d packages", len(packages)))

		formatter.Section("Checking for Changes")
		
		cacheDir := filepath.Join(cwd, ".grit", "cache")
		os.MkdirAll(cacheDir, 0755)
		
		var dirtyPackages []grit.Config
		
		for _, cfg := range packages {
			if cfg.Package.Name == "" {
				continue // Skip root config
			}
			
			cfgDir := filepath.Dir(cfg.Package.Path)
			newHash, err := calculatePackageHash(cfgDir)
			if err != nil {
				formatter.Warning(fmt.Sprintf("Could not calculate hash for %s: %v", cfg.Package.Name, err))
				dirtyPackages = append(dirtyPackages, cfg) // Include if we can't determine
				continue
			}
			
			cacheFile := filepath.Join(cacheDir, cfg.Package.Name+".hash")
			isDirty := false
			
			if cachedHash, err := os.ReadFile(cacheFile); err != nil {
				formatter.Detail(fmt.Sprintf("%s: No cache found", cfg.Package.Name))
				isDirty = true
			} else if string(cachedHash) != newHash {
				formatter.Detail(fmt.Sprintf("%s: Files changed", cfg.Package.Name))
				isDirty = true
			}
			
			if isDirty {
				dirtyPackages = append(dirtyPackages, cfg)
			}
		}
		
		formatter.Section("Results")
		if len(dirtyPackages) == 0 {
			formatter.Success("No dirty packages found")
		} else {
			formatter.Info(fmt.Sprintf("Found %d dirty packages:", len(dirtyPackages)))
			for _, pkg := range dirtyPackages {
				formatter.Detail(pkg.Package.Name)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(dirtyCmd)
}