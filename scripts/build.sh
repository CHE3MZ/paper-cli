#!/usr/bin/env sh
# Build paper from the path in current-build-version.json.
# 1. Runs the embed generator (reads current-build-path).
# 2. Builds ./cmd/paper into build/paper.
set -eu
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

echo "==> regenerating embed from current-build-version.json"
go run ./src/genembed

mkdir -p "$ROOT/build"
OUT="$ROOT/build/paper"

# Bake the CLI version into the binary. $PAPER_CLI_VERSION wins (the
# release workflow sets it to the tag being released, so no network call
# happens there, and CI sets it to "dev" so test artifacts never claim a
# release version); otherwise ask GitHub what the latest release is (local
# git tags can be stale or unpushed, so they are only a fallback);
# otherwise "dev" (plain `go build` without this flag does that). A dirty
# tree appends "-dirty" so dev builds can't masquerade as releases.
VERSION="${PAPER_CLI_VERSION:-}"
if [ -z "$VERSION" ]; then
  LATEST_JSON="$(curl -fsSL --max-time 10 https://api.github.com/repos/CHE3MZ/paper-cli/releases/latest 2>/dev/null || true)"
  if [ -n "$LATEST_JSON" ]; then
    VERSION="$(printf '%s' "$LATEST_JSON" | sed -n 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' | head -n 1)"
  fi
fi
if [ -z "$VERSION" ]; then
  VERSION="$(git describe --tags --abbrev=0 2>/dev/null || true)"
fi
if [ -z "$VERSION" ]; then
  VERSION="dev"
fi
# A binary built from a modified tree is not the release its tag names:
# mark it so `paper version` stays honest and `paper update` doesn't
# wrongly report "already up to date". (Build outputs are gitignored, so a
# clean checkout is unaffected.)
if [ "$VERSION" != "dev" ]; then
  DIRTY="$(git status --porcelain 2>/dev/null || true)"
  if [ -n "$DIRTY" ]; then
    VERSION="${VERSION}-dirty"
  fi
fi
echo "==> building $OUT (Paper CLI version $VERSION)"
go build -ldflags "-X github.com/CHE3MZ/paper-cli/src.CLIVersion=$VERSION" -o "$OUT" ./cmd/paper
echo "done: $OUT"
