#!/usr/bin/env bash
# Enforces the prose rule that nothing but the license contains em-dashes.
set -euo pipefail

emdash=$(printf '\342\200\224')
if git grep -n "$emdash" -- ':!LICENSE'; then
    echo "em-dashes found, replace them"
    exit 1
fi
