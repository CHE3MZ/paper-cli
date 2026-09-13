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
echo "==> building $OUT"
go build -o "$OUT" ./cmd/paper
echo "done: $OUT"
