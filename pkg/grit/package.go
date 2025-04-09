package grit

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3" // Add this import
)

type Package struct {
	Name         string
	Version      string
	Dependencies []string
	Hash         string
	Path         string // Add this field to store the path to grit.yaml
}

type PackageManager struct {
	workspaceRoot string
}

func NewPackageManager(root string) *PackageManager {
	return &PackageManager{
		workspaceRoot: root,
	}
}

func (pm *PackageManager) LoadPackages() ([]Config, error) {
	var packages []Config

	err := filepath.Walk(pm.workspaceRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		//fmt.Printf("Checking %s\n", path)
		if info.Name() == "grit.yaml" {
			//fmt.Printf("Found %s\n", path)
			cfg, err := parsePackageFile(path)
			if err != nil {
				return fmt.Errorf("error parsing %s: %w", path, err)
			}
			//fmt.Printf("Loaded %s\n", cfg.Package.Name)
			packages = append(packages, *cfg)
		}
		return nil
	})

	return packages, err
}

func parsePackageFile(path string) (*Config, error) {
	//fmt.Printf("Parsing %s\n", path)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}
	//fmt.Printf("Parsed %s\n", cfg)
	// Set the path to the grit.yaml file
	cfg.Package.Path = path
	return &cfg, nil
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
	Package Package               `yaml:"package"`
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
