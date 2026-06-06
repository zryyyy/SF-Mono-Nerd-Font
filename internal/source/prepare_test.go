package source

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/zryyyy/Nerd-Font-Patcher/internal/config"
)

func TestPrepareFontGroupDownloadsDirectFont(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("fake font"))
	}))
	defer server.Close()

	temp := t.TempDir()
	cfg := testConfig(temp)
	group := config.FontGroup{
		Name: "Direct",
		Sources: []config.FontSource{
			{Type: config.SourceFont, URL: server.URL + "/Direct-Regular.ttf"},
		},
	}

	inputDir, _, err := PrepareGroup(cfg, group)
	if err != nil {
		t.Fatalf("PrepareGroup failed: %v", err)
	}

	got := filepath.Join(inputDir, "Direct-Regular.ttf")
	if _, err := os.Stat(got); err != nil {
		t.Fatalf("expected downloaded font at %s: %v", got, err)
	}
}
