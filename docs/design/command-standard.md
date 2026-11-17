# The toolips command standard

## Grammar

- `toolips <noun> <verb> [args] [flags]`.
- Top-level about: `version`, `list`, `guide`, `help`, `completion`.
- One need, one command. Aliases are hidden, not second docs.

## Flags

- `-o` / `--output`: `text`, `json` or `auto` (terminal → text, else JSON).
- `--pretty`: indent JSON.
- `--dry-run`: on write verbs; print destinations, write nothing.
- `--out`: single output file. `--outdir`: batch directory.

## Help and errors

- `Short` is one line. `Long` only when behaviour is surprising
  (e.g. `pdf to-image` = embedded extract, not rasterize).
- Typed errors via `out.Errf` with an UPPER_SNAKE code and a hint.
- Usage mistakes are exit 2 (`USAGE`).

## Output contract

- Field names are only added, never renamed or removed, within contract 1.
- Success and failure share the same envelope shape (`ok`, `command`,
  `data`, `error`, `meta`).
