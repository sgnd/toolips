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
| 5 | PDF bombs: no limit on input size, and each stream could decode to 512 MiB (pdfcpu's default) | `pdfx.Limits`: files over 512 MiB and streams over 256 MiB are refused with `PDF_TOO_LARGE` (plan item 1.13) |

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
| PDF decompression bombs | fixed (row 5). This review first said pdfcpu has no cap we can set; that was wrong: pdfcpu 0.16 has `ResourceLimits`, and toolips now sets them |
| Writes replace an existing file without warning (e.g. default `merged.pdf`) | fixed after the review: `--force` (plan item 1.12) |

## Fuzz run (plan item 1.13)

Run on 2026-11-18 on macOS arm64 with the targets in
`internal/pdfx/pdfx_internal_test.go`:

```sh
GOMEMLIMIT=1GiB go test ./internal/pdfx -run '^$' -fuzz '^FuzzReadInfo$' -fuzztime=60s -fuzzminimizetime=2s -parallel 4
GOMEMLIMIT=1GiB go test ./internal/pdfx -run '^$' -fuzz '^FuzzCompress$' -fuzztime=60s -fuzzminimizetime=2s -parallel 4
```

A watchdog stopped any run above 2 GB resident memory, and each input
failed after 5 seconds, so hangs count as findings.

| Target | Inputs run | Crashes | Hangs over 5 s | Peak memory |
|---|---|---|---|---|
| `FuzzReadInfo` | 1,732 | 0 | 0 | 565 MB |
| `FuzzCompress` | 904 | 0 | 0 | 510 MB |

The input count is low because pdfcpu reads from files. A longer run is
worth doing before 1.0.0.
