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

test -z "$TMPDIR" && TMPDIR="$(mktemp -d)"
export TAR_FILE="$TMPDIR/${FILE_BASENAME}_$(uname -s)_$(uname -m).tar.gz"

(
	cd "$TMPDIR"
	echo "Downloading AssemblyAI CLI $VERSION..."
	curl -sfLo "$TAR_FILE" \
		"$RELEASES_URL/download/$VERSION/${FILE_BASENAME}_${VERSION:1}_$(uname -s)_$(uname -m).tar.gz"
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
	echo "export PATH=\"$BINARY_PATH:\$PATH\"" >> "$HOME/.zshrc"
fi
if [ -f "$HOME/.bashrc" ]; then
	echo "export PATH=\"$BINARY_PATH:\$PATH\"" >> "$HOME/.bashrc"
fi

"${BINARY_PATH}/${FILE_BASENAME}" welcome -i -o="$(uname -s)" -m="curl" -v="$VERSION" -a="$(uname -m)"

if [ -f "$HOME/.bashrc" ]; then
	source "$HOME/.bashrc"
fi
if [ -f "$HOME/.zshrc" ]; then
	zsh
	source "$HOME/.zshrc"
fi