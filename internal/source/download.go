package source

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"

	"github.com/zryyyy/Nerd-Font-Patcher/internal/config"
)

func downloadSource(cacheDir, groupName string, source config.FontSource) (string, error) {
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return "", fmt.Errorf("create cache dir: %w", err)
	}

	fileName, err := cacheFileName(groupName, source.URL)
	if err != nil {
		return "", err
	}
	dest := filepath.Join(cacheDir, fileName)
	if _, err := os.Stat(dest); err == nil {
		return dest, nil
	}

	fmt.Printf("Downloading %s\n", source.URL)
	resp, err := http.Get(source.URL)
	if err != nil {
		return "", fmt.Errorf("download %s: %w", source.URL, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("download %s: unexpected HTTP status %s", source.URL, resp.Status)
	}

	tmp := dest + ".tmp"
	out, err := os.Create(tmp)
	if err != nil {
		return "", fmt.Errorf("create cache file: %w", err)
	}
	if _, err := io.Copy(out, resp.Body); err != nil {
		out.Close()
		os.Remove(tmp)
		return "", fmt.Errorf("write cache file: %w", err)
	}
	if err := out.Close(); err != nil {
		os.Remove(tmp)
		return "", fmt.Errorf("close cache file: %w", err)
	}
	if err := os.Rename(tmp, dest); err != nil {
		os.Remove(tmp)
		return "", fmt.Errorf("store cache file: %w", err)
	}
	return dest, nil
}

func cacheFileName(groupName, rawURL string) (string, error) {
	base, err := sourceFileName(rawURL)
	if err != nil {
		return "", err
	}
	return groupName + "-" + base, nil
}

func sourceFileName(rawURL string) (string, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	base := path.Base(parsed.Path)
	if base == "." || base == "/" || base == "" {
		return "", fmt.Errorf("cannot infer file name from %q", rawURL)
	}
	return base, nil
}
