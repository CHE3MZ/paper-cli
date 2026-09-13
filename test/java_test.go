package test

import (
	"testing"

	"github.com/CHE3MZ/paper-cli/src"
)

func TestNormalizeMemory(t *testing.T) {
	cases := map[string]string{
		"4gb":   "4G",
		"4GB":   "4G",
		"4g":    "4G",
		"512mb": "512M",
		"512M":  "512M",
		"1024":  "1024M",
		"2G":    "2G",
		"":      "",
	}
	for in, want := range cases {
		got, err := src.NormalizeMemory(in)
		if err != nil {
			t.Errorf("NormalizeMemory(%q) unexpected error: %v", in, err)
			continue
		}
		if got != want {
			t.Errorf("NormalizeMemory(%q) = %q, want %q", in, got, want)
		}
	}

	for _, bad := range []string{"abc", "4tb", "gb", "-4gb", "4.5gb"} {
		if _, err := src.NormalizeMemory(bad); err == nil {
			t.Errorf("NormalizeMemory(%q) expected error, got nil", bad)
		}
	}
}

func TestBuildJavaCommand(t *testing.T) {
	bin, args, err := src.BuildJavaCommand("java", "/tmp/srv", src.RunOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertContains(t, args, []string{"-jar", "paper.jar"})
	assertNotContains(t, args, "nogui")

	_, args, err = src.BuildJavaCommand("java", "/tmp/srv", src.RunOptions{NoGUI: true, Memory: "4gb"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertContains(t, args, []string{"-Xmx4G", "-Xms4G", "-jar", "paper.jar", "nogui"})

	_, args, err = src.BuildJavaCommand("java", "/tmp/srv", src.RunOptions{Memory: "512mb"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertContains(t, args, []string{"-Xmx512M", "-Xms512M"})

	_, args, err = src.BuildJavaCommand("java", "/tmp/srv", src.RunOptions{Optimized: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertContains(t, args, []string{"-Xmx4G", "-Xms4G"})

	// Explicit memory wins over --optimized.
	_, args, err = src.BuildJavaCommand("java", "/tmp/srv", src.RunOptions{Optimized: true, Memory: "2gb"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertContains(t, args, []string{"-Xmx2G", "-Xms2G"})
	assertNotContains(t, args, "-Xmx4G")

	if _, _, err = src.BuildJavaCommand("java", "/tmp/srv", src.RunOptions{Memory: "bogus"}); err == nil {
		t.Error("expected error for bad memory, got nil")
	}
	_ = bin
}

func assertContains(t *testing.T, args, want []string) {
	t.Helper()
	set := map[string]bool{}
	for _, a := range args {
		set[a] = true
	}
	for _, w := range want {
		if !set[w] {
			t.Errorf("args %v missing %q", args, w)
		}
	}
}

func assertNotContains(t *testing.T, args []string, want string) {
	t.Helper()
	for _, a := range args {
		if a == want {
			t.Errorf("args %v should not contain %q", args, want)
		}
	}
}
