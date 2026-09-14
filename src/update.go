package src

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// CLIVersion is the Paper CLI release version (e.g. "v1.0.0").
//
// It is baked in at build time via:
//
//	go build -ldflags "-X github.com/CHE3MZ/paper-cli/src.CLIVersion=vX.Y.Z" ./cmd/paper
//
// scripts/build.sh and scripts/build.ps1 do this automatically. Resolution
// order: $PAPER_CLI_VERSION when set (the release workflow sets it to the
// tag being released), otherwise the latest GitHub release tag
// (api.github.com/repos/CHE3MZ/paper-cli/releases/latest, i.e. what
// github.com/CHE3MZ/paper-cli/releases/latest redirects to), otherwise the
// latest local git tag, otherwise "dev". Local tags are only a fallback
// because they can be stale or unpushed.
// Plain `go build` without ldflags leaves the "dev" default.
var CLIVersion = "dev"

// UpdateRepo is the GitHub repo self-update downloads from.
const UpdateRepo = "CHE3MZ/paper-cli"

// LatestReleaseAPI is the GitHub API endpoint that reports the latest
// release (i.e. what github.com/CHE3MZ/paper-cli/releases/latest points
// at). It is a var, not a const, so tests can point it at a local server.
var LatestReleaseAPI = "https://api.github.com/repos/" + UpdateRepo + "/releases/latest"

// LatestReleaseTag asks a LatestReleaseAPI-style endpoint for its tag_name
// (e.g. "v1.0.0").
func LatestReleaseTag(apiURL string, client *http.Client) (string, error) {
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Get(apiURL) //nolint:gosec,noctx // URL is the release API endpoint (or a test server)
	if err != nil {
		return "", fmt.Errorf("query latest release at %s: %w", apiURL, err)
	}
	defer resp.Body.Close() //nolint:errcheck // decode error already surfaces below
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("query latest release at %s: server returned %s", apiURL, resp.Status)
	}
	var payload struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", fmt.Errorf("query latest release at %s: %w", apiURL, err)
	}
	if strings.TrimSpace(payload.TagName) == "" {
		return "", fmt.Errorf("query latest release at %s: response has no tag_name", apiURL)
	}
	return strings.TrimSpace(payload.TagName), nil
}

// SameCLIVersion reports whether current and latest name the same release.
// A single leading "v" is ignored, so "v1.0.0" and "1.0.0" match. "dev"
// never matches a real tag, so unstamped local builds always update.
func SameCLIVersion(current, latest string) bool {
	norm := func(v string) string {
		return strings.TrimPrefix(strings.TrimSpace(v), "v")
	}
	a, b := norm(current), norm(latest)
	return a != "" && b != "" && a == b
}

// UpdateAssetForGOOS maps a GOOS value to the release asset name,
// mirroring install/linux.sh, install/macos.sh and install/windows.bat.
func UpdateAssetForGOOS(goos string) string {
	switch goos {
	case "windows":
		return "paper-windows.exe"
	case "darwin":
		return "paper-macos"
	default:
		return "paper-linux"
	}
}

// UpdateDownloadURL builds the "latest release" download URL for an asset.
func UpdateDownloadURL(repo, asset string) string {
	return "https://github.com/" + repo + "/releases/latest/download/" + asset
}

// DefaultInstallPath mirrors the install scripts: ~/.local/bin/paper
// (paper.exe on Windows).
func DefaultInstallPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return "", fmt.Errorf("resolve home directory for install path: %w", err)
	}
	name := "paper"
	if runtime.GOOS == "windows" {
		name = "paper.exe"
	}
	return filepath.Join(home, ".local", "bin", name), nil
}

// CurrentExecutablePath returns the resolved path of the running binary.
func CurrentExecutablePath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("resolve current executable: %w", err)
	}
	if resolved, err := filepath.EvalSymlinks(exe); err == nil {
		exe = resolved
	}
	return exe, nil
}

// UpdateTarget returns the file `paper update` replaces: the running
// binary when it can be resolved, otherwise the default install path.
func UpdateTarget() (string, error) {
	if exe, err := CurrentExecutablePath(); err == nil && exe != "" {
		return exe, nil
	}
	return DefaultInstallPath()
}

// UpdateTempPath returns the staging file for a download: paper_temp
// next to dest (paper_temp.exe on Windows).
func UpdateTempPath(dest string) string {
	dir := filepath.Dir(dest)
	name := "paper_temp"
	if filepath.Ext(dest) == ".exe" {
		name = "paper_temp.exe"
	}
	return filepath.Join(dir, name)
}

// DownloadFile fetches url into dest (partial downloads fail instead of
// leaving a truncated binary behind).
func DownloadFile(url, dest string, client *http.Client) error {
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Get(url) //nolint:gosec,noctx // URL is the release asset URL shown to the user
	if err != nil {
		return fmt.Errorf("download %s: %w", url, err)
	}
	defer resp.Body.Close() //nolint:errcheck // read error already surfaces via io.Copy
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download %s: server returned %s", url, resp.Status)
	}
	out, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return fmt.Errorf("write %s: %w", dest, err)
	}
	if _, err := io.Copy(out, resp.Body); err != nil {
		_ = out.Close()
		return fmt.Errorf("write %s: %w", dest, err)
	}
	if err := out.Close(); err != nil {
		return fmt.Errorf("write %s: %w", dest, err)
	}
	// Best effort on Windows (chmod is mostly a no-op there, the open
	// mode above already set what matters).
	_ = os.Chmod(dest, 0o755)
	return nil
}

// ApplyUpdate atomically swaps the staged temp file over dest.
func ApplyUpdate(dest, tmp string) error {
	if err := os.Chmod(tmp, 0o755); err != nil {
		return fmt.Errorf("chmod %s: %w", tmp, err)
	}
	if err := os.Rename(tmp, dest); err != nil {
		return fmt.Errorf("replace %s (staged update kept at %s): %w", dest, tmp, err)
	}
	return nil
}

// SelfUpdate downloads url into a paper_temp file next to dest, then
// replaces dest with it. It mirrors install/*.sh + install/windows.bat
// (same repo, same latest-download URL, same ~/.local/bin destination
// family) but runs from inside the CLI.
func SelfUpdate(dest, url string, client *http.Client) (string, error) {
	tmp := UpdateTempPath(dest)
	if err := DownloadFile(url, tmp, client); err != nil {
		return "", err
	}
	if err := ApplyUpdate(dest, tmp); err != nil {
		return tmp, err
	}
	return dest, nil
}

// VersionText renders the `paper version` output.
func VersionText() string {
	return fmt.Sprintf("%s %s\n%s %s\n",
		bold(white("Paper MC Version:")), lightBlue(EmbeddedVersion),
		bold(white("Paper CLI Version:")), lightBlue(CLIVersion),
	)
}
