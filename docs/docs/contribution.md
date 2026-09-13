# Contributing

Thanks for stopping by. This is a small project and we'd like to keep it
that way — the guiding rule is *simple over clever*.

## How to help

- **Found a bug?** Open an issue with what you ran, what you expected, and
  what happened. The output of `paper run --dry-run` is often useful.
- **Want a feature?** Open an issue first so we can agree it's in scope.
  Small, focused pull requests beat big ones.
- **Fixing docs?** Typos and unclear sentences are fair game, no issue
  needed.

## Before you open a pull request

1. Keep it stdlib-only — no new dependencies without a discussion first.
2. Keep CLI logic in `src/`, not in `cmd/paper/main.go`.
3. Run the tests: `go test ./test/...`
4. Rebuild once (`scripts/build.ps1` or `scripts/build.sh`) so you know
   the embed step still works.
5. If you changed flags or commands, update the help text in `src/cli.go`
   and the [Home](index.md) docs page to match.

That's it. No commit message police, no templates — just keep it simple
and make sure it builds.
