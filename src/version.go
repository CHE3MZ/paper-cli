// Package src holds all paper-cli logic. cmd/paper/main.go is only the
// thin CLI entrypoint.
package src

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// Version returns the bundled Paper version (e.g. "1.21.11").
func Version() string { return EmbeddedVersion }

// BundledJarPath returns the repo-relative source of paper.jar
// (e.g. "paper-server/1.21.11/paper.jar").
func BundledJarPath() string { return EmbeddedBuildPath + "/paper.jar" }

// bundledFile reads a file from the staged embedded template.
// name is the slash-separated path relative to the template root,
// e.g. "paper.jar" or "libraries/com/mojang/authlib/7.0.61/authlib-7.0.61.jar".
func bundledFile(name string) ([]byte, error) {
	data, err := bundledFS.ReadFile("bundled/" + name)
	if err != nil {
		return nil, fmt.Errorf("embedded %s not found (built with %q): %w", name, EmbeddedBuildPath, err)
	}
	return data, nil
}

// BundledPaperJar returns the bytes of the embedded paper.jar.
func BundledPaperJar() ([]byte, error) { return bundledFile("paper.jar") }

// ExtractBundled writes the whole embedded template tree (paper.jar,
// libraries/, cache/, versions/, plugins/, eula.txt, server.properties, ...)
// into dest, creating directories as needed. Files that already exist are
// overwritten only when force is true; otherwise the first existing file
// aborts with an error.
func ExtractBundled(dest string, force bool) (int, error) {
	var count int
	err := fs.WalkDir(bundledFS, "bundled", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == "bundled" {
			return nil
		}
		rel, err := filepath.Rel("bundled", filepath.FromSlash(path))
		if err != nil {
			return err
		}
		target := filepath.Join(dest, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		if _, serr := os.Stat(target); serr == nil && !force {
			return fmt.Errorf("%s already exists (use --force to overwrite)", target)
		}
		data, rerr := bundledFS.ReadFile(path)
		if rerr != nil {
			return rerr
		}
		if merr := os.MkdirAll(filepath.Dir(target), 0o755); merr != nil {
			return merr
		}
		if werr := os.WriteFile(target, data, 0o644); werr != nil {
			return fmt.Errorf("write %s: %w", target, werr)
		}
		count++
		return nil
	})
	if err != nil {
		return count, err
	}
	// Recreate dirs that were empty at build time (go:embed drops them).
	for _, dir := range bundledEmptyDirs {
		if merr := os.MkdirAll(filepath.Join(dest, filepath.FromSlash(dir)), 0o755); merr != nil {
			return count, merr
		}
	}
	return count, nil
}
