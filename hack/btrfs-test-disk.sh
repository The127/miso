#!/usr/bin/env bash
# Writes a raw disk with a GPT and one x86-64 root partition that holds a
# btrfs with an etc/os-release saying ID=miso-test. Needs no root.
set -euo pipefail

out=$1
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

mkdir -p "$work/root/etc"
echo "ID=miso-test" > "$work/root/etc/os-release"

# mkfs.btrfs cannot write at an offset, so the file system is made on its
# own and copied into the partition
truncate -s 128M "$work/fs.img"
mkfs.btrfs -q --rootdir "$work/root" "$work/fs.img"

truncate -s 130M "$out"
sfdisk --quiet "$out" <<EOF
label: gpt
start=2048, size=262144, type=4F68BCE3-E8CD-4DB1-96E7-FBCAF984B709
EOF
dd if="$work/fs.img" of="$out" bs=1M seek=1 conv=notrunc,sparse status=none
