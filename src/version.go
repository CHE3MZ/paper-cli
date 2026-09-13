// Package src holds all paper-cli logic. cmd/paper/main.go is only the
// thin CLI entrypoint.
package src

import (
	"fmt"
)

// Version returns the bundled Paper version (e.g. "1.21.11").
func Version() string { return EmbeddedVersion }

// BundledJarPath returns the repo-relative source of paper.jar
// (e.g. "paper-server/1.21.11/paper.jar").
func BundledJarPath() string { return EmbeddedBuildPath + "/paper.jar" }

// bundledFile reads a file from the staged embedded template.
// name is the base name, e.g. "paper.jar".
func bundledFile(name string) ([]byte, error) {
	data, err := bundledFS.ReadFile("bundled/" + name)
	if err != nil {
		return nil, fmt.Errorf("embedded %s not found (built with %q): %w", name, EmbeddedBuildPath, err)
	}
	return data, nil
}

// BundledPaperJar returns the bytes of the embedded paper.jar.
func BundledPaperJar() ([]byte, error) { return bundledFile("paper.jar") }
