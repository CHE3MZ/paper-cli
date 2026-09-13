#!/usr/bin/env sh
set -eu

command -v jq >/dev/null 2>&1 || { echo "jq is required but not installed."; exit 1; }
command -v curl >/dev/null 2>&1 || { echo "curl is required but not installed."; exit 1; }

VERSION="1.21.11"
BASE_URL="https://fill.papermc.io/v3/projects/paper"
INSTALL_PATH="paper-server/$VERSION"

mkdir -p "$INSTALL_PATH"

LATEST_BUILD=$(curl -s "$BASE_URL/versions/$VERSION" | jq '.builds | max')
BUILD_JSON=$(curl -s "$BASE_URL/versions/$VERSION/builds/$LATEST_BUILD")
JAR_NAME=$(echo "$BUILD_JSON" | jq -r '.downloads.application.name')
DOWNLOAD_URL="$BASE_URL/versions/$VERSION/builds/$LATEST_BUILD/downloads/$JAR_NAME"
DESTINATION="$INSTALL_PATH/$JAR_NAME"

echo "Downloading $JAR_NAME (Build $LATEST_BUILD)..."
curl -sL -o "$DESTINATION" "$DOWNLOAD_URL"
echo "Saved to $DESTINATION"