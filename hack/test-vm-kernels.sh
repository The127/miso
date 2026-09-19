#!/usr/bin/env bash
# Runs the VM tests on the host's kernel and on the base image's at once,
# and prints what each printed once both are done. Arguments go to both.
set -uo pipefail

out=$(mktemp -d)
trap 'rm -rf "$out"' EXIT

bash hack/test-vm.sh "$@" > "$out/host" 2>&1 &
host=$!
bash hack/test-vm-image-kernel.sh "$@" > "$out/image" 2>&1 &
image=$!

wait "$host"
hostcode=$?
wait "$image"
imagecode=$?

cat "$out/host" "$out/image"
if [ "$hostcode" -ne 0 ] || [ "$imagecode" -ne 0 ]; then
    exit 1
fi
