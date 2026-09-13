#!/usr/bin/env sh
set -eu

command -v jq >/dev/null 2>&1 || { echo "jq is required but not installed."; exit 1; }
command -v curl >/dev/null 2>&1 || { echo "curl is required but not installed."; exit 1; }

VERSION="1.21.11"
BASE_URL="https://fill.papermc.io/v3/projects/paper"
INSTALL_PATH="paper-server/$VERSION"

mkdir -p "$INSTALL_PATH"

echo "Fetching latest build number..."
LATEST_BUILD=$(curl -s "$BASE_URL/versions/$VERSION" | jq '.builds | max')
[ -n "$LATEST_BUILD" ] && [ "$LATEST_BUILD" != "null" ] || { echo "Failed to fetch latest build number."; exit 1; }
echo "Latest build: $LATEST_BUILD"

BUILD_JSON=$(curl -s "$BASE_URL/versions/$VERSION/builds/$LATEST_BUILD")

DOWNLOAD_INFO=$(echo "$BUILD_JSON" | jq -r '.downloads[] | select(.url != null and .name != null) | [.name, .url] | @tsv' | head -n 1)

if [ -z "$DOWNLOAD_INFO" ]; then
    echo "Error: Could not find valid download entry in build JSON."
    echo "$BUILD_JSON" | jq .downloads
    exit 1
fi

JAR_NAME=$(printf "%s" "$DOWNLOAD_INFO" | cut -f1)
DOWNLOAD_URL=$(printf "%s" "$DOWNLOAD_INFO" | cut -f2)
DESTINATION="$INSTALL_PATH/paper.jar"

echo "Downloading $JAR_NAME as paper.jar (Build $LATEST_BUILD)..."
curl -sL --fail -o "$DESTINATION" "$DOWNLOAD_URL"

if [ -s "$DESTINATION" ]; then
    echo "Success: Saved to $DESTINATION"
else
    echo "Error: Downloaded file is empty or missing."
    exit 1
fi

# Switch to install directory, generate initial files, write EULA, and launch server
cd "$INSTALL_PATH"

echo "Running initial server setup to generate files..."
java -jar paper.jar || true

echo "Configuring eula.txt..."
cat <<EOF > eula.txt
#By changing the setting below to TRUE you are indicating your agreement to our EULA (https://aka.ms/MinecraftEULA).
eula=true
EOF