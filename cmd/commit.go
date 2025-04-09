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
	
	// Show summary of changes first
	cmd := exec.Command("git", "status", "-s", pkgPath)
	statusOutput, err := cmd.Output()
	if err != nil {
		formatter.Warning(fmt.Sprintf("Failed to get status for %s: %v", pkg.Package.Name, err))
	} else if len(statusOutput) > 0 {
		formatter.Detail("Summary of changes:")
		fmt.Println(string(statusOutput))
	}
	
	// Ask if user wants to see the complete diff
	reader := bufio.NewReader(os.Stdin)
	formatter.Info("View complete diff? (y/n):")
	viewDiff, _ := reader.ReadString('\n')
	viewDiff = strings.TrimSpace(viewDiff)
	
	if strings.ToLower(viewDiff) == "y" || strings.ToLower(viewDiff) == "yes" {
		// First, temporarily add all files in the package to the index
		// This allows us to see the diff for new files too
		tempAddCmd := exec.Command("git", "add", "-N", pkgPath)
		tempAddCmd.Run() // Ignore errors, we'll still try to show what we can
		
		// Show diff for all files (including new ones)
		diffCmd := exec.Command("git", "diff", pkgPath)
		diffCmd.Stdout = os.Stdout
		diffCmd.Stderr = os.Stderr
		diffCmd.Stdin = os.Stdin
		
		formatter.Detail("Changes:")
		err := diffCmd.Run()
		if err != nil {
			formatter.Warning(fmt.Sprintf("Failed to display diff for %s: %v", pkg.Package.Name, err))
		}
		
		// Also show staged changes if any
		stagedCmd := exec.Command("git", "diff", "--cached", pkgPath)
		stagedCmd.Stdout = os.Stdout
		stagedCmd.Stderr = os.Stderr
		stagedCmd.Stdin = os.Stdin
		
		formatter.Detail("Staged changes:")
		err = stagedCmd.Run()
		if err != nil {
			formatter.Warning(fmt.Sprintf("Failed to display staged changes for %s: %v", pkg.Package.Name, err))
		}
		
		// Show untracked files
		untrackedCmd := exec.Command("git", "ls-files", "--others", "--exclude-standard", pkgPath)
		untrackedOutput, err := untrackedCmd.Output()
		if err != nil {
			formatter.Warning(fmt.Sprintf("Failed to get untracked files for %s: %v", pkg.Package.Name, err))
		} else if len(untrackedOutput) > 0 {
			formatter.Detail("Untracked files:")
			fmt.Println(string(untrackedOutput))
			
			// For each untracked file, show its content
			files := strings.Split(strings.TrimSpace(string(untrackedOutput)), "\n")
			for _, file := range files {
				if file == "" {
					continue
				}
				
				formatter.Detail(fmt.Sprintf("Content of new file: %s", file))
				catCmd := exec.Command("cat", file)
				catCmd.Stdout = os.Stdout
				catCmd.Stderr = os.Stderr
				catCmd.Run() // Ignore errors
				fmt.Println() // Add a newline after file content
			}
		}
		
		// Reset any temporary adds we did
		resetCmd := exec.Command("git", "reset", pkgPath)
		resetCmd.Run() // Ignore errors
	}
	
	// Ask for commit message
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
	
	// Show summary of changes first
	cmd := exec.Command("git", "status", "-s")
	statusOutput, err := cmd.Output()
	if err != nil {
		formatter.Warning(fmt.Sprintf("Failed to get repo status: %v", err))
	} else if len(statusOutput) > 0 {
		formatter.Detail("Summary of changes:")
		fmt.Println(string(statusOutput))
	}
	
	// Ask if user wants to see the complete diff
	reader := bufio.NewReader(os.Stdin)
	formatter.Info("View complete diff? (y/n):")
	viewDiff, _ := reader.ReadString('\n')
	viewDiff = strings.TrimSpace(viewDiff)
	
	if strings.ToLower(viewDiff) == "y" || strings.ToLower(viewDiff) == "yes" {
		// First, temporarily add all files to the index
		// This allows us to see the diff for new files too
		tempAddCmd := exec.Command("git", "add", "-N", ".")
		tempAddCmd.Run() // Ignore errors, we'll still try to show what we can
		
		// Show diff for all files (including new ones)
		diffCmd := exec.Command("git", "diff")
		diffCmd.Stdout = os.Stdout
		diffCmd.Stderr = os.Stderr
		diffCmd.Stdin = os.Stdin
		
		formatter.Detail("Changes:")
		err := diffCmd.Run()
		if err != nil {
			formatter.Warning(fmt.Sprintf("Failed to display repo diff: %v", err))
		}
		
		// Also show staged changes if any
		stagedCmd := exec.Command("git", "diff", "--cached")
		stagedCmd.Stdout = os.Stdout
		stagedCmd.Stderr = os.Stderr
		stagedCmd.Stdin = os.Stdin
		
		formatter.Detail("Staged changes:")
		err = stagedCmd.Run()
		if err != nil {
			formatter.Warning(fmt.Sprintf("Failed to display staged repo changes: %v", err))
		}
		
		// Show untracked files
		untrackedCmd := exec.Command("git", "ls-files", "--others", "--exclude-standard")
		untrackedOutput, err := untrackedCmd.Output()
		if err != nil {
			formatter.Warning(fmt.Sprintf("Failed to get untracked repo files: %v", err))
		} else if len(untrackedOutput) > 0 {
			formatter.Detail("Untracked files:")
			fmt.Println(string(untrackedOutput))
			
			// For each untracked file, show its content
			files := strings.Split(strings.TrimSpace(string(untrackedOutput)), "\n")
			for _, file := range files {
				if file == "" {
					continue
				}
				
				formatter.Detail(fmt.Sprintf("Content of new file: %s", file))
				catCmd := exec.Command("cat", file)
				catCmd.Stdout = os.Stdout
				catCmd.Stderr = os.Stderr
				catCmd.Run() // Ignore errors
				fmt.Println() // Add a newline after file content
			}
		}
		
		// Reset any temporary adds we did
		resetCmd := exec.Command("git", "reset")
		resetCmd.Run() // Ignore errors
	}
	
	// Ask for commit message
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