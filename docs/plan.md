# Planned work

The one list of planned work for toolips (CLI, apps, cloud).
Deferred items stay here with the reason and what they wait on.

## Goal

The local, private version of the online file and web tools people use every
day (iLovePDF, TinyPNG, SSL Labs, securityheaders.com, crt.sh), for people and
AI agents. The product goal and philosophy are in brainOS `AGENTS.md`
(Standards); each phase below names the items that serve them.

## Phase 1: core CLI solid

| # | Item | Status |
|---|---|---|
| 1.1 | Missing input file returns `NOT_FOUND` (exit 3) in every command | done |
| 1.2 | CLI tests for every PDF verb (write and `--dry-run`) | done |
| 1.3 | Error envelope tests: `USAGE` exit 2, network errors exit 4 | done |
| 1.4 | Raise unit coverage of `out`, `imgx`, `secx` to 75% or more | done |
| 1.5 | `completion` in `toolips list`, as the command standard says | done |
| 1.6 | Clean PDF error messages (no repeated `open ...: open ...`) | done |
| 1.7 | CI green on all six platforms; release dry run with `make dist` (checked 2026-11-18 at 74f21da: CI green on all six; archive names, `checksums.txt` and the archive layout match `install.sh` and `release.yml`) | done |
| 1.8 | Private: files never leave the machine; only `tls`, `headers` and `ct` use the network, stated in `--help` and tested | done |
| 1.9 | Secure: security review of the CLI (paths, archives, network inputs) with the `security-audit` skill; record in `docs/security/2026-11-18-cli-review.md` | done |
| 1.10 | For agents: agent skill file (`skills/toolips/SKILL.md`) like komando's | done |
| 1.11 | Killer-feature gap list: compare with the top online tools and add the missing everyday verbs as plan items (the gap list below) | done |
| 1.12 | Refuse to replace an existing output file unless `--force` is given (from the security review) | done |
| 1.13 | Fuzz PDF input for decompression bombs in a sandbox (from the security review) | done: `pdfx.Limits` plus fuzz targets; run recorded in `docs/security/2026-11-18-cli-review.md` |

## Phase 1b: killer-feature gaps

From plan item 1.11: what the top online tools offer that toolips does not
yet. Every item must run locally in pure Go (no CGO). Priority: A first
(common need, library already in `go.mod`), then B, then C.

| # | Item | Online tool it replaces | Priority | Status |
|---|---|---|---|---|
| 1.14 | `pdf protect` / `pdf unlock` (password, pdfcpu `crypto`) | iLovePDF Protect, Unlock | A | done |
| 1.15 | `pdf pages`: extract, delete and reorder pages (pdfcpu `collect`, `trim`) | iLovePDF Organize, Remove pages | A | done |
| 1.16 | `pdf info`: pages, size, metadata, encryption (pdfcpu `info`) | many PDF sites | A | done |
| 1.17 | `pdf number`: add page numbers (pdfcpu `stamp`) | iLovePDF Page numbers | A | done |
| 1.18 | Read WebP input (`golang.org/x/image/webp`, decode only) and BMP/TIFF input | iLoveIMG, Squoosh | A | done |
| 1.19 | `image strip`: remove EXIF and GPS metadata | photo privacy sites | A | done |
| 1.20 | `dns check`: A, MX, SPF, DMARC, CAA records of a domain (network) | MXToolbox | B | done, without CAA: Go's resolver cannot ask for CAA; see 1.29 |
| 1.21 | `hash`: SHA-256 and other checksums of files, verify against a sum | online hash sites | B | done (`hash sum`, `hash verify`) |
| 1.22 | `qr make`: QR code PNG from text or a URL | QR code generators | B | open |
| 1.23 | `jwt decode`: show a token's header and claims, never send it | jwt.io | B | done |
| 1.24 | `pdf nup` / `pdf booklet`: several pages per sheet (pdfcpu `nup`, `booklet`) | iLovePDF | C | done |
| 1.25 | Write WebP output | Squoosh | C | deferred: needs a pure-Go WebP encoder; none in `golang.org/x/image` |
| 1.26 | Render PDF pages to images | iLovePDF PDF to JPG | C | deferred: needs a renderer, which means CGO (see AGENTS.md rule 4) |
| 1.27 | Office to PDF and PDF to Office | iLovePDF Word to PDF | C | deferred: no pure-Go converter; needs a design decision |
| 1.28 | Background removal, upscaling, OCR | remove.bg, iLoveIMG | C | deferred: needs ML models; fits the paid apps better |
| 1.29 | CAA records in `dns check` | MXToolbox | C | open: needs a DNS client library (for example `github.com/miekg/dns`), since Go's resolver has no CAA lookup |

## Phase 2: paid plan

Starts only after phase 1 is done. Follows how komando's apps and cloud are built.
toolips never builds login, accounts, plans, billing or checkout: they come
from `efolusi/sso` and `efolusi/pay`. Items that need them say "via sso/pay".

| # | Item | Status |
|---|---|---|
| 2.1 | Study komando-apps and komando-cloud; write the toolips design | open |
| 2.2 | `efolusi/toolips-apps`: Mac, iPhone, iPad app over the CLI core | open |
| 2.3 | `efolusi/toolips-cloud`: API, sync and website; account and billing via sso/pay | open |
| 2.4 | Self-hosted: the cloud service can run on the user's own server | open |
| 2.5 | Commercial: plans and pricing for apps and cloud, as in komando; plans, billing and checkout via sso/pay | open |
