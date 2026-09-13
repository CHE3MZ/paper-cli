# Contributing to paper-cli

Thanks for stopping by. This is a small project on purpose, and the one
rule that matters is: **simple over clever**. If a change adds a dependency
to save two lines, we'd rather have the two lines.

## Ways to help

- **Report bugs** — use the bug report template so we get the details we
  need (what you ran, what you expected, what happened).
- **Suggest features** — open a feature request first so we can agree it's
  in scope before anyone writes code. Small, focused changes beat big ones.
- **Fix docs** — typos and unclear sentences are fair game, no issue needed.

## Ground rules

1. **Stdlib only.** No new dependencies without discussing it first. The
   project builds with the Go toolchain and nothing else.
2. **CLI logic lives in `src/`.** `cmd/paper/main.go` is an 8-line
   entrypoint — keep it that way.
3. **Match the existing style.** Handwritten flag parsing, white /
   light-blue / gray output, long flag with its shorthand next to it
   (`--memory, -m`). Look at neighboring code before inventing
   something new.
4. **Mind the licenses.** The binary bundles PaperMC (GPL-3.0) and Mojang
   files — see `NOTICE.md`. Don't add bundled content without updating it.

## Before you open a pull request

- [ ] On a fresh clone, run `go run ./src/genembed` first — nothing
      (`go build`, `go test`, `go vet`) compiles without the generated
      bundle it stages
- [ ] `go test ./test/...` passes
- [ ] You rebuilt once (`scripts/build.ps1` or `scripts/build.sh`) so the
      embed step (`genembed`) is happy with your changes
- [ ] New/changed flags or commands are in the help text (`src/cli.go`),
      covered by tests, and documented on the docs Home page
- [ ] `mkdocs build --strict` (in `docs/`) passes if you touched docs

There are no commit message rules and no PR rituals — just keep it simple
and make sure it builds.
