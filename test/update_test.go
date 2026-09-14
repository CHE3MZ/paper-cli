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

func TestSameCLIVersion(t *testing.T) {
	cases := []struct {
		current, latest string
		want            bool
	}{
		{"v1.0.0", "v1.0.0", true},
		{"v1.0.0", "1.0.0", true}, // leading "v" is ignored
		{"1.0.0", "v1.0.0", true},
		{"v1.0.0", "v1.0.1", false},
		{"dev", "v1.0.0", false}, // unstamped builds always update
		{"", "v1.0.0", false},
		{"v1.0.0", "", false},
	}
	for _, c := range cases {
		if got := src.SameCLIVersion(c.current, c.latest); got != c.want {
			t.Errorf("SameCLIVersion(%q, %q) = %v, want %v", c.current, c.latest, got, c.want)
		}
	}
}

func TestLatestReleaseTag(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"tag_name": "v1.2.3"}`))
	}))
	defer srv.Close()
	if got, err := src.LatestReleaseTag(srv.URL, nil); err != nil || got != "v1.2.3" {
		t.Errorf("LatestReleaseTag = %q, %v; want %q, nil", got, err, "v1.2.3")
	}

	bad := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "rate limited", http.StatusForbidden)
	}))
	defer bad.Close()
	if _, err := src.LatestReleaseTag(bad.URL, nil); err == nil {
		t.Error("LatestReleaseTag on 403 expected error, got nil")
	}
}

func TestUpdateSkipsDownloadWhenCurrent(t *testing.T) {
	// The fake API reports exactly the version this test process claims
	// to be, so `paper update` must take the early "already up to date"
	// exit without touching any binary on disk.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"tag_name": "v9.9.9-test"}`))
	}))
	defer srv.Close()

	oldVersion, oldAPI := src.CLIVersion, src.LatestReleaseAPI
	src.CLIVersion, src.LatestReleaseAPI = "v9.9.9-test", srv.URL
	defer func() { src.CLIVersion, src.LatestReleaseAPI = oldVersion, oldAPI }()

	if got := src.Run([]string{"paper", "update"}); got != 0 {
		t.Errorf("paper update when current = %d, want 0", got)
	}
}
