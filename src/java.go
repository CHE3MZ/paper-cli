package src

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// RunOptions controls how `paper run` launches the server.
type RunOptions struct {
	// NoGUI appends the "nogui" argument (java -jar paper.jar nogui).
	NoGUI bool
	// Memory is the raw user input, e.g. "4gb", "512mb". Empty = no limit flags.
	Memory string
	// Java is an optional override for the java binary. Empty = auto-detect.
	Java string
	// DryRun prints the resolved command without executing it.
	DryRun bool
	// ExtraArgs are appended after the jar args (advanced escape hatch).
	ExtraArgs []string
}

var memPattern = regexp.MustCompile(`(?i)^\s*(\d+)\s*(m|mb|g|gb)?\s*$`)

// NormalizeMemory validates user input like "4gb", "512mb", "1024", "2G" and
// returns it in java -Xmx form ("4G", "512M"). Empty input returns "".
// A bare number is treated as megabytes.
func NormalizeMemory(input string) (string, error) {
	if strings.TrimSpace(input) == "" {
		return "", nil
	}
	m := memPattern.FindStringSubmatch(input)
	if m == nil {
		return "", fmt.Errorf("invalid memory %q: use a number with optional mb/gb suffix, e.g. --memory=4gb or --memory=512mb", input)
	}
	amount := m[1]
	unit := strings.ToLower(m[2])
	switch unit {
	case "", "m", "mb":
		return amount + "M", nil
	case "g", "gb":
		return amount + "G", nil
	default:
		return "", fmt.Errorf("invalid memory %q", input)
	}
}

// ResolveJava returns the java binary to use: the override when set,
// otherwise the first "java" found on PATH.
func ResolveJava(override string) (string, error) {
	if override != "" {
		return override, nil
	}
	path, err := exec.LookPath("java")
	if err != nil {
		return "", fmt.Errorf("java not found on PATH (install a Java 21+ runtime, or pass --java=/path/to/java): %w", err)
	}
	return path, nil
}

// BuildJavaCommand builds the process invocation for a server directory.
// It always runs `java -jar paper.jar` inside dir; memory adds matching
// -Xms/-Xmx flags; nogui appends "nogui".
func BuildJavaCommand(javaBin, dir string, opts RunOptions) (bin string, args []string, err error) {
	javaBin, err = ResolveJava(javaBinOrOverride(javaBin, opts))
	if err != nil {
		return "", nil, err
	}
	mem, err := NormalizeMemory(opts.Memory)
	if err != nil {
		return "", nil, err
	}
	args = []string{}
	if mem != "" {
		args = append(args, "-Xms"+mem, "-Xmx"+mem)
	}
	args = append(args, "-jar", "paper.jar")
	if opts.NoGUI {
		args = append(args, "nogui")
	}
	args = append(args, opts.ExtraArgs...)
	_ = dir
	return javaBin, args, nil
}

func javaBinOrOverride(javaBin string, opts RunOptions) string {
	if opts.Java != "" {
		return opts.Java
	}
	return javaBin
}

// FormatCommand renders a command for --dry-run / help output.
func FormatCommand(bin string, args []string) string {
	parts := append([]string{bin}, args...)
	for i, p := range parts {
		if strings.ContainsAny(p, " \t\"") {
			parts[i] = `"` + strings.ReplaceAll(p, `"`, `\"`) + `"`
		}
	}
	return strings.Join(parts, " ")
}
