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
	echo "Downloading AssemblyAI $VERSION..."
  echo "$RELEASES_URL/download/$VERSION/${FILE_BASENAME}_${VERSION:1}_$(uname -s)_$(uname -m).tar.gz"
	curl -sfLo "$TAR_FILE" \
		"$RELEASES_URL/download/$VERSION/${FILE_BASENAME}_${VERSION:1}_$(uname -s)_$(uname -m).tar.gz"
)

COPY_PATH="$(echo "$PATH" | cut -d: -f1)"
tar -xf "$TAR_FILE" -C "$TMPDIR"
mv "${TMPDIR}/${FILE_BASENAME}" $COPY_PATH

"${COPY_PATH}/${FILE_BASENAME}" "$@"

