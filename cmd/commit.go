package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/weslien/grit/pkg/grit"
	"github.com/weslien/grit/pkg/output"
)

// Fix the formatter initialization and add the loadPackages function
var commitCmd = &cobra.Command{
	Use:   "commit",
	Short: "Commit changes in packages",
	Long:  `Iterate through packages with changes, summarize changes, and commit them individually.`,
	Run: func(cmd *cobra.Command, args []string) {
		formatter := output.New()
		formatter.Section("Grit Commit")

		// Get current working directory
		cwd, err := os.Getwd()
		if err != nil {
			formatter.Error(fmt.Sprintf("Failed to get current directory: %v", err))
			os.Exit(1)
		}

		// Load packages
		pm := grit.NewPackageManager(cwd)
		packages, err := pm.LoadPackages()
		if err != nil {
			formatter.Error(fmt.Sprintf("Failed to load packages: %v", err))
			os.Exit(1)
		}
		
		// Find packages with changes
		packagesWithChanges := findPackagesWithChanges(packages, formatter)
		
		// Check for non-package changes
		hasRepoChanges := checkForRepoChanges(packages, cwd, formatter)

		if len(packagesWithChanges) == 0 && !hasRepoChanges {
			formatter.Success("No changes to commit")
			return
		}

		// Process packages with changes
		for _, pkg := range packagesWithChanges {
			commitPackageChanges(pkg, cwd, formatter)
		}

		// Process repo-level changes if any
		if hasRepoChanges {
			commitRepoChanges(cwd, formatter)
		}

		formatter.Success("Commit process completed")
	},
}

func init() {
	rootCmd.AddCommand(commitCmd)
}

// Find packages with changes
func findPackagesWithChanges(packages []grit.Config, formatter *output.Formatter) []grit.Config {
	var packagesWithChanges []grit.Config
	
	for _, cfg := range packages {
		if cfg.Package.Name == "" {
			continue // Skip root config
		}
		
		pkgPath := filepath.Dir(cfg.Package.Path)
		
		// Check if package has changes
		cmd := exec.Command("git", "status", "--porcelain", pkgPath)
		output, err := cmd.Output()
		if err != nil {
			formatter.Warning(fmt.Sprintf("Failed to check git status for %s: %v", cfg.Package.Name, err))
			continue
		}
		
		if len(output) > 0 {
			packagesWithChanges = append(packagesWithChanges, cfg)
		}
	}
	
	formatter.Info(fmt.Sprintf("Found %d packages with changes", len(packagesWithChanges)))
	return packagesWithChanges
}

// Check for changes outside of packages
func checkForRepoChanges(packages []grit.Config, cwd string, formatter *output.Formatter) bool {
	// Get all changes
	cmd := exec.Command("git", "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		formatter.Warning(fmt.Sprintf("Failed to check git status: %v", err))
		return false
	}
	
	if len(output) == 0 {
		return false
	}
	
	// Create a map of package paths
	packagePaths := make(map[string]bool)
	for _, cfg := range packages {
		if cfg.Package.Name != "" {
			pkgPath := filepath.Dir(cfg.Package.Path)
			relPath, err := filepath.Rel(cwd, pkgPath)
			if err == nil {
				packagePaths[relPath] = true
			}
		}
	}
	
	// Check if there are changes outside package paths
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if len(line) > 3 {
			filePath := strings.TrimSpace(line[3:])
			inPackage := false
			
			for pkgPath := range packagePaths {
				if strings.HasPrefix(filePath, pkgPath) {
					inPackage = true
					break
				}
			}
			
			if !inPackage {
				return true
			}
		}
	}
	
	return false
}

// Commit changes for a specific package
func commitPackageChanges(pkg grit.Config, cwd string, formatter *output.Formatter) {
	pkgPath := filepath.Dir(pkg.Package.Path)
	
	formatter.Section(fmt.Sprintf("Package: %s", pkg.Package.Name))
	
	// Show changes
	cmd := exec.Command("git", "status", "-s", pkgPath)
	output, err := cmd.Output()
	if err != nil {
		formatter.Error(fmt.Sprintf("Failed to get status for %s: %v", pkg.Package.Name, err))
		return
	}
	
	formatter.Detail("Changes:")
	formatter.Detail(string(output))
	
	// Show diff
	cmd = exec.Command("git", "diff", "--stat", pkgPath)
	output, err = cmd.Output()
	if err != nil {
		formatter.Warning(fmt.Sprintf("Failed to get diff for %s: %v", pkg.Package.Name, err))
	} else {
		formatter.Detail("Diff summary:")
		formatter.Detail(string(output))
	}
	
	// Ask for commit message
	reader := bufio.NewReader(os.Stdin)
	formatter.Info(fmt.Sprintf("Enter commit message for %s (or 'skip' to skip):", pkg.Package.Name))
	message, _ := reader.ReadString('\n')
	message = strings.TrimSpace(message)
	
	if message == "skip" {
		formatter.Info("Skipping commit for this package")
		return
	}
	
	// Commit changes
	cmd = exec.Command("git", "add", pkgPath)
	err = cmd.Run()
	if err != nil {
		formatter.Error(fmt.Sprintf("Failed to stage changes for %s: %v", pkg.Package.Name, err))
		return
	}
	
	commitMsg := fmt.Sprintf("%s: %s", pkg.Package.Name, message)
	cmd = exec.Command("git", "commit", "-m", commitMsg)
	err = cmd.Run()
	if err != nil {
		formatter.Error(fmt.Sprintf("Failed to commit changes for %s: %v", pkg.Package.Name, err))
		return
	}
	
	formatter.Success(fmt.Sprintf("Committed changes for %s", pkg.Package.Name))
}

// Commit changes at the repo level
func commitRepoChanges(cwd string, formatter *output.Formatter) {
	formatter.Section("Repository Changes")
	
	// Show changes outside of packages
	cmd := exec.Command("git", "status", "-s")
	output, err := cmd.Output()
	if err != nil {
		formatter.Error(fmt.Sprintf("Failed to get repo status: %v", err))
		return
	}
	
	formatter.Detail("Changes:")
	formatter.Detail(string(output))
	
	// Show diff
	cmd = exec.Command("git", "diff", "--stat")
	output, err = cmd.Output()
	if err != nil {
		formatter.Warning(fmt.Sprintf("Failed to get repo diff: %v", err))
	} else {
		formatter.Detail("Diff summary:")
		formatter.Detail(string(output))
	}
	
	// Ask for commit message
	reader := bufio.NewReader(os.Stdin)
	formatter.Info("Enter commit message for repository changes (or 'skip' to skip):")
	message, _ := reader.ReadString('\n')
	message = strings.TrimSpace(message)
	
	if message == "skip" {
		formatter.Info("Skipping commit for repository changes")
		return
	}
	
	// Commit changes
	cmd = exec.Command("git", "add", ".")
	err = cmd.Run()
	if err != nil {
		formatter.Error(fmt.Sprintf("Failed to stage repository changes: %v", err))
		return
	}
	
	cmd = exec.Command("git", "commit", "-m", message)
	err = cmd.Run()
	if err != nil {
		formatter.Error(fmt.Sprintf("Failed to commit repository changes: %v", err))
		return
	}
	
	formatter.Success("Committed repository changes")
}