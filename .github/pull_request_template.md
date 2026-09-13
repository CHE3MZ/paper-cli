## What does this do?

<!-- One or two sentences. Link the issue it closes, if any: Closes #123. -->

## Checklist

- [ ] `go test ./test/...` passes
- [ ] Rebuilt once (`scripts/build.ps1` or `scripts/build.sh`) — the embed step works
- [ ] Help text (`src/cli.go`), tests, and docs Home page updated (if flags/commands changed)
- [ ] No new dependencies (stdlib only, unless discussed in the issue)
- [ ] CLI logic lives in `src/`, not `cmd/paper/main.go`
- [ ] `NOTICE.md` updated (only if bundled content changed)

## Notes for reviewers

<!-- Anything surprising, tradeoffs you made, or things you deliberately left out. -->
