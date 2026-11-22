# toolips

[![CI](https://github.com/sgnd/toolips/actions/workflows/ci.yml/badge.svg)](https://github.com/sgnd/toolips/actions/workflows/ci.yml)
[![License: AGPL-3.0](https://img.shields.io/badge/License-AGPL--3.0-blue.svg)](LICENSE)
[![efolusi.com](https://img.shields.io/badge/web-efolusi.com-0a7ea4)](https://efolusi.com)

**One local command line for your PDFs, images, QR codes, secret scans and TLS checks. For you and your AI agent.**

To merge two PDFs, shrink a photo or check a certificate, most people
upload the file to a website. toolips does the same jobs on your own
machine, with one command that you and an AI agent (Claude Code, Codex,
Cursor, Gemini CLI, your own bot) can both use:

- **Your files stay on your machine.** PDF, image and secret-scan commands
  never use the network. Only `tls`, `headers`, `ct` and `dns` connect, and only
  to the host you name.
- **Nothing is replaced by accident.** `--dry-run` shows what a command
  would write, and an existing file is kept unless you add `--force`.
- **Readable text for you, JSON for an agent**, with an exit code and a
  hint on every error.
- **One file to install**, with no account and no sign-up.

Web: [efolusi.com](https://efolusi.com).

## Install

```sh
curl -fsSL https://raw.githubusercontent.com/sgnd/toolips/main/install.sh | sh
```

Or build from source (the Go version in `go.mod`):

```sh
make build   # → bin/toolips
```

## First five minutes

```sh
toolips version
toolips list
toolips pdf merge a.pdf b.pdf --out out.pdf
toolips image resize photo.jpg --width 800 --out small.jpg
toolips tls check example.com
```

Agents: add `-o json` (or pipe stdout). Read `error.hint` on failure.
The agent skill is [skills/toolips/SKILL.md](skills/toolips/SKILL.md).

## Connect your AI agent

```sh
toolips guide install            # Claude Code, Codex, Cursor, Gemini CLI
toolips guide install --project  # only for the folder you are in
toolips guide install --dry-run  # show where it would write
```

It writes the toolips skill to `~/.claude/skills/toolips/SKILL.md` (Claude
Code) and `~/.agents/skills/toolips/SKILL.md` (Codex, Cursor, Gemini CLI and
others), and nothing else. It sends nothing anywhere. Run it again after an
update: a changed file is kept as `.bak`. Agents that read no skills can
use [AGENTS.md](AGENTS.md).

## Commands

| Command | What it does |
|---|---|
| `pdf merge / split / compress / rotate / watermark` | PDF editing (pdfcpu) |
| `pdf to-image` | Extract **embedded** images (not page rasterization) |
| `pdf from-image` | Build a PDF from images |
| `pdf info` | Pages, page size, metadata, encryption |
| `pdf protect / unlock` | Add or remove a password (AES-256) |
| `pdf pages` | Extract, reorder or delete pages |
| `pdf number` | Add page numbers |
| `pdf nup / booklet` | Several pages per sheet, or a booklet to print and fold |
| `image compress / resize / convert / crop / rotate / watermark / grayscale` | Read JPG, PNG, GIF, WebP, BMP, TIFF; write JPG, PNG, GIF (multi-file, folder, or glob) |
| `image strip` | Remove EXIF, GPS and other metadata from JPG and PNG, without re-encoding |
| `tls check` | Expiry, chain, protocol |
| `headers check` | Security header grade |
| `secret scan` | Local regex for keys/tokens |
| `ct search` | Certificate transparency (crt.sh) |
| `dns check` | A, AAAA, MX, SPF and DMARC records, with mail problems flagged |
| `hash sum / verify` | SHA-256 (or SHA-512, SHA-1, MD5) of files; check a `checksums.txt` |
| `qr make` | A QR code PNG from a text or URL (pure Go; `--level`, `--scale`) |
| `guide install` | Teach Claude Code, Codex, Cursor and Gemini CLI to use toolips |
| `jwt decode` | A token's header and claims, with readable times; the token is read from a file or stdin |
| `version` / `list` / `guide` / `completion` | About toolips, the short playbook, shell completion |

Global flags: `-o text\|json\|auto`, `--pretty`, `--dry-run` (writes).
File output: `--out` / `--outdir`.

## Size and memory

Measured on a Mac with Apple silicon, release build (`make build`,
`-s -w`), `/usr/bin/time -l`:

| What | Result |
|---|---|
| Binary | 13 MB |
| `pdf merge` of five PDFs, 53 MB in all | 72 MB peak memory, 0.05 s |
| `pdf compress` of that 53 MB PDF | 72 MB peak memory, 0.05 s |

The test PDFs hold random image data, so compress cannot shrink them.
Your files will differ.

## Build and test

```sh
CGO_ENABLED=0 go build -o bin/toolips ./cmd/toolips
go test ./...
make lint
make dist    # linux/darwin/windows amd64+arm64 archives + checksums
```

## License

AGPL-3.0 for open use; a [commercial license](COMMERCIAL-LICENSE.md) is
available when the AGPL does not fit. © Sugeng Agung Suganda.
Contact: sugandaletters@gmail.com.

See [CONTRIBUTING.md](CONTRIBUTING.md), [CLA.md](CLA.md), [SECURITY.md](SECURITY.md).
