#!/usr/bin/env bash
# Lands a branch of unsigned commits on main under the caller's signature.
# Run it on your own machine, after reading the commits it shows.
set -euo pipefail

if [ "${1:-}" = "--sign" ]; then
    who="$(git config user.name) <$(git config user.email)>"
    # a second -s would stack a duplicate once a trailer follows the first
    if git log -1 --format=%B | grep -qxF "Signed-off-by: $who"; then
        exec git commit --amend --no-edit -S
    fi
    exec git commit --amend --no-edit -S -s
fi

branch=${1:?usage: just land <branch>}

if [ -n "$(git status --porcelain)" ]; then
    echo "working tree is not clean" >&2
    exit 1
fi

start=$(git rev-parse --abbrev-ref HEAD)
git fetch origin main "$branch"

git log --stat --format='%n%h %an <%ae>%n%B' "origin/main..origin/$branch"
read -r -p "sign and land these commits on main? [y/N] " answer
[ "$answer" = "y" ] || exit 1

git checkout --detach "origin/$branch"
git rebase origin/main --exec "bash $(pwd)/hack/land.sh --sign"
git push origin HEAD:main
git checkout "$start"
