---
title: README.md
---
# dirstat

dirstat summarizes the files in a directory tree, grouped by format, with aggregate statistics (counts, sizes, lines of code) rendered as a colored terminal table or as JSON. Built for speed: no subprocesses, parallel scanning, single static binary.

## Quick start

```
go install github.com/smm-h/dirstat@latest
```

Summarize the current directory:

```
dirstat scan
```

Summarize a project, JSON output for tooling:

```
dirstat scan ~/src/myproject --output json
```

## Example

```
$ dirstat scan .

Summary
──────────────────────────────────────────────────
  Directories: 31
  Files: 60
  Files without extension: 17
  Max depth: 5
  Symlinks (skipped): 0
  Executables: 14
  Unique formats: 11
  Files sniffed: 17
  Unreadable (skipped): 0

Files by Format
──────────────────────────────────────────────────
┌────────┬───────┬────────────┬───────────┬─────────┐
│ Format │ Count │ Total Size │ Total LOC │ ...     │
├────────┼───────┼────────────┼───────────┼─────────┤
│ go     │ 21    │ 79.6KB     │ 2,893     │ ...     │
│ md     │ 1     │ 13.1KB     │ 232       │ ...     │
└────────┴───────┴────────────┴───────────┴─────────┘
```

## Commands

:-: table-commands path="."

## Grouping methods

| Method | Behavior |
|--------|----------|
| `ext` | Group by file extension only; text/binary decided by an embedded extension list; no content sniffing |
| `type` | Content-sniff every file; group by detected MIME type |
| `hybrid` (default) | Files with an extension behave as in `ext`; extensionless files are sniffed and grouped by MIME type |

## Flag overview

| Flag | Default | Purpose |
|------|---------|---------|
| `--method` | `hybrid` | Grouping method: `ext`, `type`, `hybrid` |
| `--depth` | `-1` | Max directory depth below the root (`-1` = unlimited) |
| `--exclude` | none | Exact base names to skip (repeatable) |
| `--ignored` | `exclude` | Gitignored paths: `include`, `exclude`, `only` |
| `--hidden` | `include` | Dot-prefixed entries: `include`, `exclude` |
| `--stats` | all | Stats to compute: `count`, `total-size`, `min-size`, `max-size`, `avg-size`, `total-loc`, `min-loc`, `max-loc`, `avg-loc` |
| `--type` | `both` | Filter groups: `text`, `binary`, `both` |
| `--sort-by` | `count` | Sort columns in precedence order (repeatable): `format` or any stat |
| `--sort-order` | `desc` | `asc` or `desc`, applied to all sort columns |
| `--top` | `-1` | Keep only the first N groups after sorting (table only) |
| `--output` | `table` | `table` or `json` |
| `--show` | `both` | Sections to render: `summary`, `table`, `both` (table only) |
| `--combined` | on | One merged table vs `--no-combined` split text/binary tables |
| `--singletons` | `show` | `collapse` merges one-file formats into a `(singletons)` row (table only) |
| `--list-no-ext` | off | List the paths of extensionless files |
| `--legend` | on | Text/binary color legend under the combined table |
| `--colors` | on | ANSI colors (auto-disabled when stdout is not a TTY) |
| `--human` | on | Human-readable sizes and thousands separators (table only) |
| `--style` | `unicode` | Table borders: `unicode` or `ascii` |

Rendering-only flags (`--show`, `--combined`, `--singletons`, `--legend`, `--colors`, `--human`, `--style`, `--top`) have no effect on JSON output. Note that without active colors (piped output or `--no-colors`), the table width is pinned to 80 columns so the output is byte-identical to non-TTY output.

## JSON output

`--output json` writes a single JSON object to stdout: no ANSI codes, raw integer values. The schema is a consumer contract — field names are stable and evolution is additive-only.

```json
{
  "dirstat_version": "0.1.0",
  "root": "/abs/path",
  "method": "hybrid",
  "summary": {
    "directories": 3, "files": 6, "files_without_extension": 1,
    "max_depth": 2, "symlinks": 0, "executables": 0,
    "unique_formats": 5, "files_sniffed": 1, "unreadable": 0
  },
  "groups": [
    {
      "format": "go", "text": true, "count": 2,
      "total_size": 39, "min_size": 10, "max_size": 29, "avg_size": 20,
      "total_loc": 4, "min_loc": 1, "max_loc": 3, "avg_loc": 2
    }
  ],
  "no_extension_files": ["Makefile"]
}
```

LOC fields are `null` for binary groups. Stats absent from `--stats` are omitted. `no_extension_files` is present only with `--list-no-ext`.

## Behavior notes

- Symlinks are never followed, but every symlink encountered in a traversed directory is counted.
- Unreadable directories and files are skipped and counted under `Unreadable (skipped)` — never a warning spew, never a crash.
- Gitignore matching is fully in-process (no `git` subprocess): patterns come from every `.gitignore` in the work tree, `.git/info/exclude`, and the global `core.excludesFile`. Outside a git work tree, `--ignored exclude`/`only` behave as if no patterns exist.
- LOC counting is skipped entirely (no file reads) unless a LOC stat is selected or sorted by.
- The text-extension and text-mimetype lists and both color themes are embedded in the binary; there are no runtime config files.
