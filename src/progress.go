package src

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

// progressBarWidth is the bar width in cells. Fixed (no terminal probing,
// stdlib only): 30 cells plus the trailing text stays inside 80 columns.
const progressBarWidth = 30

// isTerminal reports whether stdout is an interactive terminal as opposed
// to a pipe or file. The progress bar only draws when true, so pipes and
// log files stay clean. (Separate from colorsOn: colors may be forced on
// while piped, but a live \r bar must never go to a log.)
func isTerminal() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

// ProgressWriter counts bytes through to w while drawing a single-line
// download progress bar on stdout (terminals only; silent otherwise).
type ProgressWriter struct {
	w        io.Writer
	total    int64
	written  int64
	start    time.Time
	lastDraw time.Time
	drew     bool
}

// NewProgressWriter wraps w, expecting total bytes (<=0 when unknown).
func NewProgressWriter(w io.Writer, total int64) *ProgressWriter {
	return &ProgressWriter{w: w, total: total, start: time.Now()}
}

func (p *ProgressWriter) Write(b []byte) (int, error) {
	n, err := p.w.Write(b)
	if n > 0 {
		p.written += int64(n)
		p.maybeDraw()
	}
	return n, err
}

func (p *ProgressWriter) maybeDraw() {
	if !isTerminal() {
		return
	}
	now := time.Now()
	if p.drew && now.Sub(p.lastDraw) < 100*time.Millisecond {
		return
	}
	p.lastDraw = now
	p.drew = true
	fmt.Printf("\r%s", DownloadLine(p.written, p.total, now.Sub(p.start)))
}

// Finish draws the final line and ends it, so the next output starts clean.
func (p *ProgressWriter) Finish() {
	if !isTerminal() {
		return
	}
	p.drew = true
	fmt.Printf("\r%s\n", DownloadLine(p.written, p.total, time.Since(p.start)))
}

// Abort ends an interrupted bar line, so a following error starts clean.
func (p *ProgressWriter) Abort() {
	if p.drew {
		fmt.Print("\n")
	}
}

// DownloadLine renders one progress line (no trailing newline). Pure
// function of its inputs, so tests pin the format. When total is unknown
// (server sent no length) it degrades to a byte counter.
func DownloadLine(written, total int64, elapsed time.Duration) string {
	mb := func(b int64) float64 { return float64(b) / 1048576 }
	speed := 0.0
	if elapsed > 0 {
		speed = float64(written) / elapsed.Seconds() / 1048576
	}
	speedText := "-- MB/s"
	if speed > 0 {
		speedText = fmt.Sprintf("%.1f MB/s", speed)
	}
	if total <= 0 {
		return fmt.Sprintf("%s downloaded  %s",
			white(fmt.Sprintf("%.1f MB", mb(written))),
			gray(speedText))
	}
	pct := int(written * 100 / total)
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}
	filled := pct * progressBarWidth / 100
	bar := "[" + lightBlue(strings.Repeat("=", filled)) + gray(strings.Repeat("-", progressBarWidth-filled)) + "]"
	line := fmt.Sprintf("%s %s  %s  %s",
		bar,
		bold(white(fmt.Sprintf("%3d%%", pct))),
		fmt.Sprintf("%.1f/%.1f MB", mb(written), mb(total)),
		gray(speedText))
	if pct < 100 && speed > 0 {
		remaining := float64(total-written) / 1048576 / speed
		line += "  " + gray("ETA "+formatDur(remaining))
	}
	return line
}

// formatDur renders seconds as "12s" or "3m05s".
func formatDur(sec float64) string {
	s := int(sec + 0.5)
	if s < 0 {
		s = 0
	}
	if s < 60 {
		return fmt.Sprintf("%ds", s)
	}
	return fmt.Sprintf("%dm%02ds", s/60, s%60)
}
