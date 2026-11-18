# toolips

[![CI](https://github.com/sgnd/toolips/actions/workflows/ci.yml/badge.svg)](https://github.com/sgnd/toolips/actions/workflows/ci.yml)
[![License: AGPL-3.0](https://img.shields.io/badge/License-AGPL--3.0-blue.svg)](LICENSE)
[![efolusi.com](https://img.shields.io/badge/web-efolusi.com-0a7ea4)](https://efolusi.com)

**All-in-one everyday tools for people and AI agents.**

Local, unlimited, AGPL. One static binary — no accounts, no feature gates.
JSON for agents, text for humans. Web: [efolusi.com](https://efolusi.com).

## Install

```sh
curl -fsSL https://raw.githubusercontent.com/sgnd/toolips/main/install.sh | sh
```

Or build from source (Go 1.22+):

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
| `image compress / resize / convert / crop / rotate / watermark / grayscale` | Read JPG, PNG, GIF, WebP, BMP, TIFF; write JPG, PNG, GIF (multi-file, folder, or glob) |
| `tls check` | Expiry, chain, protocol |
| `headers check` | Security header grade |
| `secret scan` | Local regex for keys/tokens |
| `ct search` | Certificate transparency (crt.sh) |
| `version` / `list` / `guide` / `completion` | About toolips, shell completion |

Global flags: `-o text\|json\|auto`, `--pretty`, `--dry-run` (writes).
File output: `--out` / `--outdir`.

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
