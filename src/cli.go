package src

import (
	"fmt"
	"os"
	"strings"
)

// HelpText is printed by `paper help` (and --help/-h).
func HelpText() string {
	return fmt.Sprintf(`paper - simple PaperMC server deployer (bundled Paper %s)

Usage:
  paper help                          prints this help text
  paper new .                         creates a new paper server here (dot = current dir)
  paper new PATH                      creates a new paper server at PATH
  paper new [PATH] [--force]          overwrite an existing paper.jar

  paper run                           runs paper in the current directory
  paper run [PATH] [flags]            runs paper in PATH

  paper delete                        deletes everything BUT paper.jar here (asks first)
  paper delete [PATH] [--confirm]     deletes everything BUT paper.jar at PATH

Run flags:
  --nogui                 run as: java -jar paper.jar nogui
  --memory=<n>[mb|gb]     limit heap, e.g. --memory=4gb or --memory=512mb
  -m <n>[mb|gb]           shorthand for --memory (both -m=4gb and -m 4gb work)
  --java=<path>           use a specific java binary
  --dry-run               print the java command without running it

Delete flags:
  --confirm, -y           delete without asking for confirmation

Examples:
  paper new .
  paper new ./my-server
  paper run
  paper run ./my-server --nogui -m=2gb
  paper delete --confirm
`, EmbeddedVersion)
}

// Run dispatches os.Args-style arguments (including program name).
// It returns the process exit code.
func Run(argv []string) int {
	args := argv[1:] // drop program name
	if len(args) == 0 {
		fmt.Print(HelpText())
		return 0
	}
	// Global help shorthands: paper --help, paper -h
	if len(args) == 1 && (args[0] == "--help" || args[0] == "-h" || args[0] == "help") {
		fmt.Print(HelpText())
		return 0
	}

	switch args[0] {
	case "help":
		fmt.Print(HelpText())
		return 0
	case "new":
		return runNew(args[1:])
	case "run":
		return runRun(args[1:])
	case "delete":
		return runDelete(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n%s", args[0], HelpText())
		return 2
	}
}

func runNew(args []string) int {
	var pathArg string
	force := false
	for _, a := range args {
		switch {
		case a == "--force" || a == "-f":
			force = true
		case a == "--help" || a == "-h":
			fmt.Print(HelpText())
			return 0
		case strings.HasPrefix(a, "-"):
			fmt.Fprintf(os.Stderr, "unknown flag %q for `paper new`\n", a)
			return 2
		default:
			if pathArg != "" {
				fmt.Fprintln(os.Stderr, "too many arguments for `paper new`: expected [PATH]")
				return 2
			}
			pathArg = a
		}
	}
	if pathArg == "" {
		fmt.Fprintln(os.Stderr, "usage: paper new . | paper new PATH")
		return 2
	}
	dir, err := ResolveServerDir(pathArg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}
	if err := NewServer(dir, NewOptions{Force: force}); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}
	fmt.Printf("created paper %s server in %s\n", EmbeddedVersion, dir)
	return 0
}

func runRun(args []string) int {
	var pathArg string
	opts := RunOptions{}
	i := 0
	for i < len(args) {
		a := args[i]
		switch {
		case a == "--nogui":
			opts.NoGUI = true
		case a == "--dry-run":
			opts.DryRun = true
		case strings.HasPrefix(a, "--memory="):
			opts.Memory = strings.TrimPrefix(a, "--memory=")
		case a == "--memory" && i+1 < len(args):
			i++
			opts.Memory = args[i]
		case strings.HasPrefix(a, "-m="):
			opts.Memory = strings.TrimPrefix(a, "-m=")
		case a == "-m" && i+1 < len(args):
			i++
			opts.Memory = args[i]
		case len(a) > 2 && strings.HasPrefix(a, "-m") && !strings.HasPrefix(a, "--"):
			// -m4gb / -m512mb attached form
			opts.Memory = strings.TrimPrefix(a, "-m")
		case strings.HasPrefix(a, "--java="):
			opts.Java = strings.TrimPrefix(a, "--java=")
		case a == "--java" && i+1 < len(args):
			i++
			opts.Java = args[i]
		case a == "--":
			opts.ExtraArgs = append(opts.ExtraArgs, args[i+1:]...)
			i = len(args)
		case a == "--help" || a == "-h":
			fmt.Print(HelpText())
			return 0
		case strings.HasPrefix(a, "-") && a != "-":
			fmt.Fprintf(os.Stderr, "unknown flag %q for `paper run`\n", a)
			return 2
		default:
			if pathArg != "" {
				fmt.Fprintln(os.Stderr, "too many arguments for `paper run`: expected [PATH]")
				return 2
			}
			pathArg = a
		}
		i++
	}
	dir, err := ResolveServerDir(pathArg) // empty => current directory
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}
	if err := RunServer(dir, opts); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}
	return 0
}

func runDelete(args []string) int {
	var pathArg string
	opts := DeleteOptions{}
	for _, a := range args {
		switch {
		case a == "--confirm" || a == "-y":
			opts.Confirm = true
		case a == "--help" || a == "-h":
			fmt.Print(HelpText())
			return 0
		case strings.HasPrefix(a, "-"):
			fmt.Fprintf(os.Stderr, "unknown flag %q for `paper delete`\n", a)
			return 2
		default:
			if pathArg != "" {
				fmt.Fprintln(os.Stderr, "too many arguments for `paper delete`: expected [PATH]")
				return 2
			}
			pathArg = a
		}
	}
	dir, err := ResolveServerDir(pathArg) // empty => current directory
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}
	if err := DeleteServer(dir, opts); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}
	return 0
}
