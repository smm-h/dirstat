---
title: Embedded Data
description: "How dirstat classifies files as text or binary using embedded extension and MIME type lists, and which directories are excluded by default."
---
# Embedded Data

dirstat embeds several data files into the binary at build time via `go:embed`. These control file classification and default traversal behavior. There are no runtime config files for classification -- changing the lists requires a rebuild.

The embedded data lives under `internal/config/data/` and is loaded by the `internal/config` package.

## Text vs. binary classification

Every file dirstat encounters is classified as either **text** or **binary**. This classification determines:

- Whether lines of code (LOC) are counted for the file (binary files have no LOC stats)
- How the file's row is colored in table output (text and binary rows use different theme colors)
- Whether the file appears in `--type text` or `--type binary` filtered output

The classification is strict: every file is one or the other, never "unknown." The mechanism depends on the grouping method:

| Method | Files with extension | Files without extension |
| --- | --- | --- |
| `ext` | Extension list lookup | Always binary |
| `type` | Content sniffing | Content sniffing |
| `hybrid` (default) | Extension list lookup | Content sniffing |

## Text extension list

**File:** `internal/config/data/text_extensions.txt` (embedded via `go:embed`)

Files with a recognized extension are classified as text if their normalized extension (last-dot suffix, lowercased, dot stripped) appears in this list. Files whose extension is not in the list are classified as binary. This lookup is used by the `ext` and `hybrid` grouping methods.

The list contains approximately 190 extensions covering programming languages, shell scripts, web technologies, data formats, documentation markup, and build system files.

:-: table-text-extensions

## Text MIME type list

**File:** `internal/config/data/text_mimetypes.txt` (embedded via `go:embed`)

When content sniffing is performed (always in `type` mode; for extensionless files in `hybrid` mode), dirstat uses the `github.com/gabriel-vasile/mimetype` library to detect the file's MIME type. The file is classified as text if:

1. The detected MIME type starts with `text/` (e.g., `text/plain`, `text/html`), **or**
2. The detected MIME type is in the text MIME type list

The second rule exists because many text-based formats have MIME types under `application/` or other top-level types rather than `text/`. For example, `application/json` is clearly text but does not start with `text/`.

The list contains 48 MIME types:

| MIME Type | Category |
| --- | --- |
| `application/json` | Data interchange |
| `application/xml` | Markup |
| `application/javascript` | Programming |
| `application/ecmascript` | Programming |
| `application/x-javascript` | Programming (legacy) |
| `application/x-sh` | Shell |
| `application/x-shellscript` | Shell |
| `application/x-perl` | Programming |
| `application/x-python` | Programming |
| `application/x-ruby` | Programming |
| `application/x-php` | Programming |
| `application/x-httpd-php` | Programming |
| `application/x-awk` | Shell |
| `application/x-gawk` | Shell |
| `application/x-nawk` | Shell |
| `application/x-sed` | Shell |
| `application/sql` | Data |
| `application/graphql` | Data |
| `application/ld+json` | Data interchange |
| `application/manifest+json` | Data interchange |
| `application/x-yaml` | Data |
| `application/yaml` | Data |
| `application/toml` | Data |
| `application/x-toml` | Data |
| `application/x-wine-extension-ini` | Config |
| `application/xhtml+xml` | Markup |
| `application/rss+xml` | Markup |
| `application/atom+xml` | Markup |
| `application/soap+xml` | Markup |
| `application/xslt+xml` | Markup |
| `application/mathml+xml` | Markup |
| `application/x-tex` | Documentation |
| `application/x-latex` | Documentation |
| `application/rtf` | Documentation |
| `application/postscript` | Documentation |
| `application/x-troff` | Documentation |
| `application/x-troff-man` | Documentation |
| `application/x-troff-me` | Documentation |
| `application/x-troff-ms` | Documentation |
| `application/x-info` | Documentation |
| `application/x-texinfo` | Documentation |
| `application/x-maker` | Documentation |
| `application/csv` | Data |
| `application/x-empty` | Empty file |
| `inode/x-empty` | Empty file |
| `image/svg+xml` | Vector graphics (text-based) |

## Default excludes

When `--exclude` is not explicitly passed (and no config file sets `exclude`), dirstat skips these directories and files by exact base-name match during traversal. Excluded directories are pruned entirely -- their contents are not descended into, counted, or classified.

:-: table-default-excludes

## Overriding exclusions

The default exclude list is replaced entirely when you provide your own. There is no additive merging -- any explicit exclusion list completely supersedes the built-in defaults.

### Via the `--exclude` flag

Pass `--exclude` one or more times to set the exclusion list:

```
dirstat scan --exclude node_modules --exclude .git
```

This scans everything except `node_modules` and `.git`. The 18 other built-in defaults are no longer excluded.

To scan with no exclusions at all, the config file approach is required (see below).

### Via a config file

Set the `exclude` key in a TOML config file passed with `--config`:

```toml
# scan.toml
exclude = ["node_modules", ".git", "dist"]
```

```
dirstat scan --config scan.toml
```

To scan everything with no exclusions:

```toml
exclude = []
```

A key set both in the config file and on the command line is a hard error -- use one or the other. See the [spec](spec.html) (R41-R47) for full config file semantics.
