#!/usr/bin/env bash

set -euo pipefail

REPO="EduardoNGomes/mdo"
REPO_URL="https://github.com/$REPO"
BIN="mdo"
INSTALL_PATH="/usr/local/bin/$BIN"

echo "------------------------------------------------"
echo "Installing mdo"
echo
echo "mdo is an open-source Markdown reader."
echo "Source code and documentation:"
echo "$REPO_URL"
echo
echo "This binary was built from the official GitHub"
echo "repository and distributed via GitHub Releases."
echo
echo "macOS may block binaries downloaded from the"
echo "internet. To ensure proper execution, this"
echo "installer will remove the quarantine attribute."
echo "------------------------------------------------"
echo

ARCH="$(uname -m)"
case "$ARCH" in
    x86_64 | amd64)
        GOARCH="amd64"
        ;;
    arm64 | aarch64)
        GOARCH="arm64"
        ;;
    *)
        echo "Unsupported macOS architecture: $ARCH" >&2
        exit 1
        ;;
esac

FILE="mdo-darwin-$GOARCH.tar.gz"
URL="https://github.com/$REPO/releases/latest/download/$FILE"
WORKDIR="$(mktemp -d)"

cleanup() {
    rm -rf "$WORKDIR"
}
trap cleanup EXIT

echo "Downloading $FILE..."
curl -fL "$URL" -o "$WORKDIR/$FILE"

echo "Extracting $FILE..."
tar -xzf "$WORKDIR/$FILE" -C "$WORKDIR"

if [ ! -f "$INSTALL_PATH" ]; then
    echo "Installing $BIN to /usr/local/bin..."
else
    echo "Updating existing $BIN binary..."
fi

sudo install -m 755 "$WORKDIR/$BIN" "$INSTALL_PATH"

echo "Removing macOS quarantine attribute..."
sudo xattr -dr com.apple.quarantine "$INSTALL_PATH" || true

echo "Installation complete!"
echo "Run 'mdo' to get started."
echo "Docs and source: $REPO_URL"
