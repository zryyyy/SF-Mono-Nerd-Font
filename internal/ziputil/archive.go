package ziputil

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

type Entry struct {
	Source string
	Name   string
}

func ListFiles(root string) ([]string, error) {
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

func CreateFlat(zipPath string, files []string) error {
	entries := make([]Entry, 0, len(files))
	for _, file := range files {
		entries = append(entries, Entry{
			Source: file,
			Name:   filepath.Base(file),
		})
	}
	return Create(zipPath, entries)
}

func Create(zipPath string, entries []Entry) error {
	if len(entries) == 0 {
		return fmt.Errorf("no files to zip into %s", zipPath)
	}
	if err := os.MkdirAll(filepath.Dir(zipPath), 0o755); err != nil {
		return err
	}
	if err := os.Remove(zipPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	out, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer out.Close()

	zw := zip.NewWriter(out)
	seen := map[string]bool{}
	for _, entry := range entries {
		entryName, err := cleanEntryName(entry.Name)
		if err != nil {
			zw.Close()
			return err
		}
		if seen[entryName] {
			zw.Close()
			return fmt.Errorf("duplicate zip entry name %q", entryName)
		}
		seen[entryName] = true

		if err := addFileToZip(zw, entry.Source, entryName); err != nil {
			zw.Close()
			return err
		}
	}
	return zw.Close()
}

func cleanEntryName(name string) (string, error) {
	cleaned := path.Clean(filepath.ToSlash(name))
	if cleaned == "." || cleaned == "" || strings.HasPrefix(cleaned, "/") || strings.HasPrefix(cleaned, "../") || cleaned == ".." {
		return "", fmt.Errorf("unsafe zip entry name %q", name)
	}
	return cleaned, nil
}

func addFileToZip(zw *zip.Writer, filePath, entryName string) error {
	in, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer in.Close()

	info, err := in.Stat()
	if err != nil {
		return err
	}
	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	header.Name = entryName
	header.Method = zip.Deflate

	out, err := zw.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(out, in)
	return err
}
