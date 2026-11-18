# Changelog

All notable changes to toolips are listed here.

## [Unreleased]

### Fixed

- PDF commands refuse files over 512 MiB and streams that decode to more than 256 MiB (`PDF_TOO_LARGE`), so a small PDF bomb cannot exhaust memory.

## [0.0.1] - 2026-11-18

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
