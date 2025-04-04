package grit_test

import (
	_ "embed"
	"log"
	"os"
	"testing"

	"github.com/weslien/grit/pkg/grit"
	"gopkg.in/yaml.v3"
)

//go:embed test_assets/grit_valid.yaml
var gritValidYaml string

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name  string
		yaml  string
		want  *grit.RootConfig
		error bool
	}{
		{
			name: "empty file returns defaults",
			yaml: "",
			want: &grit.RootConfig{
				Types:   make(map[string]grit.TypeConfig),
				Targets: nil,
				Repo:    grit.RepoConfig{},
			},
		},
		{
			name: "valid config",
			yaml: gritValidYaml,
			want: &grit.RootConfig{
				Repo: grit.RepoConfig{
					Name:    "default",
					URL:     "",
					License: "",
					Owner:   "",
				},
				Targets: map[string]string{
					"build":   "echo build",
					"test":    "echo test",
					"lint":    "echo lint",
					"release": "echo release",
				},
				Types: map[string]grit.TypeConfig{
					"lib": {
						PackageDir:  "packages/lib",
						BuildDir:    "",
						CoverageDir: "",
						Targets: map[string]string{
							"build": "echo lib-build",
							"test":  "echo lib-test",
						},
						CanDependOn: nil,
					},
				},
			},
		},
		{
			name:  "invalid yaml",
			yaml:  "types: !!invalid",
			want:  nil,
			error: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmp := t.TempDir()
			path := tmp + "/grit.yaml"
			if tt.yaml != "" {
				log.Println(tt.yaml)
				err := os.WriteFile(path, []byte(tt.yaml), 0644)
				if err != nil {
					t.Fatal(err)
				}
			}

			got, err := grit.LoadConfig(path)
			if (err != nil) != tt.error {
				t.Fatalf("LoadConfig() error = %v, wantErr %v", err, tt.error)
			}

			if tt.error {
				return
			}

			wantYaml, _ := yaml.Marshal(tt.want)
			gotYaml, _ := yaml.Marshal(got)
			if string(gotYaml) != string(wantYaml) {
				t.Errorf("LoadConfig() = %s, want %s", gotYaml, wantYaml)
			}
		})
	}
}

func TestSaveConfig(t *testing.T) {
	cfg := &grit.RootConfig{
		Types: map[string]grit.TypeConfig{
			"app": {PackageDir: "packages/app"},
		},
	}

	tmp := t.TempDir()
	path := tmp + "/grit.yaml"

	err := grit.SaveConfig(cfg, path)
	if err != nil {
		t.Fatalf("SaveConfig() error = %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	var loaded grit.RootConfig
	err = yaml.Unmarshal(data, &loaded)
	if err != nil {
		t.Fatal(err)
	}

	wantYaml, _ := yaml.Marshal(cfg)
	gotYaml, _ := yaml.Marshal(&loaded)
	if string(gotYaml) != string(wantYaml) {
		t.Errorf("SaveConfig() saved = %s, want %s", gotYaml, wantYaml)
	}
}

func TestMergeDefaults(t *testing.T) {
	defaults := grit.TypeConfig{
		PackageDir: "packages/lib",
		Targets:    map[string]string{"build": "build", "test": "test"},
	}

	t.Run("adds missing lib type", func(t *testing.T) {
		cfg := &grit.RootConfig{Types: make(map[string]grit.TypeConfig)}
		cfg.MergeDefaults(defaults)

		if _, exists := cfg.Types["lib"]; !exists {
			t.Error("MergeDefaults() did not add lib type")
		}
	})

	t.Run("preserves existing lib type", func(t *testing.T) {
		original := grit.TypeConfig{PackageDir: "custom/lib"}
		cfg := &grit.RootConfig{
			Types: map[string]grit.TypeConfig{"lib": original},
		}
		cfg.MergeDefaults(defaults)

		if cfg.Types["lib"].PackageDir != "custom/lib" {
			t.Error("MergeDefaults() overwrote existing lib type")
		}
	})
}

func TestLoadMissingFile(t *testing.T) {
	cfg, err := grit.LoadConfig("/nonexistent")
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if cfg == nil || cfg.Types == nil {
		t.Error("LoadConfig() should return initialized config for missing file")
	}

}
