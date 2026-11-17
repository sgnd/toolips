# Code layout

| Package | Role |
|---|---|
| `cmd/toolips` | `main` → `cli.Execute()` |
| `internal/cli` | Cobra commands, flags, dry-run, printing |
| `internal/out` | Envelope, errors, exit codes |
| `internal/pdfx` | pdfcpu wrappers |
| `internal/imgx` | stdlib + `x/image` transforms |
| `internal/secx` | TLS, headers, secret scan, CT |

Rules:

- Built in only; pure Go (`CGO_ENABLED=0`).
- Open-source deps: MIT, BSD, ISC, Apache.
- Work logic stays out of `cli`; `cli` only wires flags and `out.P`.
- Functions ≤100 lines, files ≤800 where practical.
