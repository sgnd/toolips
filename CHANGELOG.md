# Changelog

All notable changes to toolips are listed here.

## [Unreleased]

### Fixed

- A missing input file now fails with `NOT_FOUND` (exit 3) in `pdf` and `secret scan`.
- A wrong number of arguments now fails with `USAGE` (exit 2) instead of exit 1.
- An unknown verb (e.g. `toolips pdf nope`) now fails with `USAGE` (exit 2) instead of printing help with exit 0.
- PDF errors show only the root cause, not pdfcpu's internal call chain.
- PDF commands run pdfcpu offline, so they never fetch links, images or revocation lists.

### Added

- `completion` is listed in `toolips list`.
- `toolips --help` and `toolips list` say which commands use the network (`tls`, `headers`, `ct`).

## [0.1.0] — 2026-11-17

### Added

- First release: PDF pack (`merge`, `split`, `compress`, `rotate`, `watermark`,
  `to-image` embedded extract, `from-image`), image pack (`compress`, `resize`,
  `convert`, `crop`, `rotate`, `watermark`, `grayscale`), security pack
  (`tls check`, `headers check`, `secret scan`, `ct search`).
- Stable JSON envelope (`-o text|json`), `--dry-run` on writes, install script
  and GitHub Actions CI/release.
