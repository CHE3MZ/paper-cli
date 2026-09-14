//go:build !windows

package src

import (
	"os/exec"
	"syscall"
)

// detach lets a spawned update helper outlive us: a new session means
// terminal signals (Ctrl+C, SIGHUP) addressed at our process group never
// reach it mid-swap. Its inherited stdio fds keep working.
func detach(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
}
