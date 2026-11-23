# Contributing

Thank you for helping. Please write issues, pull requests and docs in English.

toolips is open source under the [AGPL-3.0](LICENSE), and is also sold under a
[commercial license](COMMERCIAL-LICENSE.md) that pays for its work. Before your
first pull request is merged, please sign the
[Contributor License Agreement](CLA.md): type its
sentence in the pull request description (every commit author signs), and a
check looks for it. You keep the copyright of what you wrote.

toolips is a small CLI for PDF, image and everyday security checks, for people
and AI agents. One static binary, no menus, no accounts.

## Report a bug or suggest a feature

Open a GitHub issue. For a security problem, see [SECURITY.md](SECURITY.md).

## Before you add something

- It fits PDF, image or security checks people run every day.
- It works as `toolips <noun> <verb>`, for a person and for an agent.
- Dependencies are MIT, BSD, ISC or Apache, and pure Go (`CGO_ENABLED=0`).
- No account, billing or feature gates.

Design rules live in [docs/design/](docs/design/).

## Build and test

You need [Go](https://go.dev/dl/) (the version in [`go.mod`](go.mod) or newer)
and `make`.

```bash
make build    # compile ./bin/toolips
make test     # unit tests
make lint     # gofmt, go vet, staticcheck
```

## Add or change a command

1. Put work logic in `internal/pdfx`, `internal/imgx` or `internal/secx`.
2. Wire flags in `internal/cli/`. Print only via `out.P.Print`.
3. Use typed errors: `out.Errf(exit, "CODE", "hint", "message")`.
4. Respect `--dry-run` on write commands.
5. Keep functions ≤100 lines and files ≤800 where practical.
6. Add a unit test; network tests should skip without network or use `httptest`.

## Send a pull request

1. Run `make lint` and `make test`.
2. Add a line under `## [Unreleased]` in [`CHANGELOG.md`](CHANGELOG.md) if users will notice.
3. Open the pull request. CI runs lint, tests and a cross build.

## Merging outside pull requests

For maintainers. Squash-merge a pull request from outside, and add one
trailer per signer to the merge commit, so git keeps the CLA record:

```
CLA-Signed-By: <login> (#<number>, <YYYY-MM-DD>)
```

Never delete the `cla-signatures` branch: it holds the signatures given to
the old bot before 2026-11-23.
