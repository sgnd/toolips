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
| 1.5 | `completion` in `toolips list`, as the command standard says | open |
| 1.6 | Clean PDF error messages (no repeated `open ...: open ...`) | open |
| 1.7 | CI green on all six platforms; release dry run with `make dist` | open |
| 1.8 | Private: files never leave the machine; only `tls`, `headers` and `ct` use the network, stated in `--help` and tested | open |
| 1.9 | Secure: security review of the CLI (paths, archives, network inputs) with the `security-audit` skill | open |
| 1.10 | For agents: agent skill file (`skills/toolips/SKILL.md`) like komando's | open |
| 1.11 | Killer-feature gap list: compare with the top online tools and add the missing everyday verbs as plan items | open |

## Phase 2: paid plan

Starts only after phase 1 is done. Follows how komando's apps and cloud are built.

| # | Item | Status |
|---|---|---|
| 2.1 | Study komando-apps and komando-cloud; write the toolips design | open |
| 2.2 | `efolusi/toolips-apps`: Mac, iPhone, iPad app over the CLI core | open |
| 2.3 | `efolusi/toolips-cloud`: API, account, sync, billing, website | open |
| 2.4 | Self-hosted: the cloud service can run on the user's own server | open |
| 2.5 | Commercial: plans and pricing for apps and cloud, as in komando | open |
