#!/usr/bin/env bash
# Copies the kernel and its modules out of a qcow2 image with an ext4 root
# partition of the x86-64 root type, without root rights. Writes
# <out>/vmlinuz and <out>/modules, which test-vm.sh takes as
# MISO_VMTEST_KERNEL and MISO_VMTEST_MODULES.
# Usage: image-kernel.sh <image.qcow2> <out>
set -euo pipefail

image=$1
out=$2
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

qemu-img convert -O raw "$image" "$work/disk"
read -r start size < <(sfdisk -J "$work/disk" | python3 -c '
import json, sys
for p in json.load(sys.stdin)["partitiontable"]["partitions"]:
    if p["type"] == "4F68BCE3-E8CD-4DB1-96E7-FBCAF984B709":
        print(p["start"], p["size"])
')
dd if="$work/disk" of="$work/root" bs=512 skip="$start" count="$size" status=none

# -p prints /inode/mode/uid/gid/name/size/ per entry
version=$(debugfs -R "ls -p /usr/lib/modules" "$work/root" 2>/dev/null | cut -d/ -f6 | grep -E '^[0-9]' | head -n1)
if [ -z "$version" ]; then
    echo "no kernel in $image" >&2
    exit 1
fi

rm -rf "$out"
mkdir -p "$out"
debugfs -R "dump /boot/vmlinuz-$version $out/vmlinuz" "$work/root" 2>/dev/null
debugfs -R "rdump /usr/lib/modules/$version $out" "$work/root" 2>/dev/null
mv "$out/$version" "$out/modules"
test -s "$out/vmlinuz"
