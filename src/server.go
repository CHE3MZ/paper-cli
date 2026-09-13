package src

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

const jarName = "paper.jar"

// defaultEula is written by `paper new` when the template has no eula.txt.
// It pre-accepts the EULA so `paper run` works immediately offline.
const defaultEula = "#By changing the setting below to TRUE you are indicating your agreement to our EULA (https://aka.ms/MinecraftEULA).\neula=true\n"

// ResolveServerDir turns an optional CLI path into an absolute server dir.
// Empty string and "." both mean the current directory, so `paper run` and
// `paper delete` work with no path argument.
func ResolveServerDir(pathArg string) (string, error) {
	if strings.TrimSpace(pathArg) == "" || pathArg == "." {
		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("resolve current directory: %w", err)
		}
		return cwd, nil
	}
	abs, err := filepath.Abs(pathArg)
	if err != nil {
		return "", fmt.Errorf("resolve path %q: %w", pathArg, err)
	}
	return abs, nil
}

// JarPath returns dir/paper.jar.
func JarPath(dir string) string { return filepath.Join(dir, jarName) }

// IsPaperServer reports whether dir contains a paper.jar.
func IsPaperServer(dir string) bool {
	st, err := os.Stat(JarPath(dir))
	return err == nil && !st.IsDir()
}

// NewOptions controls `paper new`.
type NewOptions struct {
	// Force overwrites an existing paper.jar instead of failing.
	Force bool
}

// NewServer deploys the embedded paper.jar (+ template defaults) into dir.
func NewServer(dir string, opts NewOptions) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create directory %s: %w", dir, err)
	}
	jarDest := JarPath(dir)
	if _, err := os.Stat(jarDest); err == nil && !opts.Force {
		return fmt.Errorf("paper.jar already exists in %s (use --force to overwrite)", dir)
	}

	jarBytes, err := BundledPaperJar()
	if err != nil {
		return err
	}
	if err := os.WriteFile(jarDest, jarBytes, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", jarDest, err)
	}

	// eula.txt: prefer the embedded template, fall back to defaultEula.
	eulaDest := filepath.Join(dir, "eula.txt")
	if _, err := os.Stat(eulaDest); os.IsNotExist(err) || opts.Force {
		var eula []byte
		if EmbeddedHasEula {
			if b, rerr := bundledFile("eula.txt"); rerr == nil {
				eula = b
			}
		}
		if eula == nil {
			eula = []byte(defaultEula)
		}
		if werr := os.WriteFile(eulaDest, eula, 0o644); werr != nil {
			return fmt.Errorf("write %s: %w", eulaDest, werr)
		}
	}

	// server.properties: only from the embedded template when present and
	// missing in the target (first run of the jar generates it otherwise).
	propsDest := filepath.Join(dir, "server.properties")
	if EmbeddedHasServerProps {
		if _, err := os.Stat(propsDest); os.IsNotExist(err) {
			if b, rerr := bundledFile("server.properties"); rerr == nil {
				_ = os.WriteFile(propsDest, b, 0o644)
			}
		}
	}

	return nil
}

// RunServer launches `java -jar paper.jar` in dir, streaming stdio.
// dir defaults to the current directory via ResolveServerDir by the caller.
func RunServer(dir string, opts RunOptions) error {
	if !IsPaperServer(dir) {
		return fmt.Errorf("no paper.jar found in %s (run `paper new %s` first)", dir, dirLabel(dir))
	}
	javaBin, args, err := BuildJavaCommand("", dir, opts)
	if err != nil {
		return err
	}
	if opts.DryRun {
		fmt.Printf("(in %s)\n%s\n", dir, FormatCommand(javaBin, args))
		return nil
	}
	cmd := exec.Command(javaBin, args...)
	cmd.Dir = dir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		if exit, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("server exited with code %d", exit.ExitCode())
		}
		return fmt.Errorf("run server: %w", err)
	}
	return nil
}

func dirLabel(dir string) string {
	cwd, err := os.Getwd()
	if err != nil || cwd == dir {
		return "."
	}
	if rel, err := filepath.Rel(cwd, dir); err == nil && !strings.HasPrefix(rel, "..") {
		return rel
	}
	return dir
}

// DeleteOptions controls `paper delete`.
type DeleteOptions struct {
	// Confirm skips the interactive prompt (paper delete --confirm).
	Confirm bool
	// Stdin/Stdout allow tests to drive the confirmation prompt.
	Stdin  io.Reader
	Stdout io.Writer
}

// DeleteServer removes everything in dir EXCEPT paper.jar.
// It refuses to run in a directory without paper.jar (safety: that means it
// is probably not a paper server).
func DeleteServer(dir string, opts DeleteOptions) error {
	if !IsPaperServer(dir) {
		return fmt.Errorf("no paper.jar found in %s: refusing to delete (not a paper server?)", dir)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read %s: %w", dir, err)
	}
	var victims []string
	for _, e := range entries {
		if e.Name() == jarName {
			continue
		}
		victims = append(victims, filepath.Join(dir, e.Name()))
	}
	if len(victims) == 0 {
		fmt.Fprintln(outOrStdout(opts.Stdout), "nothing to delete (only paper.jar remains).")
		return nil
	}
	sort.Strings(victims)
	if !opts.Confirm {
		fmt.Fprintf(outOrStdout(opts.Stdout), "Delete %d file(s) in %s (keeping paper.jar)? [y/N]: ", len(victims), dir)
		reader := bufio.NewReader(inOrStdin(opts.Stdin))
		line, _ := reader.ReadString('\n')
		line = strings.TrimSpace(strings.ToLower(line))
		if line != "y" && line != "yes" {
			fmt.Fprintln(outOrStdout(opts.Stdout), "aborted.")
			return nil
		}
	}
	for _, v := range victims {
		if err := os.RemoveAll(v); err != nil {
			return fmt.Errorf("delete %s: %w", v, err)
		}
	}
	fmt.Fprintf(outOrStdout(opts.Stdout), "deleted %d file(s), kept %s.\n", len(victims), jarName)
	return nil
}

// WalkDeletable lists what DeleteServer would remove (used by tests/docs).
func WalkDeletable(dir string) ([]string, error) {
	var out []string
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if p == dir || filepath.Base(p) == jarName && filepath.Dir(p) == dir {
			return nil
		}
		// Only top-level entries matter for the report.
		if filepath.Dir(p) == dir {
			out = append(out, p)
		}
		return nil
	})
	return out, err
}

func inOrStdin(r io.Reader) io.Reader {
	if r != nil {
		return r
	}
	return os.Stdin
}

func outOrStdout(w io.Writer) io.Writer {
	if w != nil {
		return w
	}
	return os.Stdout
}
