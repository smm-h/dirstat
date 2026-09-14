#!/usr/bin/env bash
# Post-release hook. Runs after a successful rlsbl release.
# Rebuilds this project's docs in the unified documentation assembly.

set -euo pipefail

if command -v selfdoc &>/dev/null && [ -f selfdoc.json ]; then
  if python3 -c "import json; c=json.load(open('selfdoc.json')); exit(0 if c.get('assembly') else 1)" 2>/dev/null; then
    echo "Pushing to documentation assembly..."
    selfdoc assembly push || echo "Warning: assembly push failed (non-fatal)"
  fi
fi
