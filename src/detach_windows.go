//go:build windows

package src

import "os/exec"

// detach is a no-op on Windows: Go children are not job-bound, so the
// helper already outlives the parent, and it inherits our console for its
// result message.
func detach(*exec.Cmd) {}
