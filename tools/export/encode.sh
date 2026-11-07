#!/usr/bin/env bash
# Encodes the exported WAVs to Ogg Vorbis for assets/audio.
#   tools/export/encode.sh <wavdir>
set -euo pipefail
src=${1:-audio-wav}; dst=$(dirname "$0")/../../assets/audio
mkdir -p "$dst"
for f in "$src"/*.wav; do
  n=$(basename "$f" .wav); q=4; case $n in tune-*) q=3;; esac
  ffmpeg -loglevel error -y -i "$f" -c:a libvorbis -q:a $q "$dst/$n.ogg"
done
cp "$src/tunes.json" "$dst/tunes.json"
