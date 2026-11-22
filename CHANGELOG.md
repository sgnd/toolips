# Changelog

All notable changes to toolips are listed here.

## v0.1.0 - 2026-11-22

### Added

- `dns check`: A, AAAA, MX, SPF and DMARC records of a domain, with missing MX, SPF or DMARC, several SPF records and DMARC `p=none` flagged. `--server` asks one DNS server instead of the system's.
- `hash sum` and `hash verify`: SHA-256, SHA-512, SHA-1 or MD5 checksums of files, and a check of a sha256sum-style list such as a release's `checksums.txt` (`HASH_MISMATCH` when a file changed or is missing).
- `jwt decode`: a token's header and claims, with `exp`, `iat` and `nbf` as RFC 3339 times and an `expired` flag. The token comes from a file or stdin, and the signature is not checked.
- `pdf nup` and `pdf booklet`: several pages per sheet (`--per-sheet`), or a booklet to print and fold, on A4 or Letter (`--paper`).
- `qr make`: a QR code PNG from a text or URL (or `-` for stdin), with `--level` (L, M, Q, H) and `--scale` (the image may be at most 8192 pixels on a side). Pure Go, using `rsc.io/qr` (BSD-3-Clause).
- `guide install`: adds the toolips skill for Claude Code (`~/.claude/skills`) and for Codex, Cursor and Gemini CLI (`~/.agents/skills`). `--project` installs into the current folder, `--dry-run` shows the plan, and a changed file is kept as `.bak`. It writes those two files and sends nothing anywhere.
- `skills/toolips/SKILL.md`, `AGENTS.md` and the README list exactly the commands the binary has; a test fails when they drift.

### Fixed

- PDF commands refuse files over 512 MiB and streams that decode to more than 256 MiB (`PDF_TOO_LARGE`), so a small PDF bomb cannot exhaust memory.

## v0.0.1 - 2026-11-18

The first release.

### Added

- PDF: `merge`, `split`, `compress`, `rotate`, `watermark`, `from-image`,
  `to-image` (extracts embedded images; it does not render pages), `info`,
  `pages` (`--keep` extracts and reorders, `--remove` deletes), `number`
  (page numbers with `%p` and `%P`), and `protect` / `unlock` (AES-256, the
  password read from `--password-file` or stdin, never from a flag value).
- Images: `compress`, `resize`, `convert`, `crop`, `rotate`, `watermark`,
  `grayscale` and `strip`. They read JPG, PNG, GIF, WebP, BMP and TIFF, and
  write JPG, PNG and GIF; WebP, BMP and TIFF input is written as PNG by
  default. `strip` removes EXIF, GPS, XMP, IPTC and comments from JPG and
  text, eXIf and time chunks from PNG without re-encoding, and keeps a JPG's
  orientation. They take several files, a folder or a quoted glob.
- Security: `tls check`, `headers check`, `secret scan` and `ct search`.
  `secret scan` redacts what it finds and lists every file it did not read
  in full under `skipped`.
- One JSON envelope for agents and text for people (`-o text|json|auto`),
  with exit codes 0 ok, 1 error, 2 usage, 3 not found, 4 network, and a
  hint on every error.
- `--dry-run` on every write. An existing output file stops the command
  with `EXISTS` unless `--force` is given; a batch checks every output
  before writing any.
- Private by default: PDF, image and secret commands never use the network
  (pdfcpu runs offline). Only `tls`, `headers` and `ct` connect, to the host
  you name, and help and `toolips list` mark them.
- Safe with untrusted files: images that declare more than 100 megapixels
  are refused before decoding, and file names and remote text are escaped
  before they reach the terminal.
- `toolips guide`, `toolips list`, shell `completion`, and an agent skill in
  `skills/toolips/SKILL.md`.
- Install script with checksum check, and release archives for Linux, macOS
  and Windows on amd64 and arm64.
