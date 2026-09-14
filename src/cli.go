package src

import (
	"fmt"
	"os"
	"runtime"
	"strings"
)

// HelpText is printed by `paper help` (and --help/-h).
func HelpText() string {
	b := &strings.Builder{}
	title := bold(white("paper")) + gray(" — simple PaperMC server deployer (bundled Paper "+EmbeddedVersion+")")
	fmt.Fprintln(b, title)
	fmt.Fprintln(b)
	fmt.Fprintln(b, bold("Usage:"))
	fmt.Fprintln(b, "  "+lightBlue("paper")+gray(" <command> [options]"))
	fmt.Fprintln(b)
	fmt.Fprintln(b, bold("Commands:"))
	fmt.Fprintln(b, "  "+lightBlue("paper help")+"                           "+gray("prints this help text"))
	fmt.Fprintln(b, "  "+lightBlue("paper version")+"                        "+gray("prints the bundled PaperMC and CLI versions"))
	fmt.Fprintln(b, "  "+lightBlue("paper new")+gray(" . | PATH [--force]")+"         "+gray("creates a new paper server here or at PATH"))
	fmt.Fprintln(b, "  "+lightBlue("paper run")+gray(" [PATH] [flags]")+"             "+gray("runs the server (current directory by default)"))
	fmt.Fprintln(b, "  "+lightBlue("paper delete")+gray(" [PATH] [--confirm]")+"      "+gray("deletes everything but paper.jar (asks first)"))
	fmt.Fprintln(b, "  "+lightBlue("paper update")+"                         "+gray("updates this binary to the latest release (no-op when already current)"))
	fmt.Fprintln(b)
	fmt.Fprintln(b, bold("Run flags:"))
	fmt.Fprintln(b, "  "+lightBlue("--nogui")+gray("                              run as: java -jar paper.jar nogui"))
	fmt.Fprintln(b, "  "+lightBlue("--memory, -m")+gray(" <n>[mb|gb]              heap size, e.g. --memory=4gb or -m=512mb"))
	fmt.Fprintln(b, "  "+lightBlue("--optimized, -o")+gray("                     preset: 4GB max heap (ignored if --memory is set)"))
	fmt.Fprintln(b, "  "+lightBlue("--java")+gray(" <path>                        use a specific java binary"))
	fmt.Fprintln(b, "  "+lightBlue("--dry-run")+gray("                           print the java command without running it"))
	fmt.Fprintln(b)
	fmt.Fprintln(b, bold("Delete flags:"))
	fmt.Fprintln(b, "  "+lightBlue("--confirm, -y")+gray("                        delete without asking for confirmation"))
	fmt.Fprintln(b)
	fmt.Fprintln(b, bold("Examples:"))
	fmt.Fprintln(b, "  "+lightBlue("paper new ."))
	fmt.Fprintln(b, "  "+lightBlue("paper new ./my-server"))
	fmt.Fprintln(b, "  "+lightBlue("paper run --nogui -m=2gb"))
	fmt.Fprintln(b, "  "+lightBlue("paper run ./my-server -o"))
	fmt.Fprintln(b, "  "+lightBlue("paper delete --confirm"))
	fmt.Fprintln(b, "  "+lightBlue("paper version"))
	fmt.Fprintln(b, "  "+lightBlue("paper update"))
	return b.String()
}

// Run dispatches os.Args-style arguments (including program name).
// It returns the process exit code.
func Run(argv []string) int {
	enableConsoleANSI()
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
	case "version":
		return runVersion(args[1:])
	case "update":
		return runUpdate(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "%s unknown command %q\n\n%s", red("error:"), args[0], HelpText())
		return 2
	}
}

// errLine prints a red "error:" line to stderr.
func errLine(err error) {
	fmt.Fprintf(os.Stderr, "%s %v\n", red("error:"), err)
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
			fmt.Fprintf(os.Stderr, "%s unknown flag %q for `paper new`\n", red("error:"), a)
			return 2
		default:
			if pathArg != "" {
				fmt.Fprintln(os.Stderr, red("error:")+" too many arguments for `paper new`: expected [PATH]")
				return 2
			}
			pathArg = a
		}
	}
	if pathArg == "" {
		fmt.Fprintln(os.Stderr, red("error:")+" usage: paper new . | paper new PATH")
		return 2
	}
	dir, err := ResolveServerDir(pathArg)
	if err != nil {
		errLine(err)
		return 1
	}
	if err := NewServer(dir, NewOptions{Force: force}); err != nil {
		errLine(err)
		return 1
	}
	fmt.Printf("%s paper %s server in %s\n", green("created"), EmbeddedVersion, lightBlue(dir))
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
		case a == "--optimized" || a == "-o":
			opts.Optimized = true
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
			fmt.Fprintf(os.Stderr, "%s unknown flag %q for `paper run`\n", red("error:"), a)
			return 2
		default:
			if pathArg != "" {
				fmt.Fprintln(os.Stderr, red("error:")+" too many arguments for `paper run`: expected [PATH]")
				return 2
			}
			pathArg = a
		}
		i++
	}
	dir, err := ResolveServerDir(pathArg) // empty => current directory
	if err != nil {
		errLine(err)
		return 1
	}
	if err := RunServer(dir, opts); err != nil {
		errLine(err)
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
			fmt.Fprintf(os.Stderr, "%s unknown flag %q for `paper delete`\n", red("error:"), a)
			return 2
		default:
			if pathArg != "" {
				fmt.Fprintln(os.Stderr, red("error:")+" too many arguments for `paper delete`: expected [PATH]")
				return 2
			}
			pathArg = a
		}
	}
	dir, err := ResolveServerDir(pathArg) // empty => current directory
	if err != nil {
		errLine(err)
		return 1
	}
	if err := DeleteServer(dir, opts); err != nil {
		errLine(err)
		return 1
	}
	return 0
}

func runVersion(args []string) int {
	for _, a := range args {
		switch a {
		case "--help", "-h":
			fmt.Print(HelpText())
			return 0
		default:
			if strings.HasPrefix(a, "-") {
				fmt.Fprintf(os.Stderr, "%s unknown flag %q for `paper version`\n", red("error:"), a)
			} else {
				fmt.Fprintln(os.Stderr, red("error:")+" too many arguments for `paper version`: expected no arguments")
			}
			return 2
		}
	}
	fmt.Print(VersionText())
	return 0
}

func runUpdate(args []string) int {
	for _, a := range args {
		switch a {
		case "--help", "-h":
			fmt.Print(HelpText())
			return 0
		default:
			if strings.HasPrefix(a, "-") {
				fmt.Fprintf(os.Stderr, "%s unknown flag %q for `paper update`\n", red("error:"), a)
			} else {
				fmt.Fprintln(os.Stderr, red("error:")+" too many arguments for `paper update`: expected no arguments")
			}
			return 2
		}
	}
	asset := UpdateAssetForGOOS(runtime.GOOS)
	url := UpdateDownloadURL(UpdateRepo, asset)
	dest, err := UpdateTarget()
	if err != nil {
		errLine(err)
		return 1
	}
	// Best effort: when the check works and this binary is already the
	// latest release, say so instead of re-downloading. When the check
	// fails (offline, rate-limited), fall through to the download, which
	// is the safe default.
	if latest, lerr := LatestReleaseTag(LatestReleaseAPI, nil); lerr == nil && SameCLIVersion(CLIVersion, latest) {
		fmt.Printf("%s %s is already up to date.\n", green("paper"), lightBlue(CLIVersion))
		return 0
	}
	fmt.Printf("%s %s %s\n", gray("downloading"), lightBlue(asset), gray("from "+url))
	fmt.Printf("%s %s\n", gray("to:"), lightBlue(dest))
	if _, err := SelfUpdate(dest, url, nil); err != nil {
		errLine(err)
		return 1
	}
	// NOTE: do not print CLIVersion here: it is this (old) process's baked
	// version, not the downloaded binary's. The file on disk is new, but
	// this process image is still the old one.
	fmt.Printf("%s paper to %s\n%s\n", green("updated"), lightBlue(dest), gray("run `paper version` to confirm the new version."))
	return 0
}
