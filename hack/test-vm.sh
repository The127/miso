#!/usr/bin/env bash
# Runs the tests built with the vmtest tag, one package per VM, each test
# binary as the init of the host's kernel, which loads the modules put next
# to it. MISO_VMTEST_MODULES names the modules of another MISO_VMTEST_KERNEL.
# Arguments go to the test binary.
# MISO_VMTEST_BASE names a base image in qcow2 that is attached read-only
# with the serial miso-test-base, and once more with the serial of its
# digest, which the test finds in MISO_VMTEST_BASE_DIGEST. Small btrfs disks
# of each shape that btrfs-test-disk.sh knows are always attached with the
# serial miso-test-<shape>. A network card on QEMU's user network has the
# MAC the test finds in MISO_VMTEST_MAC, with the address and gateway for a
# run in MISO_VMTEST_ADDRESS and MISO_VMTEST_GATEWAY, and a service that
# answers miso at MISO_VMTEST_SERVICE.
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

# a network card with the MAC the test finds in MISO_VMTEST_MAC, the way a
# build hands the agent the MAC of its card
mac=52:54:00:6d:69:73
nic=(-netdev "user,id=net,net=10.0.2.0/24,host=10.0.2.2,dns=10.0.2.3,guestfwd=tcp:10.0.2.100:7-cmd:echo miso"
    -device "virtio-net-pci,netdev=net,mac=$mac")
environment+=("MISO_VMTEST_MAC=$mac" "MISO_VMTEST_ADDRESS=10.0.2.15/24" "MISO_VMTEST_GATEWAY=10.0.2.2")
# a service beyond the builder that needs no internet, QEMU answers it
environment+=("MISO_VMTEST_SERVICE=10.0.2.100:7")

# the modules the tests need that the kernel has not built in, each after
# what it depends on. Unpacked here, because kernels differ in how their
# modules are packed and not every kernel unpacks them itself. Numbered, so
# the VM loads them in order
mkdir -p "$work/root/modules"
loaded=()
for name in virtio_blk virtio_net macvlan btrfs overlay; do
    if grep -qE "/$name\.ko(\.[a-z]+)?$" "$modules/modules.builtin"; then
        continue
    fi

    line=$(grep -E "/$name\.ko(\.[a-z]+)?:" "$modules/modules.dep")
    read -ra deps <<< "${line#*:}"
    # modules.dep lists what a module needs before what that needs in turn
    for ((i = ${#deps[@]} - 1; i >= 0; i--)); do
        loaded+=("${deps[i]}")
    done

    loaded+=("${line%%:*}")
done

n=0
declare -A seen=()
for path in "${loaded[@]}"; do
    if [ -n "${seen[$path]:-}" ]; then
        continue
    fi

    seen[$path]=1
    n=$((n + 1))
    out=$(printf '%s/root/modules/%02d-%s' "$work" "$n" "$(basename "${path%%.ko*}").ko")
    case $path in
        *.ko.xz) xz -dc "$modules/$path" > "$out" ;;
        *.ko.zst) zstd -qdc "$modules/$path" > "$out" ;;
        *.ko.gz) gzip -dc "$modules/$path" > "$out" ;;
        *.ko) cp "$modules/$path" "$out" ;;
        *) echo "unknown module packing: $path" >&2; exit 1 ;;
    esac
done

# every package in a VM of its own, all at once, each with its own files
for pkg in $packages; do
    dir=$work/vm/$pkg
    mkdir -p "$dir/root"
    cp -r "$work/root/modules" "$dir/root/"
    CGO_ENABLED=0 go test -c -tags vmtest -o "$dir/root/init" "./$pkg"
    (cd "$dir/root" && find init modules | cpio --quiet -o -H newc) > "$dir/initrd"
done

for pkg in $packages; do
    dir=$work/vm/$pkg
    # the test binary stops a hanging test, the outer timeout a hanging VM.
    # Unbuffered, so the log of a killed run shows how far it got
    timeout 3m qemu-system-x86_64 -enable-kvm -cpu host -m 4G -nographic -no-reboot \
        -kernel "$kernel" -initrd "$dir/initrd" "${disks[@]}" "${nic[@]}" \
        -append "console=ttyS0 panic=-1 quiet ${environment[*]} -- -test.v -test.timeout=2m $*" \
        < /dev/null 2>&1 | sed -u 's/\r$//' > "$dir/log" &
done

wait

failed=0
for pkg in $packages; do
    echo "== $pkg on $kernel"
    cat "$work/vm/$pkg/log"
    if ! grep -qx 'miso-vmtest: exit 0' "$work/vm/$pkg/log"; then
        failed=1
    fi
done

exit "$failed"
