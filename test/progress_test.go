package test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/CHE3MZ/paper-cli/src"
)

const mib = 1048576

// Colors are off under `go test` (piped stdout), so the render is plain
// text here and these exact strings pin the format.
func TestDownloadLineHalfway(t *testing.T) {
	got := src.DownloadLine(90*mib, 180*mib, 10*time.Second)
	want := "[===============---------------]  50%  90.0/180.0 MB  9.0 MB/s  ETA 10s"
	if got != want {
		t.Errorf("got  %q\nwant %q", got, want)
	}
}

func TestDownloadLineStart(t *testing.T) {
	got := src.DownloadLine(0, 180*mib, 0)
	want := "[------------------------------]   0%  0.0/180.0 MB  -- MB/s"
	if got != want {
		t.Errorf("got  %q\nwant %q", got, want)
	}
}

func TestDownloadLineDone(t *testing.T) {
	got := src.DownloadLine(180*mib, 180*mib, 20*time.Second)
	want := "[==============================] 100%  180.0/180.0 MB  9.0 MB/s"
	if got != want {
		t.Errorf("got  %q\nwant %q", got, want)
	}
}

func TestDownloadLineUnknownTotal(t *testing.T) {
	got := src.DownloadLine(5*mib, -1, 5*time.Second)
	want := "5.0 MB downloaded  1.0 MB/s"
	if got != want {
		t.Errorf("got  %q\nwant %q", got, want)
	}
}

func TestDownloadLineLongETA(t *testing.T) {
	got := src.DownloadLine(90*mib, 180*mib, 65*time.Second)
	if !strings.Contains(got, "ETA 1m05s") || !strings.Contains(got, "1.4 MB/s") {
		t.Errorf("got %q, want it to contain ETA 1m05s and 1.4 MB/s", got)
	}
}

func TestDownloadLineClampsOverflow(t *testing.T) {
	got := src.DownloadLine(200*mib, 180*mib, 10*time.Second)
	if !strings.Contains(got, "100%") || strings.Contains(got, "ETA") {
		t.Errorf("got %q, want clamped 100%% with no ETA", got)
	}
}

func TestProgressWriterPassesBytesThrough(t *testing.T) {
	var buf bytes.Buffer
	pw := src.NewProgressWriter(&buf, 1000)
	for _, chunk := range [][]byte{{}, []byte("hello "), []byte("world")} {
		n, err := pw.Write(chunk)
		if err != nil {
			t.Fatalf("Write: %v", err)
		}
		if n != len(chunk) {
			t.Fatalf("Write returned %d, want %d", n, len(chunk))
		}
	}
	if buf.String() != "hello world" {
		t.Errorf("passthrough = %q", buf.String())
	}
	pw.Finish()
	pw.Abort() // must not panic or print when nothing was drawn
}
