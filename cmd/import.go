package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/weslien/grit/pkg/grit"
	"github.com/weslien/grit/pkg/output"
	"gopkg.in/yaml.v3"
)

var importCmd = &cobra.Command{
	Use:   "import [source] [type] [name]",
	Short: "Import code from a GitHub repo or local path",
	Long:  `Create a new package by importing code from a GitHub repository or local path.`,
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		formatter := output.New()
		formatter.Section("Grit Import")

		source := args[0]
		pkgType := args[1]
		pkgName := args[2]

		// Get current working directory
		cwd, err := os.Getwd()
		if err != nil {
			formatter.Error(fmt.Sprintf("Failed to get current directory: %v", err))
			os.Exit(1)
		}

		// Load root config
		rootConfigPath := filepath.Join(cwd, "grit.yaml")
		rootConfigData, err := os.ReadFile(rootConfigPath)
		if err != nil {
			formatter.Error(fmt.Sprintf("Failed to read root config: %v", err))
			os.Exit(1)
		}

		var rootConfig grit.RootConfig
		if err := yaml.Unmarshal(rootConfigData, &rootConfig); err != nil {
			formatter.Error(fmt.Sprintf("Invalid root config: %v", err))
			os.Exit(1)
		}

		// Check if the package type exists
		typeConfig, exists := rootConfig.Types[pkgType]
		if !exists {
			formatter.Error(fmt.Sprintf("Package type '%s' does not exist", pkgType))
			os.Exit(1)
		}

		// Determine the package directory
		pkgDir := filepath.Join(cwd, typeConfig.PackageDir, pkgName)

		// Check if the package already exists
		if _, err := os.Stat(pkgDir); !os.IsNotExist(err) {
			formatter.Error(fmt.Sprintf("Package '%s' already exists at %s", pkgName, pkgDir))
			os.Exit(1)
		}

		// Create the package directory
		if err := os.MkdirAll(pkgDir, 0755); err != nil {
			formatter.Error(fmt.Sprintf("Failed to create package directory: %v", err))
			os.Exit(1)
		}

		// Import the source
		if strings.HasPrefix(source, "https://github.com/") || strings.HasPrefix(source, "git@github.com:") {
			importFromGitHub(source, pkgDir, formatter)
		} else {
			importFromLocalPath(source, pkgDir, formatter)
		}

		// Create the package config file
		createPackageConfig(pkgDir, pkgName, pkgType, formatter)

		formatter.Success(fmt.Sprintf("Successfully imported '%s' as package '%s' of type '%s'", source, pkgName, pkgType))
	},
}

func init() {
	rootCmd.AddCommand(importCmd)
}

// Import code from a GitHub repository
func importFromGitHub(repo string, pkgDir string, formatter *output.Formatter) {
	formatter.Info(fmt.Sprintf("Cloning from GitHub: %s", repo))

	// Create a temporary directory for the clone
	tempDir, err := os.MkdirTemp("", "grit-import-*")
	if err != nil {
		formatter.Error(fmt.Sprintf("Failed to create temporary directory: %v", err))
		os.Exit(1)
	}
	defer os.RemoveAll(tempDir)

	// Clone the repository
	cmd := exec.Command("git", "clone", "--depth=1", repo, tempDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		formatter.Error(fmt.Sprintf("Failed to clone repository: %v", err))
		os.Exit(1)
	}

	// Remove .git directory
	os.RemoveAll(filepath.Join(tempDir, ".git"))

	// Copy files from temp dir to package dir
	copyDir(tempDir, pkgDir, formatter)
}

// Import code from a local path
func importFromLocalPath(path string, pkgDir string, formatter *output.Formatter) {
	formatter.Info(fmt.Sprintf("Importing from local path: %s", path))

	// Check if the source path exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		formatter.Error(fmt.Sprintf("Source path '%s' does not exist", path))
		os.Exit(1)
	}

	// Copy files from source path to package dir
	copyDir(path, pkgDir, formatter)
}

// Copy directory contents recursively
func copyDir(src string, dst string, formatter *output.Formatter) {
	// Get file info
	info, err := os.Stat(src)
	if err != nil {
		formatter.Error(fmt.Sprintf("Failed to get source info: %v", err))
		os.Exit(1)
	}

	// If source is a file, just copy it
	if !info.IsDir() {
		copyFile(src, dst, formatter)
		return
	}

	// Create destination directory
	if err := os.MkdirAll(dst, info.Mode()); err != nil {
		formatter.Error(fmt.Sprintf("Failed to create destination directory: %v", err))
		os.Exit(1)
	}

	// Read directory contents
	entries, err := os.ReadDir(src)
	if err != nil {
		formatter.Error(fmt.Sprintf("Failed to read source directory: %v", err))
		os.Exit(1)
	}

	// Copy each entry
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		// Skip .git directories
		if entry.Name() == ".git" {
			continue
		}

		if entry.IsDir() {
			// Recursively copy subdirectory
			copyDir(srcPath, dstPath, formatter)
		} else {
			// Copy file
			copyFile(srcPath, dstPath, formatter)
		}
	}
}

// Copy a single file
func copyFile(src string, dst string, formatter *output.Formatter) {
	// Read source file
	data, err := os.ReadFile(src)
	if err != nil {
		formatter.Warning(fmt.Sprintf("Failed to read source file %s: %v", src, err))
		return
	}

	// Write to destination file
	if err := os.WriteFile(dst, data, 0644); err != nil {
		formatter.Warning(fmt.Sprintf("Failed to write destination file %s: %v", dst, err))
		return
	}
}

// Create the package config file (grit.yaml)
func createPackageConfig(pkgDir string, pkgName string, pkgType string, formatter *output.Formatter) {
	formatter.Info("Creating package configuration")

	// Create a basic package config
	config := grit.Config{
		Targets: map[string]string{
			"build": "echo 'Implement build logic'",
			"test":  "echo 'Implement test logic'",
		},
		Types: map[string]grit.TypeConfig{},
		Package: grit.Package{
			Name:         pkgName,
			Version:      "0.1.0",
			Dependencies: []string{},
			Hash:         "",
			Path:         "",
		},
	}

	// Marshal to YAML
	data, err := yaml.Marshal(config)
	if err != nil {
		formatter.Error(fmt.Sprintf("Failed to create package config: %v", err))
		os.Exit(1)
	}

	// Write to file
	configPath := filepath.Join(pkgDir, "grit.yaml")
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		formatter.Error(fmt.Sprintf("Failed to write package config: %v", err))
		os.Exit(1)
	}

	formatter.Success(fmt.Sprintf("Created package config at %s", configPath))
}
