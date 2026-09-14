package test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/CHE3MZ/paper-cli/src"
)

func TestVersionTextShowsBothVersions(t *testing.T) {
	text := src.VersionText()
	for _, want := range []string{"Paper MC Version:", "Paper CLI Version:", src.EmbeddedVersion, src.CLIVersion} {
		if !strings.Contains(text, want) {
			t.Errorf("VersionText() missing %q, got:\n%s", want, text)
		}
	}
}

func TestHelpTextCoversVersionAndUpdate(t *testing.T) {
	help := src.HelpText()
	for _, want := range []string{"paper version", "paper update"} {
		if !strings.Contains(help, want) {
			t.Errorf("help text missing %q", want)
		}
	}
}

func TestVersionCommandFlags(t *testing.T) {
	if got := src.Run([]string{"paper", "version", "--bogus"}); got != 2 {
		t.Errorf("paper version --bogus = %d, want 2", got)
	}
	if got := src.Run([]string{"paper", "version", "extra"}); got != 2 {
		t.Errorf("paper version extra = %d, want 2", got)
	}
	if got := src.Run([]string{"paper", "version", "--help"}); got != 0 {
		t.Errorf("paper version --help = %d, want 0", got)
	}
	if got := src.Run([]string{"paper", "version"}); got != 0 {
		t.Errorf("paper version = %d, want 0", got)
	}
}

func TestUpdateCommandFlags(t *testing.T) {
	if got := src.Run([]string{"paper", "update", "--bogus"}); got != 2 {
		t.Errorf("paper update --bogus = %d, want 2", got)
	}
	if got := src.Run([]string{"paper", "update", "extra"}); got != 2 {
		t.Errorf("paper update extra = %d, want 2", got)
	}
	if got := src.Run([]string{"paper", "update", "--help"}); got != 0 {
		t.Errorf("paper update --help = %d, want 0", got)
	}
}

func TestUpdateAssetMapping(t *testing.T) {
	cases := map[string]string{
		"windows": "paper-windows.exe",
		"darwin":  "paper-macos",
		"linux":   "paper-linux",
	}
	for goos, want := range cases {
		if got := src.UpdateAssetForGOOS(goos); got != want {
			t.Errorf("UpdateAssetForGOOS(%q) = %q, want %q", goos, got, want)
		}
	}
	if got := src.UpdateDownloadURL("CHE3MZ/paper-cli", "paper-linux"); got != "https://github.com/CHE3MZ/paper-cli/releases/latest/download/paper-linux" {
		t.Errorf("UpdateDownloadURL = %q", got)
	}
}

func TestSelfUpdateReplacesBinary(t *testing.T) {
	newBytes := []byte("new-paper-binary")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(newBytes)
	}))
	defer srv.Close()

	dir := t.TempDir()
	dest := filepath.Join(dir, "paper")
	if err := os.WriteFile(dest, []byte("old-paper-binary"), 0o755); err != nil {
		t.Fatal(err)
	}

	if _, err := src.SelfUpdate(dest, srv.URL+"/paper-linux", nil); err != nil {
		t.Fatalf("SelfUpdate: %v", err)
	}
	got, err := os.ReadFile(dest)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(newBytes) {
		t.Errorf("dest = %q, want %q", got, newBytes)
	}
	// Staging file must be gone (renamed over dest).
	if _, err := os.Stat(src.UpdateTempPath(dest)); !os.IsNotExist(err) {
		t.Errorf("expected staging file to be consumed, stat err = %v", err)
	}
}

func TestSelfUpdateFailsCleanlyOnBadStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "no release", http.StatusNotFound)
	}))
	defer srv.Close()

	dir := t.TempDir()
	dest := filepath.Join(dir, "paper")
	if err := os.WriteFile(dest, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := src.SelfUpdate(dest, srv.URL+"/paper-linux", nil); err == nil {
		t.Fatal("expected error on 404, got nil")
	}
	if got, _ := os.ReadFile(dest); string(got) != "old" {
		t.Errorf("failed update must leave dest alone, got %q", got)
	}
}
