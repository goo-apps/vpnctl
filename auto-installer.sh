#!/bin/bash
set -euo pipefail

REPO="goo-apps/vpnctl"
GITHUB_API="https://api.github.com/repos/$REPO/releases/latest"

CURRENT_VERSION="${1:-unknown}"
echo "🔧 Running installer (current version: $CURRENT_VERSION)"

OS="$(uname | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$ARCH" in
  x86_64) ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *) echo "❌ Unsupported architecture: $ARCH"; exit 1 ;;
esac

echo "🖥️  Detected system: OS=$OS, ARCH=$ARCH"
echo "📡 Fetching latest release..."

RELEASE_JSON=$(curl -fsSL "$GITHUB_API")

VERSION=$(echo "$RELEASE_JSON" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
ASSET_URL=$(echo "$RELEASE_JSON" | grep "browser_download_url" | grep "${OS}_${ARCH}\\.zip" | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$ASSET_URL" ]; then
  echo "❌ No release asset found for ${OS}_${ARCH}"
  exit 1
fi

echo "📦 Latest version: $VERSION"
echo "🌐 Download URL: $ASSET_URL"

TMP_DIR=$(mktemp -d)
ZIP_FILE="$TMP_DIR/vpnctl.zip"

echo "⬇️  Downloading..."
curl -fsSL "$ASSET_URL" -o "$ZIP_FILE"

echo "📂 Extracting..."
unzip -q "$ZIP_FILE" -d "$TMP_DIR"

BIN_PATH=$(find "$TMP_DIR" -type f -name "vpnctl*" | head -n1)
mv "$BIN_PATH" "$TMP_DIR/vpnctl"
chmod +x "$TMP_DIR/vpnctl"

read -p "⚠️  Install vpnctl to /usr/local/bin/ ? (y/n): " CONFIRM
if [[ "$CONFIRM" != "y" && "$CONFIRM" != "Y" ]]; then
  echo "❌ Installation aborted by user."
  rm -rf "$TMP_DIR"
  exit 1
fi

echo "🚀 Installing to /usr/local/bin..."
sudo mv "$TMP_DIR/vpnctl" /usr/local/bin/vpnctl

rm -rf "$TMP_DIR"

echo ""
