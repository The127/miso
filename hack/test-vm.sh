#!/usr/bin/env bash
# Runs the tests built with the vmtest tag, one package per VM, each test
# binary as the init of the host's kernel. Arguments go to the test binary.
# MISO_VMTEST_BASE names a base image in qcow2 that is attached read-only
# with the serial miso-test-base. Small btrfs disks of each shape that
# btrfs-test-disk.sh knows are always attached with the serial
# miso-test-<shape>.
set -euo pipefail

kernel=${MISO_VMTEST_KERNEL:-/lib/modules/$(uname -r)/vmlinuz}
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

packages=$( (grep -rlx --include='*_test.go' '//go:build vmtest' cmd internal || true) | xargs -r -n1 dirname | sort -u)
if [ -z "$packages" ]; then
    echo "no vmtest packages"
    exit 0
fi

disks=()
for shape in flat subvolumes default twins missing escape; do
    bash hack/btrfs-test-disk.sh "$work/$shape.img" "$shape"
    disks+=(-drive "file=$work/$shape.img,format=raw,if=none,readonly=on,id=$shape"
        -device "virtio-blk-pci,drive=$shape,serial=miso-test-$shape")
done

if [ -n "${MISO_VMTEST_BASE:-}" ]; then
    disks+=(-drive "file=$MISO_VMTEST_BASE,format=qcow2,if=none,readonly=on,id=base"
        -device virtio-blk-pci,drive=base,serial=miso-test-base)
fi

failed=0
for pkg in $packages; do
    echo "== $pkg"
    mkdir -p "$work/root"
    CGO_ENABLED=0 go test -c -tags vmtest -o "$work/root/init" "./$pkg"
    (cd "$work/root" && echo init | cpio --quiet -o -H newc) > "$work/initrd"

    # the test binary stops a hanging test, the outer timeout a hanging VM
    timeout 10m qemu-system-x86_64 -enable-kvm -cpu host -m 1G -nographic -no-reboot \
        -kernel "$kernel" -initrd "$work/initrd" "${disks[@]}" \
        -append "console=ttyS0 panic=-1 quiet -- -test.v -test.timeout=5m $*" \
        | tr -d '\r' | tee "$work/log"

    if ! grep -qx 'miso-vmtest: exit 0' "$work/log"; then
        failed=1
    fi
done

exit "$failed"
