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

type Config struct {
	Targets []string              `yaml:"targets"`
	Types   map[string]TypeConfig `yaml:"types"`
	Package PackageConfig         `yaml:"package"`
}

type TypeConfig struct {
	PackageDir   string            `yaml:"packageDir"`
	BuildDir     string            `yaml:"buildDir"`
	CoverageDir  string            `yaml:"coverageDir"`
	DefaultTasks map[string]string `yaml:"defaultTasks"`
}

type PackageConfig struct {
	Version      string   `yaml:"version"`
	Name         string   `yaml:"name"`
	Dependencies []string `yaml:"dependencies"`
	Hash         string   `yaml:"hash"`
}
