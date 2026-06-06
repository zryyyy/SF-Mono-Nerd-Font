package source

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
)

func extractZipFonts(zipPath, destDir string, include []string) error {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("open zip: %w", err)
	}
	defer reader.Close()

	seen := map[string]bool{}
	for _, file := range reader.File {
		if file.FileInfo().IsDir() || !isFontFile(file.Name) || !matchesInclude(file.Name, include) {
			continue
		}
		if _, err := safeJoin(destDir, file.Name); err != nil {
			return err
		}
		base := path.Base(filepath.ToSlash(file.Name))
		if seen[base] {
			return fmt.Errorf("duplicate font file name in zip: %s", base)
		}
		seen[base] = true
		dest, err := safeJoin(destDir, base)
		if err != nil {
			return err
		}
		if err := extractZipFile(file, dest); err != nil {
			return err
		}
	}
	return nil
}

func extractZipFile(file *zip.File, dest string) error {
	in, err := file.Open()
	if err != nil {
		return fmt.Errorf("open zip entry %s: %w", file.Name, err)
	}
	defer in.Close()

	out, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("create %s: %w", dest, err)
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return fmt.Errorf("extract %s: %w", file.Name, err)
	}
	if err := out.Close(); err != nil {
		return fmt.Errorf("close %s: %w", dest, err)
	}
	return nil
}
