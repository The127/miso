#!/usr/bin/env bash
# Runs test-vm.sh on the kernel of the base image in MISO_VMTEST_BASE instead
# of the host's, so the tests also pass on a kernel that builds as modules
# what the host's has built in. The kernel is copied out once per image and
# kept in the user's cache. Arguments go to test-vm.sh.
set -euo pipefail

if [ -z "${MISO_VMTEST_BASE:-}" ]; then
    echo "MISO_VMTEST_BASE names no base image to take a kernel from" >&2
    exit 1
fi

digest=$(sha256sum "$MISO_VMTEST_BASE" | cut -c1-16)
kernel=${XDG_CACHE_HOME:-$HOME/.cache}/miso-vmtest/kernel-$digest
if [ ! -s "$kernel/vmlinuz" ]; then
    bash hack/image-kernel.sh "$MISO_VMTEST_BASE" "$kernel"
fi

MISO_VMTEST_KERNEL=$kernel/vmlinuz MISO_VMTEST_MODULES=$kernel/modules exec bash hack/test-vm.sh "$@"
