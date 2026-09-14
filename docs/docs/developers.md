# Developers

This page is for people who want to read or change the source. The project
is deliberately boring: standard library only, no frameworks, no code
generation at runtime. If you know a little Go you can hold the whole thing
in your head.

## Project structure

```
paper-cli/
├── cmd/paper/main.go      # CLI entrypoint (8 lines, nothing else lives here)
├── src/                   # all application logic
│   ├── cli.go             # command dispatch, flag parsing, help text
│   ├── server.go          # new / run / delete implementation
│   ├── java.go            # memory parsing, java resolution, command building
│   ├── version.go         # reading the embedded bundle out of the binary
│   ├── update.go          # paper version / paper update (self-update, CLI version)
│   ├── progress.go        # download progress bar (terminal only, silent when piped)
│   ├── detach_unix.go     # helper detach: new session so signals miss it
│   ├── detach_windows.go  # helper detach: no-op (children outlive parents)
│   ├── style.go           # ANSI colors (white, light blue, gray)
│   ├── console_windows.go # enables ANSI colors on Windows consoles
│   ├── console_other.go   # no-op on other platforms
│   ├── embed_generated.go # generated: what got embedded at build time
│   └── genembed/main.go   # build-time generator (JSON -> bundle -> embed)
├── test/                  # go test ./test/...
├── paper-server/1.21.11/  # the Paper template that gets bundled
├── scripts/
│   ├── build.ps1 / build.sh   # the build (both do the same thing)
│   └── pull-paper-jar.sh      # downloads a fresh Paper + trims it
├── current-build-version.json # which template version the binary uses
└── docs/                  # these docs (mkdocs project)
```

The rule is simple: `cmd/paper/main.go` only calls into `src`. Everything
else lives in `src`.

## How the build works

`go:embed` needs its file list at compile time, but which Paper version to
embed is decided by `current-build-version.json` (`current-build-path`).
The build scripts bridge that gap in two steps:

1. `go run ./src/genembed` reads the JSON, packs that template directory
   into `src/bundled/bundle.tar.gz` (gzip, best compression), and writes
   `src/embed_generated.go` with the `//go:embed` line plus some constants
   (build path, version, file count, sizes).
2. `go build -o build/paper ./cmd/paper` compiles the binary with the
   archive inside it. The build scripts add `-ldflags "-X
   github.com/CHE3MZ/paper-cli/src.CLIVersion=<tag>"` so `paper version`
   reports the release tag (`dev` when unstamped).

Not everything in the template is packed: `versions/`, `logs/`, and
`plugins/.paper-remapped/` are skipped because the server recreates them
on boot (we've booted a fresh deploy to prove it). Line endings are pinned
to LF by `.gitattributes`, so the packed bytes are identical on every OS —
a CRLF checkout would silently change the bundle and flip release builds
to `-dirty`. The jars themselves are
already compressed, so gzip only saves a few percent — the real saving is
skipping those regenerable files.

At runtime, `paper new` gunzips the archive and untars it into the target
folder. Decompression is lossless and CRC-checked: a corrupt bundle errors
out instead of deploying half-broken files, and tar entries are validated
so nothing can escape the target directory.

## What each file does

**cli.go** — Parses `os.Args` and routes to `runNew`, `runRun`, or
`runDelete`. Flag parsing is handwritten (no CLI library): it accepts both
`--memory=4gb` and `--memory 4gb` forms, plus attached shorthands like
`-m4gb`. Also owns the help text, so `paper help` always matches the real
flags. Colors go through the helpers in `style.go`, which turn themselves
off when output is piped.

**server.go** — The server commands (new / run / delete):

- `NewServer` refuses if `paper.jar` already exists (unless `--force`),
  then extracts the whole bundle and guarantees an `eula.txt` exists.
- `RunServer` checks for `paper.jar`, builds the java invocation, and
  execs it with the server folder as working directory and your terminal
  attached. `--dry-run` prints the command instead.
- `DeleteServer` lists everything except `paper.jar`, asks `[y/N]` unless
  `--confirm`, and refuses entirely if there's no `paper.jar` (so you
  can't nuke a random folder by mistake).

**java.go** — `NormalizeMemory` turns `4gb`, `512mb`, `1024`, `2G` into
Java-style `4G` / `512M` / `1024M` / `2G` (bare numbers mean megabytes).
`BuildJavaCommand` assembles the final invocation: optional `-Xms`/`-Xmx`,
then `-jar paper.jar`, then `nogui`. `--optimized` just fills in `4gb`
when no explicit memory was given — explicit flags always win.

**version.go** — `ExtractBundled` (tar.gz extraction with the safety checks
above). The `Embedded*` constants next to it record what the binary was
built with (build path, version, file count, sizes).

**update.go** — `paper version` and `paper update`. `CLIVersion` holds the
CLI's own release version: it defaults to `dev` and is stamped at build
time with `-ldflags "-X
github.com/CHE3MZ/paper-cli/src.CLIVersion=<tag>"`. `scripts/build.sh`
and `scripts/build.ps1` do this for you (they use `$PAPER_CLI_VERSION`
when set — the `release-all` workflow sets it to the tag being released,
CI sets it to `dev` so test artifacts never claim a release — else the
latest GitHub release tag (what
`github.com/CHE3MZ/paper-cli/releases/latest` points at), else the latest
local git tag, else `dev`). Local tags are only a fallback because they
can be stale or never pushed. A dirty tree appends `-dirty`, so a dev
build can't masquerade as a release (and `paper update` won't wrongly call
it up to date). `VersionText` prints the bundled
PaperMC version next to it. The updater mirrors the install scripts
(same repo, same `releases/latest/download` URL, per-OS asset names) but
runs from inside the CLI: it downloads into a `paper_temp.<pid>` file next to
the running binary (`UpdateTempPath` — PID-suffixed so concurrent updates
don't share a file), verifies its SHA256 against the
digest published with the release (`VerifyFileSHA256`, from
`FetchReleaseInfo`), then swaps it over the binary. The swap runs in a
helper, which is just this binary re-executed as the hidden
`__finish-update` command (`spawnFinishHelper`, `runFinishUpdate`) — so
the install stays one file and there is no generated script to tamper
with. The swap (`SwapStaged`) tries a direct rename, then the rename-aside
swap against a still-running exe — deliberately no wait-and-retry loop and
no Windows API: on Windows the helper itself runs from dest, so a direct
rename can never succeed while it lives, while the aside swap works
immediately. The
parent waits for the helper everywhere except Windows, where it must exit
first and hands off instead. `detach_unix.go` / `detach_windows.go` let
the helper outlive terminal signals. Leftovers (.old, orphaned staging
files) are swept best-effort on version/update invocations
(`SweepUpdateLeftovers`). The target is the resolved current
executable (`UpdateTarget`), falling back to `~/.local/bin/paper`.

**progress.go** — `ProgressWriter`, an `io.Writer` wrapper that counts
download bytes through to the file while drawing a single-line `\r` bar
(percent, megabytes, speed, ETA) on interactive terminals only — piped
output stays byte-clean. The line renderer (`DownloadLine`) is a pure
function so tests pin the exact format.

**style.go, console_windows.go, console_other.go** — Bold/white/light
blue/gray/green/red paint functions. Windows consoles need virtual-terminal
processing switched on before ANSI codes work, which is what the
`console_*.go` pair does (stdlib `syscall` on Windows, no-op elsewhere).

**src/genembed/main.go** — Described in "How the build works" above. Run
it by hand any time with `go run ./src/genembed`.

## Common recipes

**Add a run flag.** Add the field to `RunOptions` in `java.go`, handle it
in `BuildJavaCommand`, parse it in `runRun` in `cli.go`, document it in
`HelpText` and on the Home docs page, and cover it in `test/java_test.go`.

**Add a command.** Add a `runX` function in `cli.go`, a case in the `Run`
switch, the implementation in `server.go` (or `update.go` for
self-management commands), help text, docs, and tests.

**Bump the Paper version.** Edit `VERSION` in `scripts/pull-paper-jar.sh`
and run it (needs `jq`, `curl`, `java`), point `current-build-path` in
`current-build-version.json` at the new folder, then rebuild. `genembed`
will fail loudly if `paper.jar` is missing, so a typo can't silently ship
an empty bundle.

**Change the color scheme.** It's all in `src/style.go` — the ANSI codes
at the top plus the six one-line helpers.

## Testing

```
go run ./src/genembed   # required once per fresh clone: stages the bundle
go test ./test/...      # without it, nothing compiles (go:embed needs the file)
```

Tests live in `test/` and cover memory parsing, command building, and a
full `new` → `delete` round trip in temp dirs (including a check that the
deployed file count matches the embedded count). Run them before opening a
pull request, and rebuild once to make sure `genembed` is happy with your
changes.
