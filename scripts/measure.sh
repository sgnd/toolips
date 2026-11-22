#!/bin/sh
# Prints the binary size and the peak memory of `pdf merge` and `pdf compress`
# as a markdown table. The test PDFs hold random image data, so compress
# cannot shrink them. Usage: scripts/measure.sh ./bin/toolips
set -eu

bin=$(cd "$(dirname "$1")" && pwd)/$(basename "$1")
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

size=$(wc -c <"$bin" | tr -d ' ')

python3 - "$work" <<'PY'
import os, struct, sys, zlib
def png(path, w=1600, h=2200):
    raw = b"".join(b"\0" + os.urandom(w * 3) for _ in range(h))
    def chunk(t, d):
        c = struct.pack(">I", len(d)) + t + d
        return c + struct.pack(">I", zlib.crc32(t + d) & 0xFFFFFFFF)
    open(path, "wb").write(b"\x89PNG\r\n\x1a\n" + chunk(b"IHDR", struct.pack(">IIBBBBB", w, h, 8, 2, 0, 0, 0)) + chunk(b"IDAT", zlib.compress(raw, 1)) + chunk(b"IEND", b""))
for i in range(5):
    png(os.path.join(sys.argv[1], f"p{i}.png"))
PY

cd "$work"
for i in 0 1 2 3 4; do "$bin" pdf from-image "p$i.png" --out "p$i.pdf" -o json >/dev/null; done

# peak resident memory in MB of one command: GNU time (-v, kB) or BSD time (-l, bytes)
peak() {
  if /usr/bin/time -v true >/dev/null 2>&1; then
    /usr/bin/time -v "$@" 2>&1 >/dev/null | awk '/Maximum resident set size/ { printf "%d", $NF / 1024 }'
  else
    /usr/bin/time -l "$@" 2>&1 >/dev/null | awk '/maximum resident set size/ { printf "%d", $1 / 1048576 }'
  fi
}

in_bytes=$(cat p0.pdf p1.pdf p2.pdf p3.pdf p4.pdf | wc -c | tr -d ' ')
merge_mb=$(peak "$bin" pdf merge p0.pdf p1.pdf p2.pdf p3.pdf p4.pdf --out merged.pdf --force -o json)
compress_mb=$(peak "$bin" pdf compress merged.pdf --out small.pdf --force -o json)
[ -s merged.pdf ] && [ -s small.pdf ] || { echo "measure: merge or compress wrote nothing" >&2; exit 1; }

mb() { awk -v b="$1" 'BEGIN { printf "%.0f", b / 1048576 }'; }
echo "| What | Result |"
echo "|---|---|"
echo "| Binary ($(uname -s) $(uname -m)) | $(mb "$size") MB |"
echo "| \`pdf merge\` of five PDFs, $(mb "$in_bytes") MB in all | ${merge_mb} MB peak memory |"
echo "| \`pdf compress\` of the merged $(mb "$(wc -c <merged.pdf)") MB PDF | ${compress_mb} MB peak memory |"
