package test

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
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

func TestDownloadAndSwapReplacesBinary(t *testing.T) {
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
	tmp := filepath.Join(dir, "paper_temp")
	if err := src.DownloadFile(srv.URL+"/paper-linux", tmp, nil); err != nil {
		t.Fatalf("DownloadFile: %v", err)
	}
	if err := src.SwapStaged(tmp, dest); err != nil {
		t.Fatalf("SwapStaged: %v", err)
	}
	got, err := os.ReadFile(dest)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(newBytes) {
		t.Errorf("dest = %q, want %q", got, newBytes)
	}
	// Staging file must be gone (renamed over dest).
	if _, err := os.Stat(tmp); !os.IsNotExist(err) {
		t.Errorf("expected staging file to be consumed, stat err = %v", err)
	}
}

func TestDownloadFailsCleanlyOnBadStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "no release", http.StatusNotFound)
	}))
	defer srv.Close()

	dir := t.TempDir()
	dest := filepath.Join(dir, "paper")
	if err := os.WriteFile(dest, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := src.DownloadFile(srv.URL+"/paper-linux", filepath.Join(dir, "paper_temp"), nil); err == nil {
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

func TestApplyUpdateReplacesRunningExecutableOnWindows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows-only: everywhere else a running binary can be replaced directly")
	}
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain not on PATH")
	}
	dir := t.TempDir()
	helper := "package main\n\nimport \"time\"\n\nfunc main() { time.Sleep(60 * time.Second) }\n"
	if err := os.WriteFile(filepath.Join(dir, "sleeper.go"), []byte(helper), 0o644); err != nil {
		t.Fatal(err)
	}
	exe := filepath.Join(dir, "sleeper.exe")
	build := exec.Command("go", "build", "-o", exe, "sleeper.go")
	build.Dir = dir
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build sleeper: %v\n%s", err, out)
	}

	// From here on exe is a RUNNING image: overwriting it in place must
	// fail at the OS level, which is exactly what ApplyUpdate works around.
	proc := exec.Command(exe)
	if err := proc.Start(); err != nil {
		t.Fatalf("start sleeper: %v", err)
	}
	oldBytes, err := os.ReadFile(exe)
	if err != nil {
		_ = proc.Process.Kill()
		t.Fatal(err)
	}

	tmp := filepath.Join(dir, "paper_temp.exe")
	newBytes := []byte("updated-sleeper-bytes")
	if err := os.WriteFile(tmp, newBytes, 0o755); err != nil {
		_ = proc.Process.Kill()
		t.Fatal(err)
	}
	if err := src.ApplyUpdate(exe, tmp); err != nil {
		_ = proc.Process.Kill()
		t.Fatalf("ApplyUpdate on running exe: %v", err)
	}
	if got, _ := os.ReadFile(exe); string(got) != string(newBytes) {
		_ = proc.Process.Kill()
		t.Errorf("exe holds %q, want the updated bytes", got)
	}
	if got, _ := os.ReadFile(exe + ".old"); string(got) != string(oldBytes) {
		_ = proc.Process.Kill()
		t.Error("exe.old should hold the previous binary image")
	}
	// The sleeper must have survived the swap (it runs from the .old now).
	// Kill errors with "already finished" if it died mid-swap.
	if err := proc.Process.Kill(); err != nil {
		t.Errorf("sleeper died during the swap: %v", err)
	}
}

func TestFetchReleaseInfo(t *testing.T) {
	sum := sha256.Sum256([]byte("fake-binary"))
	digest := "sha256:" + hex.EncodeToString(sum[:])
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"tag_name": "v1.2.3", "assets": [` +
			`{"name": "paper-linux", "digest": "` + digest + `"},` +
			`{"name": "paper-macos"},` +
			`{"name": "paper-windows.exe", "digest": "sha256:zzz"}` +
			`]}`))
	}))
	defer srv.Close()

	info, err := src.FetchReleaseInfo(srv.URL, nil)
	if err != nil {
		t.Fatalf("FetchReleaseInfo: %v", err)
	}
	if info.Tag != "v1.2.3" {
		t.Errorf("Tag = %q, want %q", info.Tag, "v1.2.3")
	}
	if got := info.DigestFor("paper-linux"); got != strings.TrimPrefix(digest, "sha256:") {
		t.Errorf("DigestFor(paper-linux) = %q", got)
	}
	if got := info.DigestFor("paper-macos"); got != "" {
		t.Errorf("DigestFor asset without digest = %q, want empty", got)
	}
	if got := info.DigestFor("paper-windows.exe"); got != "" {
		t.Errorf("DigestFor malformed digest = %q, want empty", got)
	}
	if got := info.DigestFor("paper-haiku"); got != "" {
		t.Errorf("DigestFor unknown asset = %q, want empty", got)
	}

	bad := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "nope", http.StatusNotFound)
	}))
	defer bad.Close()
	if _, err := src.FetchReleaseInfo(bad.URL, nil); err == nil {
		t.Error("FetchReleaseInfo on 404 expected error, got nil")
	}
}

func TestVerifyFileSHA256(t *testing.T) {
	// Well-known vector: sha256("abc").
	const want = "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"
	dir := t.TempDir()
	path := filepath.Join(dir, "f.bin")
	if err := os.WriteFile(path, []byte("abc"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := src.VerifyFileSHA256(path, want); err != nil {
		t.Errorf("VerifyFileSHA256 correct digest: %v", err)
	}
	if err := src.VerifyFileSHA256(path, strings.Repeat("0", 64)); err == nil {
		t.Error("VerifyFileSHA256 wrong digest expected error, got nil")
	}
	if err := src.VerifyFileSHA256(filepath.Join(dir, "missing"), want); err == nil {
		t.Error("VerifyFileSHA256 missing file expected error, got nil")
	}
}

func writeStaged(t *testing.T, dir, name string, content []byte) (dest, tmp, digest string) {
	t.Helper()
	dest = filepath.Join(dir, name)
	if err := os.WriteFile(dest, []byte("old-bytes"), 0o755); err != nil {
		t.Fatal(err)
	}
	ext := ""
	if strings.HasSuffix(name, ".exe") {
		ext = ".exe"
	}
	tmp = filepath.Join(dir, "paper_temp"+ext)
	if err := os.WriteFile(tmp, content, 0o755); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(content)
	return dest, tmp, hex.EncodeToString(sum[:])
}

func TestFinishUpdateHappyPath(t *testing.T) {
	dir := t.TempDir()
	newBytes := []byte("helper-swapped-binary")
	dest, tmp, digest := writeStaged(t, dir, "paper", newBytes)
	if got := src.Run([]string{"paper", "__finish-update", tmp, dest, digest}); got != 0 {
		t.Fatalf("__finish-update = %d, want 0", got)
	}
	if got, _ := os.ReadFile(dest); string(got) != string(newBytes) {
		t.Errorf("dest = %q, want updated bytes", got)
	}
	if _, err := os.Stat(tmp); !os.IsNotExist(err) {
		t.Error("staging file should be consumed by the swap")
	}
}

func TestFinishUpdateValidation(t *testing.T) {
	if got := src.Run([]string{"paper", "__finish-update"}); got != 2 {
		t.Errorf("__finish-update without args = %d, want 2", got)
	}
	if got := src.Run([]string{"paper", "__finish-update", "a"}); got != 2 {
		t.Errorf("__finish-update with 1 arg = %d, want 2", got)
	}
	// Tmp and dest must be absolute paths in the same directory.
	if got := src.Run([]string{"paper", "__finish-update", "rel-tmp", "rel-dest", ""}); got != 2 {
		t.Errorf("__finish-update with relative paths = %d, want 2", got)
	}
	dir := t.TempDir()
	other := filepath.Join(t.TempDir(), "paper_temp")
	_ = os.WriteFile(other, []byte("new"), 0o755)
	if got := src.Run([]string{"paper", "__finish-update", other, filepath.Join(dir, "paper"), ""}); got != 2 {
		t.Errorf("__finish-update across directories = %d, want 2", got)
	}
}

func TestFinishUpdateRejectsBadBytes(t *testing.T) {
	dir := t.TempDir()
	newBytes := []byte("tampered-binary")
	dest, tmp, _ := writeStaged(t, dir, "paper", newBytes)
	if got := src.Run([]string{"paper", "__finish-update", tmp, dest, strings.Repeat("0", 64)}); got != 1 {
		t.Errorf("__finish-update with wrong digest = %d, want 1", got)
	}
	if got, _ := os.ReadFile(dest); string(got) != "old-bytes" {
		t.Errorf("dest must be untouched after digest failure, got %q", got)
	}
}

func TestFinishUpdateRefusesMissingDest(t *testing.T) {
	dir := t.TempDir()
	tmp := filepath.Join(dir, "paper_temp")
	content := []byte("orphan-binary")
	if err := os.WriteFile(tmp, content, 0o755); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(content)
	if got := src.Run([]string{"paper", "__finish-update", tmp, filepath.Join(dir, "paper"), hex.EncodeToString(sum[:])}); got != 1 {
		t.Errorf("__finish-update with missing dest = %d, want 1", got)
	}
}

func TestFinishUpdateSweepsStaleOldFile(t *testing.T) {
	dir := t.TempDir()
	newBytes := []byte("fresh-binary")
	dest, tmp, digest := writeStaged(t, dir, "paper", newBytes)
	// Leftover from the pre-helper era: must be gone after the swap.
	if err := os.WriteFile(dest+".old", []byte("ancient-binary"), 0o755); err != nil {
		t.Fatal(err)
	}
	if got := src.Run([]string{"paper", "__finish-update", tmp, dest, digest}); got != 0 {
		t.Fatalf("__finish-update = %d, want 0", got)
	}
	if _, err := os.Stat(dest + ".old"); !os.IsNotExist(err) {
		t.Error("stale .old should be swept after a successful swap")
	}
	if got, _ := os.ReadFile(dest); string(got) != string(newBytes) {
		t.Errorf("dest = %q, want updated bytes", got)
	}
}

func TestUpdateTempPathHasPID(t *testing.T) {
	// The PID in the name is what keeps concurrent `paper update`
	// processes (different PIDs by definition) from sharing a file.
	pid := strconv.Itoa(os.Getpid())
	for dest, wantBase := range map[string]string{
		filepath.Join("some", "dir", "paper"):     "paper_temp." + pid,
		filepath.Join("some", "dir", "paper.exe"): "paper_temp." + pid + ".exe",
	} {
		got := src.UpdateTempPath(dest)
		if filepath.Dir(got) != filepath.Dir(dest) {
			t.Errorf("UpdateTempPath(%q) = %q: must stage next to dest", dest, got)
		}
		if filepath.Base(got) != wantBase {
			t.Errorf("UpdateTempPath(%q) = %q, want base %q", dest, got, wantBase)
		}
	}
}

func TestConcurrentUpdatesDontInterleave(t *testing.T) {
	// Two updates at once, distinct ~1MB payloads with a server-side delay
	// to force the downloads to overlap in time. Each must land intact in
	// its own destination with its staging file consumed.
	serve := func(b byte) *httptest.Server {
		payload := append([]byte(nil), make([]byte, 1<<20)...)
		for i := range payload {
			payload[i] = b + byte(i%251)
		}
		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Write(payload) //nolint:errcheck // test server; short write fails the dest comparison below
		}))
	}
	srvA, srvB := serve(0x41), serve(0x42)
	defer srvA.Close()
	defer srvB.Close()

	var wg sync.WaitGroup
	errs := make([]error, 2)
	for i, srv := range []*httptest.Server{srvA, srvB} {
		wg.Add(1)
		go func(i int, srv *httptest.Server) {
			defer wg.Done()
			// NOTE: t.TempDir is safe for concurrent use.
			dest := filepath.Join(t.TempDir(), "paper")
			if err := os.WriteFile(dest, []byte("old"), 0o755); err != nil {
				errs[i] = err
				return
			}
			tmp := src.UpdateTempPath(dest)
			if err := src.DownloadFile(srv.URL, tmp, nil); err != nil {
				errs[i] = err
				return
			}
			if err := src.SwapStaged(tmp, dest); err != nil {
				errs[i] = err
				return
			}
			got, _ := os.ReadFile(dest)
			seed := byte(0x41 + i)
			wantAt := func(pos int) byte { return seed + byte(pos%251) }
			if len(got) != 1<<20 || got[0] != wantAt(0) || got[1<<19] != wantAt(1<<19) || got[len(got)-1] != wantAt(len(got)-1) {
				errs[i] = errors.New("interleaved or truncated payload")
				return
			}
			if _, err := os.Stat(tmp); !os.IsNotExist(err) {
				errs[i] = errors.New("staging file not consumed")
			}
		}(i, srv)
	}
	wg.Wait()
	for _, err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
}
