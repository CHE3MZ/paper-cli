package test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/CHE3MZ/paper-cli/src"
)

func TestResolveServerDir_DefaultsToCwd(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for _, in := range []string{"", "."} {
		got, err := src.ResolveServerDir(in)
		if err != nil {
			t.Fatalf("ResolveServerDir(%q): %v", in, err)
		}
		if got != cwd {
			t.Errorf("ResolveServerDir(%q) = %q, want cwd %q", in, got, cwd)
		}
	}
}

func TestNewAndDeleteServer(t *testing.T) {
	dir := t.TempDir()
	srv := filepath.Join(dir, "my-server")

	if err := src.NewServer(srv, src.NewOptions{}); err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	if !src.IsPaperServer(srv) {
		t.Fatal("expected paper.jar after NewServer")
	}
	jarInfo, _ := os.Stat(filepath.Join(srv, "paper.jar"))
	if jarInfo.Size() == 0 {
		t.Fatal("paper.jar is empty")
	}

	// Offline bundle: the whole template tree must be deployed so the
	// server starts with no downloads (libraries, caches, versions).
	for _, want := range []string{
		"libraries",
		"cache",
		"versions",
		"eula.txt",
		"server.properties",
		filepath.Join("plugins", ".paper-remapped"), // dot-dir: needs all: embed prefix
	} {
		if _, err := os.Stat(filepath.Join(srv, want)); err != nil {
			t.Fatalf("expected bundled %q in new server: %v", want, err)
		}
	}
	if got := countFiles(t, srv); got != src.EmbeddedFileCount {
		t.Fatalf("deployed %d files, embedded %d", got, src.EmbeddedFileCount)
	}

	// Second new without --force must fail.
	if err := src.NewServer(srv, src.NewOptions{}); err == nil {
		t.Fatal("expected error on duplicate NewServer without force")
	}
	// With force it succeeds.
	if err := src.NewServer(srv, src.NewOptions{Force: true}); err != nil {
		t.Fatalf("NewServer force: %v", err)
	}

	// Plant extra files that delete must remove.
	extras := []string{"server.properties", "logs", "world"}
	for _, name := range extras {
		p := filepath.Join(srv, name)
		if strings.HasSuffix(name, "properties") {
			_ = os.WriteFile(p, []byte("x=1\n"), 0o644)
		} else {
			_ = os.MkdirAll(p, 0o755)
			_ = os.WriteFile(filepath.Join(p, "dummy.txt"), []byte("hi"), 0o644)
		}
	}

	var out bytes.Buffer
	if err := src.DeleteServer(srv, src.DeleteOptions{Confirm: true, Stdout: &out}); err != nil {
		t.Fatalf("DeleteServer: %v", err)
	}
	if !src.IsPaperServer(srv) {
		t.Fatal("paper.jar must survive DeleteServer")
	}
	entries, _ := os.ReadDir(srv)
	if len(entries) != 1 || entries[0].Name() != "paper.jar" {
		names := []string{}
		for _, e := range entries {
			names = append(names, e.Name())
		}
		t.Fatalf("expected only paper.jar to remain, got %v", names)
	}
}

func TestDeleteServerRefusesWithoutJar(t *testing.T) {
	dir := t.TempDir()
	if err := src.DeleteServer(dir, src.DeleteOptions{Confirm: true}); err == nil {
		t.Fatal("expected refusal when no paper.jar present")
	}
}

func TestDeleteServerConfirmation_Abort(t *testing.T) {
	dir := t.TempDir()
	srv := filepath.Join(dir, "srv")
	if err := src.NewServer(srv, src.NewOptions{}); err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	_ = os.WriteFile(filepath.Join(srv, "junk.txt"), []byte("x"), 0o644)

	var out bytes.Buffer
	in := strings.NewReader("n\n")
	if err := src.DeleteServer(srv, src.DeleteOptions{Stdin: in, Stdout: &out}); err != nil {
		t.Fatalf("DeleteServer: %v", err)
	}
	if _, err := os.Stat(filepath.Join(srv, "junk.txt")); err != nil {
		t.Fatal("abort should keep junk.txt")
	}
}

func TestHelpTextCoversCommands(t *testing.T) {
	help := src.HelpText()
	for _, want := range []string{"paper new", "paper run", "paper delete", "paper help", "--nogui", "--memory", "--confirm"} {
		if !strings.Contains(help, want) {
			t.Errorf("help text missing %q", want)
		}
	}
}

func countFiles(t *testing.T, dir string) int {
	t.Helper()
	var n int
	err := filepath.WalkDir(dir, func(_ string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			n++
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return n
}
