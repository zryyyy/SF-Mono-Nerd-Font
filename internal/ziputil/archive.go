package ziputil

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
)

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
	if len(files) == 0 {
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
	for _, file := range files {
		entryName := filepath.Base(file)
		if seen[entryName] {
			zw.Close()
			return fmt.Errorf("duplicate zip entry name %q", entryName)
		}
		seen[entryName] = true

		if err := addFileToZip(zw, file, entryName); err != nil {
			zw.Close()
			return err
		}
	}
	return zw.Close()
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
