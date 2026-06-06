package source

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

type commandChecker func(string) (string, error)

func CheckDMGTools() error {
	if _, err := find7z(exec.LookPath); err != nil {
		return errors.New("required command not found for dmg source: 7z or 7zz")
	}
	return nil
}

func find7z(lookPath commandChecker) (string, error) {
	for _, name := range []string{"7z", "7zz"} {
		if path, err := lookPath(name); err == nil {
			return path, nil
		}
	}
	return "", errors.New("7z not found")
}

func extractDMGFonts(dmgPath, inputDir, extractRoot string, include []string) error {
	sevenZip, err := find7z(exec.LookPath)
	if err != nil {
		return err
	}

	queue := []string{dmgPath}
	seen := map[string]bool{}
	for len(queue) > 0 {
		archive := queue[0]
		queue = queue[1:]
		if seen[archive] {
			continue
		}
		seen[archive] = true

		outDir := filepath.Join(extractRoot, fmt.Sprintf("archive-%d", len(seen)))
		if err := os.MkdirAll(outDir, 0o755); err != nil {
			return fmt.Errorf("create dmg extract dir: %w", err)
		}
		if err := run7z(sevenZip, archive, outDir); err != nil {
			return err
		}

		nested, err := findNestedArchives(outDir)
		if err != nil {
			return err
		}
		queue = append(queue, nested...)
	}

	copied, err := copyFontsFromTree(extractRoot, inputDir, include)
	if err != nil {
		return err
	}
	if copied == 0 {
		if len(include) > 0 {
			return fmt.Errorf("dmg extraction produced no matching font files for include patterns: %q", include)
		}
		return errors.New("dmg extraction produced no matching font files")
	}
	return nil
}

func findNestedArchives(root string) ([]string, error) {
	var archives []string
	err := filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		name := strings.ToLower(d.Name())
		ext := strings.ToLower(filepath.Ext(name))
		if isNestedArchiveName(name, ext) {
			archives = append(archives, p)
		}
		return nil
	})
	sort.Strings(archives)
	return archives, err
}

func isNestedArchiveName(name, ext string) bool {
	if name == "payload" || strings.HasPrefix(name, "payload~") {
		return true
	}
	switch ext {
	case ".dmg", ".pkg", ".xar", ".zip", ".gz", ".tgz", ".tar", ".cpio":
		return true
	default:
		return false
	}
}

func copyFontsFromTree(root, destDir string, include []string) (int, error) {
	count := 0
	err := filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !isFontFile(p) || !matchesInclude(filepath.Base(p), include) {
			return nil
		}
		dest := filepath.Join(destDir, filepath.Base(p))
		if err := copyFile(p, dest); err != nil {
			return err
		}
		count++
		return nil
	})
	return count, err
}

func run7z(sevenZip, archive, outDir string) error {
	cmd := exec.Command(sevenZip, "x", "-y", "-o"+outDir, archive)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("extract %s with 7z: %w", archive, err)
	}
	return nil
}
