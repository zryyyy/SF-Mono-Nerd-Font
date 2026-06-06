package app

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/zryyyy/Nerd-Font-Patcher/internal/config"
	"github.com/zryyyy/Nerd-Font-Patcher/internal/docker"
	"github.com/zryyyy/Nerd-Font-Patcher/internal/source"
	"github.com/zryyyy/Nerd-Font-Patcher/internal/ziputil"
)

func Run(configPath string, dryRun bool) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return err
	}
	if err := config.Validate(cfg); err != nil {
		return err
	}

	groups := config.EnabledGroups(cfg.Fonts)
	printPlan(cfg, groups)

	if err := docker.Check(); err != nil {
		return err
	}
	if config.UsesDMG(cfg) {
		if err := source.CheckDMGTools(); err != nil {
			return err
		}
	}
	if dryRun {
		return nil
	}

	if err := os.MkdirAll(cfg.CacheDir, 0o755); err != nil {
		return fmt.Errorf("create cache dir: %w", err)
	}
	if err := os.MkdirAll(cfg.WorkDir, 0o755); err != nil {
		return fmt.Errorf("create work dir: %w", err)
	}
	if err := os.MkdirAll(cfg.OutputDir, 0o755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}

	var allPatched []string
	for _, group := range groups {
		fmt.Printf("Preparing %s\n", group.Name)
		inputDir, outputDir, err := source.PrepareGroup(cfg, group)
		if err != nil {
			return err
		}

		fmt.Printf("Patching %s\n", group.Name)
		if err := docker.Run(cfg.Patcher, inputDir, outputDir); err != nil {
			return err
		}

		groupZip := filepath.Join(cfg.OutputDir, group.Name+".zip")
		files, err := ziputil.ListFiles(outputDir)
		if err != nil {
			return err
		}
		if len(files) == 0 {
			return fmt.Errorf("%s produced no patched fonts", group.Name)
		}
		if err := ziputil.CreateFlat(groupZip, files); err != nil {
			return err
		}
		allPatched = append(allPatched, files...)
		fmt.Printf("Wrote %s\n", groupZip)
	}

	if len(allPatched) == 0 {
		return errors.New("no patched fonts were produced")
	}
	allZip := filepath.Join(cfg.OutputDir, "all_fonts.zip")
	if err := ziputil.CreateFlat(allZip, allPatched); err != nil {
		return err
	}
	fmt.Printf("Wrote %s\n", allZip)
	return nil
}

func printPlan(cfg config.Config, groups []config.FontGroup) {
	fmt.Printf("Config: cache=%s work=%s output=%s\n", cfg.CacheDir, cfg.WorkDir, cfg.OutputDir)
	fmt.Printf("Patcher: image=%s args=%s parallel=%d\n", cfg.Patcher.Image, strings.Join(cfg.Patcher.Args, " "), cfg.Patcher.Parallel)
	for _, group := range groups {
		fmt.Printf("Font group: %s (%d sources)\n", group.Name, len(group.Sources))
	}
}
