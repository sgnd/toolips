# Planned work

The one list of planned work for toolips (CLI, apps, cloud).
Deferred items stay here with the reason and what they wait on.

## Phase 1: core CLI solid

| # | Item | Status |
|---|---|---|
| 1.1 | Missing input file returns `NOT_FOUND` (exit 3) in every command | done |
| 1.2 | CLI tests for every PDF verb (write and `--dry-run`) | open |
| 1.3 | Error envelope tests: `USAGE` exit 2, network errors exit 4 | open |
| 1.4 | Raise unit coverage of `out`, `imgx`, `secx` to 75% or more | open |
| 1.5 | `completion` in `toolips list`, as the command standard says | open |
| 1.6 | Clean PDF error messages (no repeated `open ...: open ...`) | open |
| 1.7 | CI green on all six platforms; release dry run with `make dist` | open |

## Phase 2: paid plan

Starts only after phase 1 is done. Follows how komando's apps and cloud are built.

| # | Item | Status |
|---|---|---|
| 2.1 | Study komando-apps and komando-cloud; write the toolips design | open |
| 2.2 | `efolusi/toolips-apps`: Mac, iPhone, iPad app over the CLI core | open |
| 2.3 | `efolusi/toolips-cloud`: API, account, sync, billing, website | open |
