# CLI security review, 2026-11-18

Focused source review of `internal/cli`, `internal/pdfx`, `internal/imgx`
and `internal/secx` (plan item 1.9). Not a full audit: one reviewer, no
fuzzing.

## Trust model

The person or agent running toolips is trusted and names every input. The
untrusted parts are the content of input files (images, PDFs, scanned
files, their names) and network replies (`tls`, `headers`, `ct`).

## Fixed

| # | Issue | Fix |
|---|---|---|
| 1 | A small image could declare a huge size and exhaust memory on decode | `imgx.Open` reads the header first and refuses more than 100 megapixels |
| 2 | `secret scan` dropped every finding in a file once one line passed 1 MiB, and skipped files over 2 MiB without saying so; the result still read "no secrets found" | Findings before the long line are kept; every file not read in full is listed under `skipped` with the reason, in JSON and text |
| 3 | File names in `secret scan` text and in written-file lists reached the terminal unescaped; the escaper also passed carriage returns, which can overwrite a printed line | `out.SafeText` (moved from `secx`) escapes all control characters except tab and newline, and is applied to those names |
| 4 | pdfcpu can fetch URLs (link checks, remote images, revocation lists) | Always runs offline (plan item 1.8) |

## Checked, no change

| Area | Why it is fine |
|---|---|
| Network targets in `tls`, `headers`, `ct` | The user names the host, so reaching it is the intended action, not SSRF. Redirects stop at 5; `ct` reads at most 8 MiB, `headers` at most 64 KiB |
| Header, certificate and CT text | Escaped with `SafeText`; JSON output is escaped by the encoder |
| Secret snippets | The match is replaced with `[redacted]`; at most 24 characters of context each side |
| Globs and folders for images | Only `.jpg`, `.jpeg`, `.png`, `.gif` names are kept; folders are not walked recursively |

## Open

| Item | Status |
|---|---|
| PDF decompression bombs: pdfcpu has no size cap we can set; needs a fuzz run in a sandbox | needs validation, plan item 1.13 |
| Writes replace an existing file without warning (e.g. default `merged.pdf`) | plan item 1.12 |
