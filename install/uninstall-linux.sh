#!/bin/sh
# Uninstall paper for Linux from ~/.local/bin/paper
set -eu
TARGET="$HOME/.local/bin/paper"

echo "Uninstalling paper..."
echo "  from: $TARGET"

if [ -e "$TARGET" ]; then
  rm -f "$TARGET"
  echo "Removed $TARGET"
else
  echo "Nothing to do: $TARGET not found."
fi
