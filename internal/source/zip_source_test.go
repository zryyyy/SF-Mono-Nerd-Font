package source

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExtractZipFontsFiltersFonts(t *testing.T) {
	temp := t.TempDir()
	zipPath := filepath.Join(temp, "fonts.zip")
	writeZip(t, zipPath, map[string][]byte{
		"keep.ttf":        []byte("ttf"),
		"nested/keep.otf": []byte("otf"),
		"skip.txt":        []byte("txt"),
	})

	dest := filepath.Join(temp, "out")
	if err := os.MkdirAll(dest, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := extractZipFonts(zipPath, dest, nil); err != nil {
		t.Fatalf("extractZipFonts failed: %v", err)
	}

	files, err := listFiles(dest)
	if err != nil {
		t.Fatal(err)
	}
	names := basenames(files)
	if want := []string{"keep.otf", "keep.ttf"}; !equalStrings(names, want) {
		t.Fatalf("unexpected extracted files: got %v want %v", names, want)
	}
}

func TestExtractZipFontsRejectsZipSlip(t *testing.T) {
	temp := t.TempDir()
	zipPath := filepath.Join(temp, "bad.zip")
	writeZip(t, zipPath, map[string][]byte{
		"../evil.ttf": []byte("bad"),
	})

	dest := filepath.Join(temp, "out")
	if err := os.MkdirAll(dest, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := extractZipFonts(zipPath, dest, nil); err == nil {
		t.Fatal("expected zip slip error")
	}
	if _, err := os.Stat(filepath.Join(temp, "evil.ttf")); !os.IsNotExist(err) {
		t.Fatalf("zip slip wrote outside destination: %v", err)
	}
}
