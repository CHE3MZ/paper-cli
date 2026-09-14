<img src="images/logo.png" alt="Paper logo" width="320" />

# Paper CLI
---

Paper CLI is a small tool that deploys PaperMC servers. The whole thing is
one binary called `paper`, and it works offline: the server files are packed
inside the binary itself, so creating a new server is just unpacking — no
downloads, no installers, no browser.

If you just want to run a server, the core loop is three commands. Everything
else in these docs is detail you can read when you need it.

## Install

```
install\windows.bat
sh install/macos.sh  # macOS -> ~/.local/bin/paper
sh install/linux.sh  # Linux -> ~/.local/bin/paper
```

### Uninstall

```
install\uninstall-windows.bat
sh install/uninstall-macos.sh
sh install/uninstall-linux.sh
```

## Quickstart

```
paper new ./my-server     # create a server
paper run ./my-server     # run it (java -jar paper.jar, basically)
paper delete ./my-server  # wipe everything but paper.jar when you're done
```

Leave out the path and the command uses the folder you're standing in:

```
cd my-server
paper run
```

Pick your platform on the left for setup instructions (Java, getting the
binary), then come back here for the full command list.

## Commands

**paper help** — prints the help text. `paper --help` and `paper -h` do the
same thing.

**paper new . | PATH** — creates a new server here (`.`) or at `PATH`.
The folder is created if it doesn't exist. If a `paper.jar` is already
there, it refuses — pass `--force` (or `-f`) to overwrite it.

**paper run [PATH]** — starts the server. Under the hood this runs
`java -jar paper.jar` inside the server folder, with your terminal attached
so you get the console, commands, and Ctrl+C handling for free.

| Flag | What it does |
| ---- | ------------ |
| `--nogui` | Appends `nogui`, so no server GUI window pops up |
| `--memory, -m <n>[mb\|gb]` | Heap size, e.g. `--memory=4gb` or `-m=512mb` (sets `-Xms` and `-Xmx` to the same value) |
| `--optimized, -o` | Shortcut for 4GB of heap. Ignored if you also pass `--memory` |
| `--java <path>` | Use a specific Java binary instead of the one on your PATH |
| `--dry-run` | Print the java command without running it — handy for checking flags |

```
paper run ./my-server --nogui -m=2gb
paper run ./my-server -o
paper run --dry-run --memory=512mb
```

**paper delete [PATH]** — deletes everything in the server folder
*except* `paper.jar`: worlds, configs, logs, all of it. It asks for
confirmation first, unless you pass `--confirm` (or `-y`). As a safety
rule, it refuses to touch a folder with no `paper.jar` in it.

```
paper delete --confirm
paper delete ./my-server -y
```

**paper version** — prints the two versions that matter: the bundled
PaperMC version (`Paper MC Version`, e.g. `1.21.11`) and the CLI's own
release version (`Paper CLI Version`, e.g. `v1.0.0`, baked in at build
time from the release tag). A local build without version stamping
reports `dev`.

```
paper version
```

**paper update** — self-updates to the latest release. It downloads the
right binary for your OS from
`github.com/CHE3MZ/paper-cli/releases/latest` (the same file the
install scripts fetch), stages it as `paper_temp.<pid>` next to the binary
you're running, then swaps it in. On an interactive terminal a live
progress bar tracks the download (percent, megabytes, speed, ETA); piped
output stays clean. The swap itself runs in a helper —
which is just `paper` re-executed with a hidden internal command, so the
install stays one file and there is no script to tamper with. The download
is checksum-verified against the digest published with the release before
anything is swapped.

```
paper update
```

If you're already on the latest release it just says so and downloads
nothing. On Linux/macOS the helper runs synchronously (same `updated`
confirmation as always); on Windows a running `.exe` can't be overwritten
in place, so there the helper runs detached after `paper` exits and prints
the result itself — run `paper version` to confirm. On Linux/macOS nothing
is left behind: the staging file is consumed by the swap. On Windows the
helper itself runs from the binary being replaced, so every update renames
it aside to `paper.exe.old` first; that file is cleaned up automatically
the next time you run paper, and at most one ever lingers. Stale staging
files from interrupted downloads are swept the same way.

## Colors

Output is colored on a real terminal (white, light blue, gray) and plain
everywhere else, so pipes and log files stay clean. `PAPER_COLOR=1` forces
colors on, `NO_COLOR=1` forces them off.

## Why offline?

A normal Paper setup downloads its libraries and version files on first
launch. Paper CLI ships those files inside the binary instead
(`paper.jar`, `libraries/`, `cache/`, configs), so `paper new` works on a
machine with no internet. The only things left out are files the server
recreates by itself on boot (`versions/`, `logs/`, plugin caches) — which
also keeps the binary smaller (~177MB instead of ~244MB).

Want to know how all of this is put together? The
[Developers](developers.md) page walks through the source file by file.

## Legal in brief

Your servers run PaperMC (GPL-3.0) and Mojang's server code (proprietary).
Running one means you accept the
[Minecraft EULA](https://aka.ms/MinecraftEULA) — note that `paper new`
pre-accepts it in `eula.txt` so the server starts right away. The full
attribution, the GPL copy, and the redistribution tradeoff of the offline
bundle are documented in `NOTICE.md` in the repo. This project isn't
affiliated with Mojang, Microsoft, or PaperMC.

<img src="images/logo-small.png" alt="Paper logo small" width="96" />