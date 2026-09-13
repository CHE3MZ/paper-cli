//go:build !windows

package src

// enableConsoleANSI is a no-op outside Windows (ANSI works by default).
func enableConsoleANSI() {}
