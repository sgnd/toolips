# Security

toolips edits local files (PDFs and images) and makes outbound HTTPS
requests for TLS, header and certificate-transparency checks. A security
bug can leak secrets from a scan or trust a bad TLS answer.

## Report a problem

Please don't open a public issue for a security problem. Report it privately instead:

1. Open the [Security tab](https://github.com/sgnd/toolips/security) of this repo.
2. Click **Report a vulnerability**.

Include what you found, how to reproduce it, and what an attacker could do
with it. You get an answer within a week. Once it is fixed, a new release
says what changed.

## What counts

- A way to make `secret scan` miss a clear credential pattern that the
  documented rules should catch, or to crash it on crafted input.
- Path traversal or unexpected overwrite from `--out` / `--outdir`.
- Accepting a TLS connection that Go's verifier would reject, without an
  explicit insecure flag (toolips has none on purpose).
- Anything that lets an AI agent read more than the user asked for from a
  file or network call.

A problem in pdfcpu, Go's TLS stack or crt.sh belongs to that project.

## How toolips keeps you safe

- Pure Go (`CGO_ENABLED=0`); no account, billing or feature gates in the CLI.
- Write commands support `--dry-run`.
- Output for agents is one JSON envelope (`-o json`); findings from
  `secret scan` are reported as snippets so you can fix them, not forwarded
  anywhere.
- If toolips later stores secrets on disk, they must be encrypted at rest
  and never printed in agent output (same spirit as komando’s vault).
- Release archives ship with `checksums.txt`. Prefer installing via
  [`install.sh`](install.sh), which verifies the checksum.

## Supported versions

Only the latest release gets security fixes. Older tags are unsupported.
