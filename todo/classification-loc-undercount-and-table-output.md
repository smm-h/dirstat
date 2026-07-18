# Text/binary classification gaps, LOC undercounting, and table format-name truncation

## Context

Observed while exercising `dirstat scan` v0.2.1 across a large mixed directory tree (Go, Python, TypeScript, Godot, and data-heavy projects), always with the default `--method hybrid`. Three distinct issues surfaced, all reproducible with plain scans and no exotic flags. They share a root cause in the classification pipeline, so they are filed together.

## Problem 1: files with unrecognized extensions are classified binary without sniffing

Hybrid mode sniffs only extensionless files (as documented). But files whose extension is *not* in the built-in extension map appear to be classified as binary outright, with no content sniff. Since LOC is only computed for text-classified groups, every text format missing from the map silently contributes zero LOC and lands in the Binary table under `--no-combined`.

Observed consequences:

- In a Godot game project, `.gd` (GDScript), `.tscn`, `.tres`, `.gdshader`, `.godot`, and `.uid` — all plain-text formats — showed `-` for every LOC column. 41 GDScript files totaling ~300KB of source counted as zero lines of code.
- `go.mod` and `go.sum` (grouped as `mod` / `sum`) showed no LOC despite being text.
- A `uv.lock` file (TOML text) was grouped as `lock` and rendered in the **Binary Files by Format** table.
- A small `.example` text file (37 bytes) was classified binary.

The net effect: total-loc and the text/binary split are silently wrong for any project using formats outside the map. This contradicts the tool's purpose — the summary reads as authoritative but undercounts.

## Problem 2: empty (0-byte) files are classified binary

A 0-byte file appeared in the Binary table. An empty file has no bytes to sniff, and the classifier apparently defaults to binary on empty input. Empty is not binary: an empty source file is a text file with 0 LOC. This also interacts with Problem 1 — empty files inflate the binary side of the split for no reason.

## Problem 3: format names truncate ambiguously in tables

Sniffed MIME-type formats render as `applicat..` in the Format column. Two *different* `application/*` groups in the same table both display as `applicat..` and are indistinguishable — the top row of a large scan (8GB+ of files) is unreadable without switching to `--output json`. Truncation keeps the prefix, which for MIME types is the least informative part.

## Solutions

### Problem 1

- **Option A — sniff unknown extensions (extend hybrid semantics):** extension map hit → trust the map; miss → content-sniff, same as extensionless files. Pros: correct for every format ever encountered, no map maintenance treadmill, matches user intuition of what "hybrid" means. Cons: perf cost on trees with many unknown-extension files (bounded — only map misses are sniffed); grouping label question (keep the extension as the group name, use sniff result only for text/binary + LOC).
- **Option B — expand the built-in extension map:** add Godot formats (`gd`, `tscn`, `tres`, `gdshader`, `godot`, `uid`), `mod`, `sum`, and common lockfiles. Pros: trivial, zero perf cost. Cons: whack-a-mole forever; `lock` is genuinely ambiguous (uv/poetry/cargo locks are text, some `.lock` files are binary or empty sentinels), so a static map entry is wrong in one direction or the other.
- **Option C — both (most correct):** expand the map for common known-text formats *and* sniff on map miss. The map handles the hot path cheaply; sniffing makes the long tail correct. Ambiguous extensions like `lock` can be deliberately left out of the map so they always get sniffed per-file.

Option C is the most correct solution regardless of effort. Option A alone is acceptable; Option B alone is not (it fixes symptoms, not the class of bug).

### Problem 2

Classify 0-byte files as text with 0 LOC. Alternatively introduce a distinct "empty" classification excluded from both tables, but that adds schema surface (JSON output, `--type` filter semantics) for little gain. Recommended: empty → text, 0 LOC.

### Problem 3

- Widen or auto-size the Format column before truncating (it truncated even with terminal width apparently available).
- If truncation is unavoidable, truncate the middle or keep the *suffix* for MIME names (`..n/x-sharedlib` beats `applicat..`), since the subtype is what disambiguates.
- At minimum, never let two distinct groups render an identical truncated label — disambiguate or don't truncate.

## Affected areas

(Located by behavior, not by reading source — verify with grep.)

- Extension→format map and the text/binary classifier
- Hybrid-method sniffing gate (currently: extensionless only)
- Empty-file handling in the sniffer
- Table renderer column-width / truncation logic

## Testing

Per repo policy, red-green: add fixture files reproducing each case first (a `.gd` file, a text `uv.lock`, a 0-byte file, two distinct `application/*` binaries) and assert current wrong behavior fails, then fix. The classification fixes and the truncation fix are independent and can be separate commits.

## Effort estimate

Small-to-medium. Map additions and empty-file handling are trivial; sniff-on-map-miss is a contained change to the classification gate plus tests; truncation fix is renderer-local. Roughly a day including fixtures.
