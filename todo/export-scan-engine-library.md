# Export the scan engine as an importable library

## Context

A planned machine-scoped disk tool (indexing, per-directory declaration
files, space reclaim, filesystem watching) wants to reuse dirstat's scanning
core rather than reimplement it: the directory walker with its filter order
(exclude set, hidden handling, depth, symlink policy, non-regular-file
skipping), the parallel classification pool with deterministic by-index
result storage, the in-process gitignore matching via go-git, and the
embedded text-extension/mimetype classification data.

All of that currently lives under `internal/`, which Go makes unimportable
from outside this module. dirstat itself stays exactly what it is — the
focused, read-only, single-command stats instrument; none of its spec
commitments change. This is a packaging refactor, not a behavior change.

## Problem

`internal/scan` (walk, classification, gitignore), `internal/config`
(embedded classification data), and the option types they expose cannot be
consumed by another module. The consumer would otherwise copy the code,
creating a second diverging implementation of the walker and classification
logic — the same-fact-in-two-places situation the fleet engineers away.

## Work

1. Decide the exported API surface: walker entry point + options (excludes,
   hidden, depth, ignored-mode, workers), the file-entry/result types,
   classification (method, text/binary verdicts, format grouping), gitignore
   matching, and whether `internal/scanconfig` (TOML scan-config) stays
   CLI-private (recommended: it stays private — it is CLI surface, not
   engine).
2. Promote the chosen packages from `internal/` to exported package paths
   within this module, and update the CLI to consume the exported packages.
3. Exported identifiers become released API surface: name them deliberately,
   document them (package doc comments feed the selfdoc-generated internal
   docs — those docs' paths/titles will need to follow the packages), and
   keep `CGO_ENABLED=0` compatibility.
4. Tests move/stay with their packages; behavior is pinned by the existing
   suite and the JSON snapshot tests — a release of this refactor should
   show zero behavioral diff.

## Options

1. **Export in place, same module** (recommended): packages move out of
   `internal/` within this repo. Single authority stays here, one release
   train, consumers import versioned tags of this module. Cost: the engine
   API is now semver-visible surface of this project (fine in 0.x, but name
   things carefully).
2. **Separate library repo/module**: engine extracted into its own repo,
   dirstat becomes its first consumer. Cleanest separation of API identity,
   but adds a repo, a release train, and a naming decision, for one known
   consumer besides dirstat itself. Not justified yet.

## Affected files

- `internal/scan/` (walk, scan, gitignore) — moves/renames to exported paths
- `internal/config/` (embedded data) — exported or wrapped by the exported
  engine package
- Root `scan.go` — import updates
- `docs/` — selfdoc-generated per-package docs follow the moves
- Spec: an entry noting the engine is exported library surface (the CLI's
  behavioral requirements are unchanged)

## Effort

Small-to-medium: mostly mechanical moves plus one real design pass over the
exported API names and option types. Zero intended behavior change,
verifiable by the existing suite and snapshot tests.
