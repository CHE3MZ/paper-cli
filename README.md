# Paper CLI

Simple offline-first PaperMC server deployer. The `paper` binary bundles a
Paper jar (see `current-build-version.json` → `current-build-path`) and
deploys it with `paper new`.

## Build

```sh
# Windows (PowerShell)
powershell -ExecutionPolicy Bypass -File scripts/build.ps1
# Linux/macOS
sh scripts/build.sh
```

The build reads `current-build-version.json`, stages that version's
`paper.jar` (+ `eula.txt` / `server.properties` when present) via
`go run ./src/genembed`, then outputs `build/paper` (`build/paper.exe`).

## Usage

```sh
paper help
paper new .                  # new server here
paper new ./my-server        # new server at PATH
paper run                    # java -jar paper.jar in current dir
paper run ./my-server --nogui -m=2gb
paper delete                 # delete everything but paper.jar (asks first)
paper delete ./my-server --confirm
```

Run flags: `--nogui`, `--memory=4gb` / `-m=512mb`, `--java=<path>`,
`--dry-run`. Delete flag: `--confirm` / `-y`.

## Layout

- `cmd/paper/main.go` — thin CLI entrypoint only
- `src/` — all logic (`cli.go`, `server.go`, `java.go`, `version.go`)
- `src/genembed/` — build-time generator (JSON → staged embed)
- `test/` — `go test ./test/...`
- `build/` — build output (gitignored)
