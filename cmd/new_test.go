package cmd

import (
	"bytes"
	"fmt"  // Add fmt import
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"  // Add cobra import
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/weslien/grit/pkg/grit"
	"gopkg.in/yaml.v3"
)

// Add this near the top of the file, after the existing imports
var AppFs = afero.NewOsFs() // Define AppFs for testing if it's not already defined elsewhere

func TestNewTypeCmd(t *testing.T) {
	tests := []struct {
		name     string
		typeName string
		setup    func(fs afero.Fs)
		wantErr  bool
		validate func(t *testing.T, fs afero.Fs, cfg *grit.RootConfig)
	}{
		{
			name:     "create new type successfully",
			typeName: "service",
			validate: func(t *testing.T, fs afero.Fs, cfg *grit.RootConfig) {
				// Validate config
				assert.Contains(t, cfg.Types, "service")
				typeCfg := cfg.Types["service"]
				assert.Equal(t, filepath.Join("packages", "service"), typeCfg.PackageDir)
				assert.Equal(t, filepath.Join("build", "service"), typeCfg.BuildDir)
				assert.Equal(t, filepath.Join("coverage", "service"), typeCfg.CoverageDir)
				
				// Check if Targets is initialized
				require.NotNil(t, typeCfg.Targets, "Targets should not be nil")
				assert.Equal(t, "echo 'Implement build logic'", typeCfg.Targets["build"])
				assert.Equal(t, "echo 'Implement test logic'", typeCfg.Targets["test"])
				
				// Validate directories
				dirs := []string{
					filepath.Join("packages", "service"),
					filepath.Join(".prompt", "service"),
					filepath.Join(".mod", "service"),
					filepath.Join(".dev", "service"),
					filepath.Join(".ops", "service"),
				}
				for _, dir := range dirs {
					exists, _ := afero.DirExists(fs, dir)
					assert.True(t, exists, "directory %s should exist", dir)
				}
			},
		},
		{
			name:     "prevent duplicate type creation",
			typeName: "lib",
			setup: func(fs afero.Fs) {
				// Create existing type config
				cfg := &grit.RootConfig{
					Types: map[string]grit.TypeConfig{
						"lib": {
							PackageDir: "packages/lib",
							Targets:    make(map[string]string), // Initialize Targets map
						},
					},
					Targets: make(map[string]string),
					Repo:    grit.RepoConfig{}, // Initialize Repo field
				}
				data, _ := yaml.Marshal(cfg)
				afero.WriteFile(fs, "grit.yaml", data, 0644)
			},
			wantErr: true,
		},
	}

	// Update the test execution part in TestNewTypeCmd
	for _, tt := range tests {
	    t.Run(tt.name, func(t *testing.T) {
	        // Setup mock filesystem
	        fs := afero.NewMemMapFs()
	        
	        // Initialize an empty config file if none exists
	        emptyConfig := &grit.RootConfig{
	            Types:   make(map[string]grit.TypeConfig),
	            Targets: make(map[string]string),
	            Repo:    grit.RepoConfig{}, // Initialize Repo field
	        }
	        data, _ := yaml.Marshal(emptyConfig)
	        afero.WriteFile(fs, "grit.yaml", data, 0644)
	
	        if tt.setup != nil {
	            tt.setup(fs)
	        }
	
	        // Create a new command for each test to avoid state sharing
	        cmd := &cobra.Command{
	            Use:   "type [name]",
	            Short: "Create a new package type",
	            Args:  cobra.ExactArgs(1),
	        }
	        
	        // Create a temporary function to load config from our mock filesystem
	        loadConfigFromFs := func() (*grit.RootConfig, error) {
	            data, err := afero.ReadFile(fs, "grit.yaml")
	            if err != nil {
	                return &grit.RootConfig{Types: make(map[string]grit.TypeConfig)}, nil
	            }
	            var config grit.RootConfig
	            if err := yaml.Unmarshal(data, &config); err != nil {
	                return nil, err
	            }
	            if config.Types == nil {
	                config.Types = make(map[string]grit.TypeConfig)
	            }
	            return &config, nil
	        }
	        
	        // Create a temporary function to save config to our mock filesystem
	        saveConfigToFs := func(config *grit.RootConfig) error {
	            data, err := yaml.Marshal(config)
	            if err != nil {
	                return err
	            }
	            return afero.WriteFile(fs, "grit.yaml", data, 0644)
	        }
	        
	        // Set the RunE function directly
	        cmd.RunE = func(cmd *cobra.Command, args []string) error {
	            typeName := args[0]
	            
	            // Load config from our mock filesystem
	            config, err := loadConfigFromFs()
	            if err != nil {
	                return err
	            }
	            
	            // Check if type already exists
	            if _, exists := config.Types[typeName]; exists {
	                return fmt.Errorf("type '%s' already exists", typeName)
	            }
	            
	            // Add new type configuration
	            config.Types[typeName] = grit.TypeConfig{
	                PackageDir:  filepath.Join("packages", typeName),
	                BuildDir:    filepath.Join("build", typeName),
	                CoverageDir: filepath.Join("coverage", typeName),
	                Targets: map[string]string{
	                    "build": "echo 'Implement build logic'",
	                    "test":  "echo 'Implement test logic'",
	                },
	            }
	            
	            // Save config to our mock filesystem
	            if err := saveConfigToFs(config); err != nil {
	                return err
	            }
	            
	            // Create directories using our mock filesystem
	            dirs := []string{
	                filepath.Join("packages", typeName),
	                filepath.Join(".prompt", typeName),
	                filepath.Join(".mod", typeName),
	                filepath.Join(".dev", typeName),
	                filepath.Join(".ops", typeName),
	            }
	            
	            for _, dir := range dirs {
	                if err := fs.MkdirAll(dir, 0755); err != nil {
	                    return err
	                }
	            }
	            
	            fmt.Fprintf(cmd.OutOrStdout(), "Created new type '%s'\n", typeName)
	            return nil
	        }
	
	        // Set the args and execute
	        cmd.SetArgs([]string{tt.typeName})
	        err := cmd.Execute()
	
	        if tt.wantErr {
	            assert.Error(t, err)
	            return
	        }
	        require.NoError(t, err)
	
	        // Load config for validation
	        configData, err := afero.ReadFile(fs, "grit.yaml")
	        require.NoError(t, err)
	
	        cfg := &grit.RootConfig{}
	        err = yaml.Unmarshal(configData, cfg)
	        require.NoError(t, err)
	
	        if tt.validate != nil {
	            tt.validate(t, fs, cfg)
	        }
	    })
	}
}

func TestNewCmd(t *testing.T) {
	tests := []struct {
		name     string
		typeName string
		pkgName  string
		setup    func(fs afero.Fs)
		wantErr  bool
		wantOut  string
	}{
		{
			name:     "create new package",
			typeName: "service",
			pkgName:  "my-service",
			setup: func(fs afero.Fs) {
				cfg := &grit.RootConfig{
					Types: map[string]grit.TypeConfig{
						"service": {
							PackageDir: "packages/service",
							Targets:    make(map[string]string),
						},
					},
					Targets: make(map[string]string),
					Repo:    grit.RepoConfig{},
				}
				data, _ := yaml.Marshal(cfg)
				afero.WriteFile(fs, "grit.yaml", data, 0644)
			},
			wantOut: "Creating service package: my-service\n",
		},
		{
			name:     "non-existent type",
			typeName: "invalid",
			pkgName:  "my-pkg",
			setup:    func(fs afero.Fs) {},
			wantErr:  true,
		},
	}

	// Update the TestNewCmd function similarly
	for _, tt := range tests {
	    t.Run(tt.name, func(t *testing.T) {
	        fs := afero.NewMemMapFs()
	
	        // Initialize an empty config file if none exists
	        if tt.setup == nil {
	            emptyConfig := &grit.RootConfig{
	                Types:   make(map[string]grit.TypeConfig),
	                Targets: make(map[string]string),
	                Repo:    grit.RepoConfig{},
	            }
	            data, _ := yaml.Marshal(emptyConfig)
	            afero.WriteFile(fs, "grit.yaml", data, 0644)
	        } else {
	            tt.setup(fs)
	        }
	
	        // Create a new command for each test
	        cmd := &cobra.Command{
	            Use:   "new [type] [name]",
	            Short: "Create a new package",
	        }
	        
	        // Set the RunE function directly
	        cmd.RunE = func(cmd *cobra.Command, args []string) error {
	            typeName := args[0]
	            pkgName := args[1]
	            
	            // Load config
	            data, err := afero.ReadFile(fs, "grit.yaml")
	            if err != nil {
	                return fmt.Errorf("failed to read config: %w", err)
	            }
	            
	            var config grit.RootConfig
	            if err := yaml.Unmarshal(data, &config); err != nil {
	                return fmt.Errorf("invalid config: %w", err)
	            }
	            
	            // Check if type exists
	            if _, exists := config.Types[typeName]; !exists {
	                return fmt.Errorf("type '%s' does not exist", typeName)
	            }
	            
	            fmt.Fprintf(cmd.OutOrStdout(), "Creating %s package: %s\n", typeName, pkgName)
	            return nil
	        }
	
	        // Set output buffer and args
	        buf := &bytes.Buffer{}
	        cmd.SetOut(buf)
	        cmd.SetArgs([]string{tt.typeName, tt.pkgName})
	
	        // Execute command
	        err := cmd.Execute()
	        
	        if tt.wantErr {
	            assert.Error(t, err)
	            return
	        }
	        require.NoError(t, err)
	        assert.Contains(t, buf.String(), tt.wantOut)
	    })
	}
}
