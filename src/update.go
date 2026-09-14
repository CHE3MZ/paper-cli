package src

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// CLIVersion is the Paper CLI release version (e.g. "v1.0.0").
//
// It is baked in at build time via:
//
//	go build -ldflags "-X github.com/CHE3MZ/paper-cli/src.CLIVersion=vX.Y.Z" ./cmd/paper
//
// scripts/build.sh and scripts/build.ps1 do this automatically. Resolution
// order: $PAPER_CLI_VERSION when set (the release workflow sets it to the
// tag being released, CI sets it to "dev" so test artifacts never claim a
// release), otherwise the latest GitHub release tag
// (api.github.com/repos/CHE3MZ/paper-cli/releases/latest, i.e. what
// github.com/CHE3MZ/paper-cli/releases/latest redirects to), otherwise the
// latest local git tag, otherwise "dev". Local tags are only a fallback
// because they can be stale or unpushed. A dirty tree appends "-dirty", so
// a dev build can't masquerade as a release (and `paper update` won't
// wrongly call it up to date).
// Plain `go build` without ldflags leaves the "dev" default.
var CLIVersion = "dev"

// UpdateRepo is the GitHub repo self-update downloads from.
const UpdateRepo = "CHE3MZ/paper-cli"

// apiTimeout bounds release-API calls; downloadTimeout bounds the ~180MB
// binary download. The download limit is generous on purpose: slow
// connections need minutes, and a truly stalled one still fails instead of
// hanging forever (the stdlib default client has no timeout at all).
const apiTimeout = 30 * time.Second
const downloadTimeout = 15 * time.Minute

// defaultHTTPClient is used whenever callers pass a nil client.
func defaultHTTPClient() *http.Client {
	return &http.Client{Timeout: apiTimeout}
}

// downloadHTTPClient bounds the release binary download separately: the
// API timeout would abort slow-but-healthy downloads mid-stream.
func downloadHTTPClient() *http.Client {
	return &http.Client{Timeout: downloadTimeout}
}

// LatestReleaseAPI is the GitHub API endpoint that reports the latest
// release (i.e. what github.com/CHE3MZ/paper-cli/releases/latest points
// at). It is a var, not a const, so tests can point it at a local server.
var LatestReleaseAPI = "https://api.github.com/repos/" + UpdateRepo + "/releases/latest"

// LatestReleaseTag asks a LatestReleaseAPI-style endpoint for its tag_name
// (e.g. "v1.0.0").
func LatestReleaseTag(apiURL string, client *http.Client) (string, error) {
	info, err := FetchReleaseInfo(apiURL, client)
	if err != nil {
		return "", err
	}
	return info.Tag, nil
}

// ReleaseInfo is the slice of a GitHub latest-release response we care
// about: the tag plus the per-asset SHA256 digests for download verification.
type ReleaseInfo struct {
	Tag string
	// Digests maps asset name ("paper-linux", ...) to lowercase hex SHA256
	// (without any "sha256:" prefix). Assets without a usable digest are
	// simply absent.
	Digests map[string]string
}

// DigestFor returns the expected hex SHA256 for an asset, or "" when the
// release carries none (verification is then skipped, not failed).
func (r *ReleaseInfo) DigestFor(asset string) string {
	if r == nil {
		return ""
	}
	return r.Digests[asset]
}

// FetchReleaseInfo asks a LatestReleaseAPI-style endpoint for its tag_name
// and asset digests.
func FetchReleaseInfo(apiURL string, client *http.Client) (*ReleaseInfo, error) {
	if client == nil {
		client = defaultHTTPClient()
	}
	resp, err := client.Get(apiURL) //nolint:gosec,noctx // URL is the release API endpoint (or a test server)
	if err != nil {
		return nil, fmt.Errorf("query latest release at %s: %w", apiURL, err)
	}
	defer resp.Body.Close() //nolint:errcheck // decode error already surfaces below
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("query latest release at %s: server returned %s", apiURL, resp.Status)
	}
	var payload struct {
		TagName string `json:"tag_name"`
		Assets  []struct {
			Name   string `json:"name"`
			Digest string `json:"digest"`
		} `json:"assets"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("query latest release at %s: %w", apiURL, err)
	}
	if strings.TrimSpace(payload.TagName) == "" {
		return nil, fmt.Errorf("query latest release at %s: response has no tag_name", apiURL)
	}
	info := &ReleaseInfo{Tag: strings.TrimSpace(payload.TagName), Digests: map[string]string{}}
	for _, a := range payload.Assets {
		d, ok := strings.CutPrefix(strings.TrimSpace(a.Digest), "sha256:")
		if a.Name == "" || !ok || len(d) != 64 {
			continue
		}
		if _, err := hex.DecodeString(d); err != nil {
			continue
		}
		info.Digests[a.Name] = strings.ToLower(d)
	}
	return info, nil
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

// UpdateTempPath returns the staging file for a download:
// paper_temp.<pid> next to dest (paper_temp.<pid>.exe on Windows). The PID
// suffix keeps concurrent `paper update` processes from interleaving into
// one file; each process removes its own staging file on graceful failure
// paths, so only a violently killed download can orphan one.
func UpdateTempPath(dest string) string {
	dir := filepath.Dir(dest)
	name := fmt.Sprintf("paper_temp.%d", os.Getpid())
	if filepath.Ext(dest) == ".exe" {
		name += ".exe"
	}
	return filepath.Join(dir, name)
}

// DownloadFile fetches url into dest (partial downloads fail instead of
// leaving a truncated binary behind).
func DownloadFile(url, dest string, client *http.Client) error {
	if client == nil {
		client = downloadHTTPClient()
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

// VerifyFileSHA256 checks path against a lowercase hex SHA256 digest.
func VerifyFileSHA256(path, expectedHex string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("verify %s: %w", path, err)
	}
	defer f.Close() //nolint:errcheck // hash already finalized or copy already failed
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return fmt.Errorf("verify %s: %w", path, err)
	}
	if got := hex.EncodeToString(h.Sum(nil)); got != strings.ToLower(strings.TrimSpace(expectedHex)) {
		return fmt.Errorf("verify %s: checksum mismatch (not the published release binary?)", path)
	}
	return nil
}

// ApplyUpdate swaps the staged temp file over dest. The plain rename
// covers Unix (running binaries can be replaced) and any non-running file
// on Windows. A running Windows .exe cannot be overwritten in place, so
// there it is first renamed aside to dest+".old" — renaming a running exe
// only touches the directory entry, the running image stays mapped — and
// the staged file is moved into the freed name. The .old cannot be deleted
// while a process still runs from it, so the next update removes it first
// (best effort); at most one .old ever lingers.
func ApplyUpdate(dest, tmp string) error {
	return SwapStaged(tmp, dest)
}

// SwapStaged swaps the staged temp file over dest: one direct rename
// attempt, then the rename-aside fallback. There is deliberately no
// wait-and-retry loop: on Windows the helper itself runs from dest, so a
// direct rename can never succeed while it lives (retrying would just burn
// 30s before falling back anyway); the aside swap works against running
// images immediately, which the parent's exit is not even needed for.
func SwapStaged(tmp, dest string) error {
	if err := os.Chmod(tmp, 0o755); err != nil {
		return fmt.Errorf("chmod %s: %w", tmp, err)
	}
	if err := os.Rename(tmp, dest); err == nil {
		return nil
	}
	old := dest + ".old"
	_ = os.Remove(old) // leftover from a previous update, if any
	if err := os.Rename(dest, old); err != nil {
		return fmt.Errorf("replace %s (staged update kept at %s): %w", dest, tmp, err)
	}
	if err := os.Rename(tmp, dest); err != nil {
		_ = os.Rename(old, dest) // roll back: renaming back is allowed too
		return fmt.Errorf("replace %s (staged update kept at %s): %w", dest, tmp, err)
	}
	return nil
}

// spawnFinishHelper re-executes this binary as the hidden __finish-update
// command. The helper IS paper itself, so the install stays one file and
// there is no generated script to tamper with; detach() lets it outlive the
// parent process that spawned it.
func spawnFinishHelper(tmp, dest, digest string) (*exec.Cmd, error) {
	exe, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("resolve current executable: %w", err)
	}
	cmd := exec.Command(exe, "__finish-update", tmp, dest, digest)
	cmd.Stdin = nil
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	detach(cmd)
	return cmd, nil
}

// runHelperSync waits for a spawned helper and returns its exit code. The
// helper prints its own outcome to our (inherited) stdio, so there is
// nothing left to print here. A helper killed by a signal has no real exit
// code (Go reports -1); that maps to 1, not 255.
func runHelperSync(cmd *exec.Cmd) int {
	if err := cmd.Run(); err != nil {
		var exit *exec.ExitError
		if errors.As(err, &exit) && exit.ExitCode() >= 0 {
			return exit.ExitCode()
		}
		errLine(err)
		return 1
	}
	return 0
}

// runFinishUpdate implements the hidden `paper __finish-update` command: the
// second half of a self-update, run by a re-executed copy of this binary.
// It performs no version check (the parent already decided) and verifies the
// staged bytes before swapping them in.
func runFinishUpdate(args []string) int {
	if len(args) != 3 || args[0] == "" || args[1] == "" {
		fmt.Fprintln(os.Stderr, red("error:")+" usage: paper __finish-update TMP DEST SHA256")
		return 2
	}
	tmp, dest, digest := filepath.Clean(args[0]), filepath.Clean(args[1]), args[2]
	if !filepath.IsAbs(tmp) || !filepath.IsAbs(dest) || filepath.Dir(tmp) != filepath.Dir(dest) {
		fmt.Fprintln(os.Stderr, red("error:")+" refusing __finish-update: tmp and dest must be absolute paths in the same directory")
		return 2
	}
	if digest != "" {
		if err := VerifyFileSHA256(tmp, digest); err != nil {
			errLine(err)
			return 1
		}
	} else if st, err := os.Stat(tmp); err != nil || st.IsDir() || st.Size() == 0 {
		errLine(fmt.Errorf("refusing __finish-update: staged file missing or empty: %s", tmp))
		return 1
	}
	if _, err := os.Stat(dest); err != nil {
		errLine(fmt.Errorf("refusing __finish-update: destination missing: %s", dest))
		return 1
	}
	if err := SwapStaged(tmp, dest); err != nil {
		errLine(err)
		return 1
	}
	// Sweep a stale .old from the pre-helper era (or any earlier swap).
	// Best effort: when the fallback path just created one, this process
	// still runs from that image, so the remove fails and the note below
	// correctly reports it as kept.
	_ = os.Remove(dest + ".old")
	fmt.Printf("%s paper to %s\n%s\n", green("updated"), lightBlue(dest), gray("run `paper version` to confirm the new version."))
	if _, err := os.Stat(dest + ".old"); err == nil {
		fmt.Printf("%s\n", gray("old binary kept at "+dest+".old (removed on the next update)."))
	}
	return 0
}

// VersionText renders the `paper version` output.
func VersionText() string {
	return fmt.Sprintf("%s %s\n%s %s\n",
		bold(white("Paper MC Version:")), lightBlue(EmbeddedVersion),
		bold(white("Paper CLI Version:")), lightBlue(CLIVersion),
	)
}
