// Package src holds all paper-cli logic. cmd/paper/main.go is only the
// thin CLI entrypoint.
package src

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Version returns the bundled Paper version (e.g. "1.21.11").
func Version() string { return EmbeddedVersion }

// BundledJarPath returns the repo-relative source of paper.jar
// (e.g. "paper-server/1.21.11/paper.jar").
func BundledJarPath() string { return EmbeddedBuildPath + "/paper.jar" }

// ExtractBundled writes the whole embedded template archive (paper.jar,
// libraries/, cache/, eula.txt, server.properties, ...) into dest, creating
// directories as needed. The archive is gzip-compressed: decompression is
// lossless (bit-identical files) and CRC-checked, so a corrupt bundle fails
// here instead of deploying bad data. Files that already exist are
// overwritten only when force is true; otherwise the first existing file
// aborts with an error. It returns the number of files deployed.
func ExtractBundled(dest string, force bool) (int, error) {
	gr, err := gzip.NewReader(bytes.NewReader(bundledArchive))
	if err != nil {
		return 0, fmt.Errorf("embedded bundle is corrupt (built with %q): %w", EmbeddedBuildPath, err)
	}
	defer gr.Close()

	var count int
	tr := tar.NewReader(gr)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return count, fmt.Errorf("read embedded bundle (built with %q): %w", EmbeddedBuildPath, err)
		}
		if err := extractEntry(tr, hdr, dest, force); err != nil {
			return count, err
		}
		if hdr.Typeflag == tar.TypeReg {
			count++
		}
	}
	return count, nil
}

func extractEntry(tr *tar.Reader, hdr *tar.Header, dest string, force bool) error {
	// Reject absolute paths and .. escapes: entries must stay inside dest.
	rel := filepath.FromSlash(filepath.Clean("/" + hdr.Name))
	target, err := filepath.Abs(filepath.Join(dest, rel))
	if err != nil {
		return err
	}
	base, err := filepath.Abs(dest)
	if err != nil {
		return err
	}
	if target != base && !strings.HasPrefix(target, base+string(filepath.Separator)) {
		return fmt.Errorf("embedded bundle entry escapes target dir: %q", hdr.Name)
	}

	switch hdr.Typeflag {
	case tar.TypeDir:
		return os.MkdirAll(target, 0o755)
	case tar.TypeReg:
		if _, serr := os.Stat(target); serr == nil && !force {
			return fmt.Errorf("%s already exists (use --force to overwrite)", target)
		}
		if merr := os.MkdirAll(filepath.Dir(target), 0o755); merr != nil {
			return merr
		}
		out, cerr := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
		if cerr != nil {
			return cerr
		}
		if _, cerr := io.Copy(out, tr); cerr != nil {
			out.Close()
			return fmt.Errorf("write %s: %w", target, cerr)
		}
		return out.Close()
	default:
		return fmt.Errorf("unsupported entry in embedded bundle: %q", hdr.Name)
	}
}
