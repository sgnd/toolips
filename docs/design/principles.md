# toolips design principles

In short: **one way to do each thing, words people understand, safe when
in doubt, and every error says what to do next.**

## What toolips is

toolips is a small CLI for everyday PDF, image and security checks. It is
for two users, equally:

- a person at a terminal (`-o text` or auto);
- an AI agent or script (`-o json`, never prompted).

There are no accounts, billing gates or feature flags in the CLI.

## Grammar

```
toolips <noun> <verb> [files...] [flags]
```

Nouns: `pdf`, `image`, `tls`, `headers`, `secret`, `ct`, plus `version`,
`list`, `guide`.

## Output

- One JSON envelope on stdout for agents (`contract: 1`).
- Text for humans. Global `-o` is the format (`text|json|auto`).
- File destinations use `--out` / `--outdir` (never overload `-o`).
- Exit codes: 0 ok, 1 error, 2 usage, 3 not found, 4 network.
- Write commands honour `--dry-run`.

## Technical rules

- Pure Go, `CGO_ENABLED=0`.
- Dependencies MIT / BSD / ISC / Apache only.
- Functions ≤100 lines, files ≤800 where practical.
- If toolips later persists secrets locally, encrypt at rest and never
  print values in envelopes (same spirit as komando’s vault/masking).
