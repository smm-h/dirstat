---
title: CLAUDE.md
---
# dirstat

:-: var key="project.description"

The authoritative specification is `docs/spec.md` (numbered requirements R1–R40).

## Project structure

Root files:

- `main.go` — app setup, registers the scan command via strictcli
- `scan.go` — scan command registration, flag validation, orchestration
- `exitcodes.go` — exit codes: 0 success, 1 general, 2 usage
- `version.go` — version from ldflags or debug.ReadBuildInfo

Internal packages (each paragraph is the package's doc comment, extracted from source):

:-: prose-desc path="internal/classify"

:-: prose-desc path="internal/config"

:-: prose-desc path="internal/jsonout"

:-: prose-desc path="internal/render"

:-: prose-desc path="internal/scan"

:-: prose-desc path="internal/test"

:-: prose-desc path="internal/testutil"

## Build and test

```bash
CGO_ENABLED=0 go build .        # build (must work without CGO)
go test ./... -race -count=1    # full suite
go vet ./... && gofmt -l .      # lint gate
go test ./internal/test -run TestGolden -args -update   # regenerate golden JSON
```

## Key conventions

- **CLI framework:** strictcli (`github.com/smm-h/strictcli/go/strictcli`). Handlers receive `map[string]interface{}` and return the exit code; never call `os.Exit` in handlers; errors go to stderr with an `error:` prefix.
- **JSON schema is a consumer contract:** field names are stable, evolution is additive-only. Never remove or rename fields.
- **Rendering-only flags never affect JSON** (`--show`, `--combined`, `--singletons`, `--legend`, `--colors`, `--human`, `--style`, `--top`).
- **Determinism:** worker-pool results are stored by index; groups are sorted with a full tie-break. Two runs must produce identical output.
- **Embedded-only configuration:** the text lists and themes are `go:embed` data under `internal/config/data/`; changing them means a rebuild. No runtime config files.
- **`.strictcli/schema.json`** is committed; regenerate with `./dirstat --dump-schema` after changing the CLI surface.
- **Releases:** rlsbl-managed (`.rlsbl/`), JSONL changelog required for every commit.
