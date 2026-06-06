package source

import (
	"archive/zip"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/zryyyy/Nerd-Font-Patcher/internal/config"
)

func testConfig(root string) config.Config {
	return config.Config{
		CacheDir:  filepath.Join(root, ".cache"),
		WorkDir:   filepath.Join(root, "work"),
		OutputDir: filepath.Join(root, "dist"),
		Patcher: config.PatcherConfig{
			Image:    "nerdfonts/patcher",
			Args:     []string{"--complete"},
			Parallel: 10,
		},
	}
}

func writeZip(t *testing.T, zipPath string, entries map[string][]byte) {
	t.Helper()
	out, err := os.Create(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	defer out.Close()

	zw := zip.NewWriter(out)
	for name, content := range entries {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write(content); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
}

func basenames(paths []string) []string {
	names := make([]string, 0, len(paths))
	for _, p := range paths {
		names = append(names, filepath.Base(p))
	}
	sort.Strings(names)
	return names
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
