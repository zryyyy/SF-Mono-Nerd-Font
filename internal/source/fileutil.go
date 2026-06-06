package source

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

func isFontFile(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	return ext == ".ttf" || ext == ".otf"
}

func matchesInclude(name string, include []string) bool {
	if len(include) == 0 {
		return true
	}
	base := path.Base(filepath.ToSlash(name))
	slashed := filepath.ToSlash(name)
	for _, pattern := range include {
		if pattern == base || pattern == slashed {
			return true
		}
		if ok, _ := path.Match(pattern, base); ok {
			return true
		}
		if ok, _ := path.Match(pattern, slashed); ok {
			return true
		}
	}
	return false
}

func safeJoin(base, name string) (string, error) {
	slashed := strings.ReplaceAll(name, "\\", "/")
	if strings.Contains(slashed, "\x00") || strings.HasPrefix(slashed, "/") {
		return "", fmt.Errorf("unsafe archive path: %s", name)
	}
	if filepath.IsAbs(filepath.FromSlash(slashed)) {
		return "", fmt.Errorf("unsafe archive path: %s", name)
	}
	for _, part := range strings.Split(slashed, "/") {
		if part == ".." || strings.Contains(part, ":") {
			return "", fmt.Errorf("unsafe archive path: %s", name)
		}
	}

	target := filepath.Join(base, filepath.FromSlash(slashed))
	rel, err := filepath.Rel(base, target)
	if err != nil {
		return "", err
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("unsafe archive path: %s", name)
	}
	return target, nil
}

func copyFile(src, dest string) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}

func listFiles(root string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			files = append(files, p)
		}
		return nil
	})
	sort.Strings(files)
	return files, err
}
