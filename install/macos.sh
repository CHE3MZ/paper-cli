#!/bin/sh
# Install paper for macOS into ~/.local/bin/paper
set -eu
REPO="CHE3MZ/paper-cli"
ASSET="paper-macos"
URL="https://github.com/$REPO/releases/latest/download/$ASSET"
DEST="$HOME/.local/bin/paper"

echo "Installing paper for macOS..."
echo "  from: $URL"
echo "  to:   $DEST"

mkdir -p "$(dirname "$DEST")"
TMP="$DEST.tmp.$$"
trap 'rm -f "$TMP"' EXIT INT TERM

if command -v curl >/dev/null 2>&1; then
  curl -fL -o "$TMP" "$URL"
elif command -v wget >/dev/null 2>&1; then
  wget -O "$TMP" "$URL"
else
  echo "error: need curl or wget to download $URL" >&2
  exit 1
fi

chmod +x "$TMP"
mv "$TMP" "$DEST"
trap - EXIT INT TERM

echo "Installed to $DEST"
case ":$PATH:" in
  *":$HOME/.local/bin:"*) ;;
  *) echo "NOTE: ~/.local/bin is not on your PATH. Add: export PATH=\"\$HOME/.local/bin:\$PATH\"" ;;
esac
echo "Test with: paper help"
