package source

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/zryyyy/Nerd-Font-Patcher/internal/config"
)

func PrepareGroup(cfg config.Config, group config.FontGroup) (string, string, error) {
	inputDir := filepath.Join(cfg.WorkDir, "input", group.Name)
	outputDir := filepath.Join(cfg.WorkDir, "patched", group.Name)
	extractDir := filepath.Join(cfg.WorkDir, "extract", group.Name)

	for _, dir := range []string{inputDir, outputDir, extractDir} {
		if err := os.RemoveAll(dir); err != nil {
			return "", "", fmt.Errorf("clean %s: %w", dir, err)
		}
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return "", "", fmt.Errorf("create %s: %w", dir, err)
		}
	}

	for i, source := range group.Sources {
		downloaded, err := downloadSource(cfg.CacheDir, group.Name, source)
		if err != nil {
			return "", "", err
		}
		switch source.Type {
		case config.SourceFont:
			sourceName, err := sourceFileName(source.URL)
			if err != nil {
				return "", "", err
			}
			if err := copyFontFile(downloaded, inputDir, sourceName, group.Include); err != nil {
				return "", "", fmt.Errorf("%s source %d: %w", group.Name, i+1, err)
			}
		case config.SourceZip:
			if err := extractZipFonts(downloaded, inputDir, group.Include); err != nil {
				return "", "", fmt.Errorf("%s source %d: %w", group.Name, i+1, err)
			}
		case config.SourceDMG:
			if err := extractDMGFonts(downloaded, inputDir, extractDir, group.Include); err != nil {
				return "", "", fmt.Errorf("%s source %d: %w", group.Name, i+1, err)
			}
		}
	}

	files, err := listFiles(inputDir)
	if err != nil {
		return "", "", err
	}
	if len(files) == 0 {
		return "", "", fmt.Errorf("%s has no input font files after extraction", group.Name)
	}
	return inputDir, outputDir, nil
}

func copyFontFile(src, destDir, fileName string, include []string) error {
	if !isFontFile(src) {
		return fmt.Errorf("not a supported font file: %s", src)
	}
	if !matchesInclude(fileName, include) {
		return nil
	}
	dest := filepath.Join(destDir, fileName)
	return copyFile(src, dest)
}
