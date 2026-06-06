package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"
)

const (
	SourceFont = "font"
	SourceZip  = "zip"
	SourceDMG  = "dmg"
)

var safeNamePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*$`)

type Config struct {
	CacheDir  string        `json:"cacheDir"`
	WorkDir   string        `json:"workDir"`
	OutputDir string        `json:"outputDir"`
	Patcher   PatcherConfig `json:"patcher"`
	Fonts     []FontGroup   `json:"fonts"`
}

type PatcherConfig struct {
	Image    string   `json:"image"`
	Args     []string `json:"args"`
	Parallel int      `json:"parallel"`
}

type FontGroup struct {
	Name    string       `json:"name"`
	Enabled *bool        `json:"enabled,omitempty"`
	Sources []FontSource `json:"sources"`
	Include []string     `json:"include,omitempty"`
}

type FontSource struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

func Load(configPath string) (Config, error) {
	f, err := os.Open(configPath)
	if err != nil {
		return Config{}, fmt.Errorf("open config: %w", err)
	}
	defer f.Close()

	var cfg Config
	decoder := json.NewDecoder(f)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&cfg); err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}

	applyConfigDefaults(&cfg)
	return cfg, nil
}

func applyConfigDefaults(cfg *Config) {
	if cfg.CacheDir == "" {
		cfg.CacheDir = ".cache"
	}
	if cfg.WorkDir == "" {
		cfg.WorkDir = "work"
	}
	if cfg.OutputDir == "" {
		cfg.OutputDir = "dist"
	}
	if cfg.Patcher.Image == "" {
		cfg.Patcher.Image = "nerdfonts/patcher"
	}
	if len(cfg.Patcher.Args) == 0 {
		cfg.Patcher.Args = []string{"--complete"}
	}
	if cfg.Patcher.Parallel <= 0 {
		cfg.Patcher.Parallel = 10
	}
}

func Validate(cfg Config) error {
	if cfg.CacheDir == "" || cfg.WorkDir == "" || cfg.OutputDir == "" {
		return errors.New("cacheDir, workDir, and outputDir must not be empty")
	}
	if cfg.Patcher.Image == "" {
		return errors.New("patcher.image must not be empty")
	}
	if len(cfg.Fonts) == 0 {
		return errors.New("fonts must not be empty")
	}

	seen := map[string]bool{}
	for _, group := range cfg.Fonts {
		if !safeNamePattern.MatchString(group.Name) {
			return fmt.Errorf("invalid font group name %q", group.Name)
		}
		if seen[group.Name] {
			return fmt.Errorf("duplicate font group name %q", group.Name)
		}
		seen[group.Name] = true

		if !IsEnabled(group) {
			continue
		}
		if len(group.Sources) == 0 {
			return fmt.Errorf("%s has no sources", group.Name)
		}
		for _, include := range group.Include {
			if _, err := path.Match(include, "ignored"); err != nil {
				return fmt.Errorf("%s has invalid include pattern %q: %w", group.Name, include, err)
			}
		}
		for _, source := range group.Sources {
			if err := validateSource(group.Name, source); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateSource(groupName string, source FontSource) error {
	switch source.Type {
	case SourceFont, SourceZip, SourceDMG:
	default:
		return fmt.Errorf("%s has unsupported source type %q", groupName, source.Type)
	}
	parsed, err := url.Parse(source.URL)
	if err != nil {
		return fmt.Errorf("%s has invalid source URL %q: %w", groupName, source.URL, err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("%s source URL must use http or https: %q", groupName, source.URL)
	}

	ext := strings.ToLower(path.Ext(parsed.Path))
	switch source.Type {
	case SourceFont:
		if ext != ".ttf" && ext != ".otf" {
			return fmt.Errorf("%s font source must end with .ttf or .otf: %q", groupName, source.URL)
		}
	case SourceZip:
		if ext != ".zip" {
			return fmt.Errorf("%s zip source must end with .zip: %q", groupName, source.URL)
		}
	case SourceDMG:
		if ext != ".dmg" {
			return fmt.Errorf("%s dmg source must end with .dmg: %q", groupName, source.URL)
		}
	}
	return nil
}

func EnabledGroups(groups []FontGroup) []FontGroup {
	var enabled []FontGroup
	for _, group := range groups {
		if IsEnabled(group) {
			enabled = append(enabled, group)
		}
	}
	return enabled
}

func IsEnabled(group FontGroup) bool {
	return group.Enabled == nil || *group.Enabled
}

func UsesDMG(cfg Config) bool {
	for _, group := range EnabledGroups(cfg.Fonts) {
		for _, source := range group.Sources {
			if source.Type == SourceDMG {
				return true
			}
		}
	}
	return false
}
