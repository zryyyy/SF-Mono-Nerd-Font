package source

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindNestedArchivesIncludesPayloadTilde(t *testing.T) {
	temp := t.TempDir()
	payloadTilde := filepath.Join(temp, "SFMonoFonts", "SF Mono Fonts.pkg", "SFMonoFonts.pkg", "Payload", "Payload~")
	if err := os.MkdirAll(filepath.Dir(payloadTilde), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(payloadTilde, []byte("payload archive"), 0o644); err != nil {
		t.Fatal(err)
	}

	archives, err := findNestedArchives(temp)
	if err != nil {
		t.Fatal(err)
	}
	if len(archives) != 1 || archives[0] != payloadTilde {
		t.Fatalf("unexpected archives: %v", archives)
	}
}

func TestCopyFontsFromTreeFindsSFMonoPayloadFonts(t *testing.T) {
	temp := t.TempDir()
	fontPath := filepath.Join(temp, "SFMonoFonts", "SF Mono Fonts.pkg", "SFMonoFonts.pkg", "Payload", "Payload~", "Library", "Fonts", "SF-Mono-Regular.otf")
	if err := os.MkdirAll(filepath.Dir(fontPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(fontPath, []byte("font"), 0o644); err != nil {
		t.Fatal(err)
	}

	dest := filepath.Join(temp, "out")
	copied, err := copyFontsFromTree(temp, dest, []string{"SF-Mono-*.otf"})
	if err != nil {
		t.Fatal(err)
	}
	if copied != 1 {
		t.Fatalf("expected 1 copied font, got %d", copied)
	}
	if _, err := os.Stat(filepath.Join(dest, "SF-Mono-Regular.otf")); err != nil {
		t.Fatalf("expected copied SF Mono font: %v", err)
	}
}
