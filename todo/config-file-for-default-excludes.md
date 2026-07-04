# Config file for default excludes

## Problem

Running dirstat on large trees (e.g. ~/Projects with 692k files) requires passing ~20 `--exclude` flags every time to skip build artifacts, caches, and vendored dependencies. The exclude list is stable across invocations and should not need to be repeated on every call.

## Known excludes (from a real ~/Projects scan)

Package managers: `node_modules`, `.venv`, `vendor`
Version control: `.git`
Build output: `build`, `dist`, `target`, `zig-out`, `zig-pkg`, `.next`, `.svelte-kit`
Bytecode/caches: `__pycache__`, `.mypy_cache`, `.ruff_cache`, `.pytest_cache`, `.hypothesis`
IDE/tooling: `.gradle`, `.idea`, `.wrangler`, `.vscode`

Without excludes: 692k files, 12.5s. With all excludes: 71k files, 2.0s.

## Scope

Design and implement a file-based configuration mechanism so default excludes (and potentially other scan defaults) don't need to be passed as CLI flags every time.

## Open questions (for the implementor to decide)

- Config format and location (project-local, user-level, both with precedence?)
- Whether dirstat ships built-in default excludes or starts blank
- How CLI flags interact with config (additive, override, or both)
- Whether this extends to other scan defaults beyond excludes (depth, method, hidden, ignored, etc.)
