package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestValidateConfigRejectsUnsafeGroupName(t *testing.T) {
	cfg := Config{
		CacheDir:  ".cache",
		WorkDir:   "work",
		OutputDir: "dist",
		Patcher: PatcherConfig{
			Image:    "nerdfonts/patcher",
			Args:     []string{"--complete"},
			Parallel: 10,
		},
		Fonts: []FontGroup{
			{
				Name: "../bad",
				Sources: []FontSource{
					{Type: SourceFont, URL: "https://example.com/a.ttf"},
				},
			},
		},
	}

	if err := Validate(cfg); err == nil {
		t.Fatal("expected invalid group name error")
	}
}

func TestEnabledGroupsSkipsDisabled(t *testing.T) {
	groups := EnabledGroups([]FontGroup{
		{Name: "on"},
		{Name: "off", Enabled: new(false)},
	})

	if len(groups) != 1 || groups[0].Name != "on" {
		t.Fatalf("unexpected enabled groups: %#v", groups)
	}
}

func TestLoadConfigAppliesDefaults(t *testing.T) {
	temp := t.TempDir()
	configPath := filepath.Join(temp, "font.json")
	raw := map[string]any{
		"fonts": []map[string]any{
			{
				"name": "Example",
				"sources": []map[string]string{
					{"type": SourceFont, "url": "https://example.com/font.ttf"},
				},
			},
		},
	}
	data, err := json.Marshal(raw)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(configPath, data, 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.CacheDir != ".cache" || cfg.WorkDir != "work" || cfg.OutputDir != "dist" {
		t.Fatalf("defaults were not applied: %#v", cfg)
	}
	if cfg.Patcher.Image != "nerdfonts/patcher" || cfg.Patcher.Parallel != 10 {
		t.Fatalf("patcher defaults were not applied: %#v", cfg.Patcher)
	}
}
