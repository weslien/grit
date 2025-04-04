package grit

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Package struct {
	Name         string
	Version      string
	Dependencies []Dependency
	Hash         string
}

type Dependency struct {
	Name    string
	Version string
}

type PackageManager struct {
	workspaceRoot string
}

func NewPackageManager(root string) *PackageManager {
	return &PackageManager{
		workspaceRoot: root,
	}
}

func (pm *PackageManager) LoadPackages() ([]Package, error) {
	var packages []Package

	err := filepath.Walk(pm.workspaceRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.Name() == "grit.yaml" {
			pkg, err := parsePackageFile(path)
			if err != nil {
				return fmt.Errorf("error parsing %s: %w", path, err)
			}
			packages = append(packages, *pkg)
		}
		return nil
	})

	return packages, err
}

func parsePackageFile(path string) (*Package, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var pkg Package
	err = json.Unmarshal(data, &pkg)
	return &pkg, err
}

/**
 * The root grit.yaml config file
 */
type RootConfig struct {
	Repo    RepoConfig            `yaml:"repo"`
	Targets map[string]string     `yaml:"targets"`
	Types   map[string]TypeConfig `yaml:"types"`
}

/**
 * The config for a package type
 */
type TypeConfig struct {
	PackageDir  string            `yaml:"package_dir"`
	BuildDir    string            `yaml:"build_dir"`
	CoverageDir string            `yaml:"coverage_dir"`
	Targets     map[string]string `yaml:"targets"`
	CanDependOn []string          `yaml:"can_depend_on"`
}

/**
 * The grit.yaml config file for a package
 */
type Config struct {
	Targets map[string]string     `yaml:"targets"`
	Types   map[string]TypeConfig `yaml:"types"`
	Package PackageConfig         `yaml:"package"`
}

/**
 * The package config section
 */
type PackageConfig struct {
	Version      string   `yaml:"version"`
	Name         string   `yaml:"name"`
	Dependencies []string `yaml:"dependencies"`
	Hash         string   `yaml:"hash"`
}

/**
 * The repo config section
 */
type RepoConfig struct {
	URL     string `yaml:"url"`
	Name    string `yaml:"name"`
	License string `yaml:"license"`
	Owner   string `yaml:"owner"`
}
