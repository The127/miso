#!/usr/bin/env bash
# Runs the tests built with the vmtest tag, one package per VM, each test
# binary as the init of the host's kernel, which loads the modules put next
# to it. MISO_VMTEST_MODULES names the modules of another MISO_VMTEST_KERNEL.
# Arguments go to the test binary.
# MISO_VMTEST_BASE names a base image in qcow2 that is attached read-only
# with the serial miso-test-base, and once more with the serial of its
# digest, which the test finds in MISO_VMTEST_BASE_DIGEST. Small btrfs disks
# of each shape that btrfs-test-disk.sh knows are always attached with the
# serial miso-test-<shape>.
set -euo pipefail

kernel=${MISO_VMTEST_KERNEL:-/lib/modules/$(uname -r)/vmlinuz}
modules=${MISO_VMTEST_MODULES:-/lib/modules/$(uname -r)}
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

environment=()
if [ -n "${MISO_VMTEST_BASE:-}" ]; then
    disks+=(-drive "file=$MISO_VMTEST_BASE,format=qcow2,if=none,readonly=on,id=base"
        -device virtio-blk-pci,drive=base,serial=miso-test-base)

    # attached once more the way a build attaches it, by the serial of its
    # digest, which the kernel hands to the test in its environment
    digest=sha256:$(sha256sum "$MISO_VMTEST_BASE" | cut -c1-64)
    serial=${digest#sha256:}
    disks+=(-drive "file=$MISO_VMTEST_BASE,format=qcow2,if=none,readonly=on,id=digest"
        -device "virtio-blk-pci,drive=digest,serial=${serial:0:20}")
    environment+=("MISO_VMTEST_BASE_DIGEST=$digest")
fi

# the modules the tests need that the host's kernel has not built in
mkdir -p "$work/root/modules"
cp "$modules/kernel/fs/overlayfs/overlay.ko.xz" "$work/root/modules/"

failed=0
for pkg in $packages; do
    echo "== $pkg"
    CGO_ENABLED=0 go test -c -tags vmtest -o "$work/root/init" "./$pkg"
    (cd "$work/root" && find init modules | cpio --quiet -o -H newc) > "$work/initrd"

    # the test binary stops a hanging test, the outer timeout a hanging VM.
    # Unbuffered, so the log of a killed run shows how far it got
    timeout 3m qemu-system-x86_64 -enable-kvm -cpu host -m 4G -nographic -no-reboot \
        -kernel "$kernel" -initrd "$work/initrd" "${disks[@]}" \
        -append "console=ttyS0 panic=-1 quiet ${environment[*]} -- -test.v -test.timeout=2m $*" \
        | sed -u 's/\r$//' | tee "$work/log"

    if ! grep -qx 'miso-vmtest: exit 0' "$work/log"; then
        failed=1
    fi
done

exit "$failed"
