#!/bin/sh
set -e

echo "Using the OSS distribution..."
RELEASES_URL="https://github.com/AssemblyAI/assemblyai-cli/releases"
FILE_BASENAME="assemblyai"

test -z "$VERSION" && VERSION="$(curl -sfL -o /dev/null -w %{url_effective} "$RELEASES_URL/latest" |
		rev |
		cut -f1 -d'/'|
		rev)"

test -z "$VERSION" && {
	echo "Unable to get assemblyai version." >&2
	exit 1
}

OS=$(uname | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case $ARCH in
	armv6*) ARCH="armv6" ;;
	armv7*) ARCH="armv7" ;;
	aarch64) ARCH="arm64" ;;
	x86) ARCH="386" ;;
	x86_64) ARCH="amd64" ;;
	i686) ARCH="386" ;;
	i386) ARCH="386" ;;
	arm64) ARCH="arm64" ;;
	amd64) ARCH="amd64" ;;
esac

test -z "$TMPDIR" && TMPDIR="$(mktemp -d)"
export TAR_FILE="$TMPDIR/${FILE_BASENAME}_${OS}_${ARCH}.tar.gz"

(
	cd "$TMPDIR"
	echo "Downloading AssemblyAI CLI $VERSION..."
	curl -sfLo "$TAR_FILE" \
		"$RELEASES_URL/download/$VERSION/${FILE_BASENAME}_${VERSION:1}_${OS}_${ARCH}.tar.gz"
)

BINARY_PATH="$HOME/.assemblyai-cli"
mkdir -p "$BINARY_PATH"
echo "Installing AssemblyAI to $BINARY_PATH"
tar -xzf "$TAR_FILE" -C "$BINARY_PATH"
chmod +x "$BINARY_PATH"
echo "$BINARY_PATH" >> "$HOME/.assemblyai"

if [ ! -f "$HOME/.bashrc" ]; then
	touch "$HOME/.bashrc"
fi

if [ ! -f "$HOME/.zshrc" ]; then
	touch "$HOME/.zshrc"
fi

if [ -f "$HOME/.zshrc" ]; then
	echo "" >> "$HOME/.zshrc"
	echo "export PATH=\"$BINARY_PATH:\$PATH\"" >> "$HOME/.zshrc"
fi
if [ -f "$HOME/.bashrc" ]; then
	echo "" >> "$HOME/.bashrc"
	echo "export PATH=\"$BINARY_PATH:\$PATH\"" >> "$HOME/.bashrc"
fi

"${BINARY_PATH}/${FILE_BASENAME}" welcome -i -o="$OS" -m="curl" -v="$VERSION" -a="$ARCH"

if [ -f "$HOME/.bashrc" ]; then
	source "$HOME/.bashrc"
fi
if [ -f "$HOME/.zshrc" ]; then
	zsh
	source "$HOME/.zshrc"
fi