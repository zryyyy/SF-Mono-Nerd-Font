package ziputil

import (
	"archive/zip"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"testing"
)

func TestCreateFlatZipFileList(t *testing.T) {
	temp := t.TempDir()
	a := filepath.Join(temp, "A.ttf")
	b := filepath.Join(temp, "B.otf")
	if err := os.WriteFile(a, []byte("a"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(b, []byte("b"), 0o644); err != nil {
		t.Fatal(err)
	}

	zipPath := filepath.Join(temp, "out", "fonts.zip")
	if err := CreateFlat(zipPath, []string{a, b}); err != nil {
		t.Fatalf("CreateFlat failed: %v", err)
	}

	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()

	var names []string
	for _, file := range reader.File {
		names = append(names, file.Name)
	}
	sort.Strings(names)
	if want := []string{"A.ttf", "B.otf"}; !slices.Equal(names, want) {
		t.Fatalf("unexpected zip entries: got %v want %v", names, want)
	}
}
