package src

import (
	"os"
)

// Minimal ANSI styling (stdlib only, no dependencies).
// Palette: white, light blue, gray, with green/red reserved for
// success/error accents.

const (
	ansiReset = "\x1b[0m"
	ansiBold  = "\x1b[1m"
	ansiGray  = "\x1b[90m"
	ansiRed   = "\x1b[91m"
	ansiGreen = "\x1b[92m"
	ansiBlue  = "\x1b[94m"
	ansiWhite = "\x1b[97m"
)

// colorsOn is decided once: off when piped, when NO_COLOR is set, or on a
// dumb terminal. PAPER_COLOR=1 / CLICOLOR_FORCE=1 forces colors on.
var colorsOn = autoColor()

func autoColor() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	if os.Getenv("TERM") == "dumb" {
		return false
	}
	if os.Getenv("PAPER_COLOR") == "1" || os.Getenv("CLICOLOR_FORCE") == "1" {
		return true
	}
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

func paint(code, s string) string {
	if !colorsOn || s == "" {
		return s
	}
	return code + s + ansiReset
}

func bold(s string) string      { return paint(ansiBold, s) }
func white(s string) string     { return paint(ansiWhite, s) }
func lightBlue(s string) string { return paint(ansiBlue, s) }
func gray(s string) string      { return paint(ansiGray, s) }
func green(s string) string     { return paint(ansiGreen, s) }
func red(s string) string       { return paint(ansiRed, s) }
