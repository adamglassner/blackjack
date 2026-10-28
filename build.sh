#!/usr/bin/env bash
# Builds Blackjack - The Gold Room for Windows.
# Output: dist/Blackjack/ with Blackjack.exe and its text files.
#
# The Windows icon/version resource is made by mkres, which lives in the
# Ironite Launcher source (il/ironite-cgc/tools/mkres). Point LAUNCHER at that
# folder. Without it, the exe still builds, just without the embedded icon.
set -euo pipefail
cd "$(dirname "$0")"
ROOT=$PWD
OUT=$ROOT/dist/Blackjack
B=$ROOT/build
TITLE="Blackjack"
DESC="Blackjack - The Gold Room"
VERSION=$(grep -m1 -oP '^## \[\K[0-9.]+' web/CHANGELOG.md)

rm -rf "$ROOT/dist" "$B" && mkdir -p "$OUT" "$B"
cp shell/main.go shell/platform_windows.go shell/platform_other.go "$B/"
cp -r web "$B/web"
printf 'module ironite/blackjack\n\ngo 1.24\n' > "$B/go.mod"

if [ -n "${LAUNCHER:-}" ] && [ -d "$LAUNCHER/tools/mkres" ]; then
  ( cd "$LAUNCHER/tools/mkres" && GOFLAGS=-mod=mod GOPROXY=off GOSUMDB=off go build -o "$B/mkres" . )
  "$B/mkres" -version "$VERSION" -icon icon-1024.png -out "$B/rsrc_windows_amd64.syso" \
    -name "$TITLE" -exe Blackjack.exe -internal Blackjack -identity Ironite.Blackjack -desc "$DESC"
else
  echo "LAUNCHER not set: building without the embedded icon/version resource." >&2
fi

( cd "$B" && GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -trimpath \
    -ldflags "-H windowsgui -s -w -X main.gameID=Blackjack -X \"main.title=$TITLE\" -X main.version=$VERSION" \
    -o "$OUT/Blackjack.exe" . )

crlf() { sed 's/\r$//; s/$/\r/' "$1" > "$2"; }
crlf web/CHANGELOG.md "$OUT/CHANGELOG.txt"
crlf web/fonts/OFL.txt "$OUT/OFL.txt"
for f in README CREDITS; do
  sed "s/@VERSION@/$VERSION/g" "$f.txt" > "$OUT/$f.tmp" && crlf "$OUT/$f.tmp" "$OUT/$f.txt" && rm "$OUT/$f.tmp"
done
printf '%s\r\n' "$VERSION" > "$OUT/VERSION"
echo "Blackjack $VERSION -> $OUT/Blackjack.exe"
