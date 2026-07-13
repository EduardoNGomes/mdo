#!/usr/bin/env bash

set -euo pipefail

REPO="EduardoNGomes/mdo"
BIN="mdo"
INSTALL_PATH="/usr/local/bin/$BIN"

ARCH="$(uname -m)"
case "$ARCH" in
    x86_64 | amd64)
        GOARCH="amd64"
        ;;
    aarch64 | arm64)
        GOARCH="arm64"
        ;;
    *)
        echo "Unsupported Linux architecture: $ARCH" >&2
        exit 1
        ;;
esac

FILE="mdo-linux-$GOARCH.tar.gz"
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

echo "Installing $BIN to $INSTALL_PATH..."
sudo install -m 755 "$WORKDIR/$BIN" "$INSTALL_PATH"

echo "Installation complete!"
