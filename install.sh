#!/bin/sh
# Install the latest toolips release.
#
#   curl -fsSL https://raw.githubusercontent.com/sgnd/toolips/main/install.sh | sh
#
# Installs to /usr/local/bin when writable, otherwise ~/.local/bin.
# Override with TOOLIPS_INSTALL_DIR=/some/dir.
# Install one version with TOOLIPS_VERSION=v0.0.1.
set -eu

REPO="sgnd/toolips"

os=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$os" in
  linux | darwin) ;;
  mingw* | msys* | cygwin*)
    echo "This installer is for Linux and macOS. On Windows, download" >&2
    echo "toolips_windows_amd64.zip (or _arm64.zip) from https://github.com/$REPO/releases," >&2
    echo "or run this line inside WSL." >&2
    exit 1 ;;
  *) echo "toolips supports Linux, macOS and Windows; there is no release for $os." >&2; exit 1 ;;
esac

case "$(uname -m)" in
  x86_64 | amd64) arch=amd64 ;;
  arm64 | aarch64) arch=arm64 ;;
  *) echo "There is no toolips release for this CPU ($(uname -m)): amd64 and arm64 only." >&2; exit 1 ;;
esac
for tool in curl tar; do
  command -v "$tool" >/dev/null 2>&1 || { echo "The installer needs $tool. Install it, then run the line again." >&2; exit 1; }
done

if [ -n "${TOOLIPS_INSTALL_DIR:-}" ]; then
  dir="$TOOLIPS_INSTALL_DIR"
  mkdir -p "$dir" 2>/dev/null || true
  if [ ! -d "$dir" ] || [ ! -w "$dir" ]; then
    echo "TOOLIPS_INSTALL_DIR=$dir is not a folder you can write to." >&2
    echo "Choose another folder, or leave TOOLIPS_INSTALL_DIR out to use /usr/local/bin or ~/.local/bin." >&2
    exit 1
  fi
else
  dir="/usr/local/bin"
  if [ ! -w "$dir" ]; then
    dir="$HOME/.local/bin"
    mkdir -p "$dir"
  fi
fi

file="toolips_${os}_${arch}.tar.gz"
base="https://github.com/$REPO/releases/latest/download"
if [ -n "${TOOLIPS_VERSION:-}" ]; then
  case "$TOOLIPS_VERSION" in v*) v=$TOOLIPS_VERSION ;; *) v=v$TOOLIPS_VERSION ;; esac
  base="https://github.com/$REPO/releases/download/$v"
fi
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
prev=$(command -v toolips 2>/dev/null || true)

get() {
  if curl -fsSL "$1" -o "$2"; then return 0; fi
  if [ -n "${v:-}" ]; then
    echo "Could not download toolips $v. Is there such a version? See https://github.com/$REPO/releases" >&2
  else
    echo "Could not download toolips. Check the internet connection (and HTTPS_PROXY), then run the line again." >&2
  fi
  exit 1
}

echo "Downloading $file ..."
get "$base/$file" "$tmp/$file"
get "$base/checksums.txt" "$tmp/checksums.txt"

echo "Verifying checksum ..."
expected=$(grep " $file\$" "$tmp/checksums.txt" | cut -d' ' -f1)
if command -v sha256sum >/dev/null 2>&1; then
  actual=$(sha256sum "$tmp/$file" | cut -d' ' -f1)
else
  actual=$(shasum -a 256 "$tmp/$file" | cut -d' ' -f1)
fi
if [ -z "$expected" ] || [ "$expected" != "$actual" ]; then
  echo "The download does not match checksums.txt of its release; nothing was installed." >&2
  echo "Run the line again; if it happens again, report it at https://github.com/$REPO/issues" >&2
  exit 1
fi

tar --no-same-owner -xzf "$tmp/$file" -C "$tmp"

new="$dir/.toolips.new.$$"
trap 'rm -rf "$tmp"; rm -f "$new"' EXIT
cp "$tmp/toolips" "$new"
chmod 755 "$new"
mv -f "$new" "$dir/toolips"

echo "Installed: $dir/toolips"
case ":$PATH:" in
  *":$dir:"*) ;;
  *)
    case "${SHELL:-}" in
      */zsh) rc="$HOME/.zshrc" ;;
      */bash) if [ "$os" = darwin ]; then rc="$HOME/.bash_profile"; else rc="$HOME/.bashrc"; fi ;;
      *) rc="your shell's startup file" ;;
    esac
    echo
    echo "Your terminal does not look in $dir yet. Run this line now,"
    echo "and add it to $rc so it stays:"
    echo
    echo "  export PATH=\"$dir:\$PATH\""
    ;;
esac
hash -r 2>/dev/null || true
now=$(command -v toolips 2>/dev/null || true)
if [ -n "$now" ] && [ "$now" != "$dir/toolips" ]; then
  echo
  echo "Note: your terminal runs another toolips first: $now"
  echo "Remove that one, or install into its folder:"
  echo "  curl -fsSL https://raw.githubusercontent.com/$REPO/main/install.sh | TOOLIPS_INSTALL_DIR=${now%/*} sh"
fi
echo
if [ -n "$prev" ]; then
  echo "Updated. Try: toolips version"
else
  echo "Next:"
  echo "  toolips list"
  echo "  toolips guide"
fi
echo "Help: toolips --help, or https://github.com/$REPO"
