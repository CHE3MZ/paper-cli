# Paper CLI

Simple offline-first PaperMC server deployer. The `paper` binary bundles the
Paper template from `current-build-version.json` → `current-build-path` as a
single gzip-compressed tarball (`paper.jar`, `libraries/`, `cache/`,
`eula.txt`, `server.properties`, ...) and `paper new` extracts it, so new
servers start with no downloads. Server-regenerated output (`logs/`,
`versions/`, `plugins/.paper-remapped`) is excluded — a boot recreates it,
verified by test. Expect a ~177MB binary: the jars are already compressed,
so gzip only saves ~3%; the real saving (~60MB) is skipping regenerable
files. Compression is lossless (gzip CRC-checked): extraction is
bit-identical or fails loudly.

## Build

```sh
# Windows (PowerShell)
powershell -ExecutionPolicy Bypass -File scripts/build.ps1
# Linux/macOS
sh scripts/build.sh
```

The build reads `current-build-version.json`, packs that version's directory
(minus regenerable output) into a `.tar.gz` via `go run ./src/genembed`,
then outputs `build/paper` (`build/paper.exe`).

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
