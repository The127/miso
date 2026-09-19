#!/usr/bin/env bash
# Runs the tests built with the vmtest tag, one package per VM, each test
# binary as the init of the host's kernel, which loads the modules put next
# to it. MISO_VMTEST_MODULES names the modules of another MISO_VMTEST_KERNEL.
# Arguments go to the test binary. MISO_VMTEST_PACKAGES picks packages,
# for example internal/agent, instead of all.
# MISO_VMTEST_BASE names a base image in qcow2 that is attached read-only
# with the serial miso-test-base, and once more with the serial of its
# digest, which the test finds in MISO_VMTEST_BASE_DIGEST. Small btrfs disks
# of each shape that btrfs-test-disk.sh knows are always attached with the
# serial miso-test-<shape>. A cache disk made by hack/cachedisk is attached
# writable with the serial miso-cache, and a disk with no file system with
# the serial miso-test-blank, and a small cache disk read-only with the
# serial miso-test-crash. A network card on QEMU's user network has the
# MAC the test finds in MISO_VMTEST_MAC, with the address and gateway for a
# run in MISO_VMTEST_ADDRESS and MISO_VMTEST_GATEWAY, their IPv6 twins in
# MISO_VMTEST_ADDRESS6 and MISO_VMTEST_GATEWAY6, a service that
# answers miso at MISO_VMTEST_SERVICE and the meeting point of meet.sh at
# MISO_VMTEST_MEET. A service on the host's loopback, which a run must never
# reach, is at MISO_VMTEST_HOST, and one on the internet reached over IPv6 at
# MISO_VMTEST_SERVICE6, and one on a host's own network at
# MISO_VMTEST_LOCAL6, and one on an IPv4 address reached through NAT64 at
# MISO_VMTEST_MAPPED6, and through a network's own NAT64 prefix at
# MISO_VMTEST_LOCALMAPPED6. The nameservers are at MISO_VMTEST_NAMESERVER and
# MISO_VMTEST_NAMESERVER6, answering every name with 192.0.2.53 and
# 2001:db8::53.
set -euo pipefail

# QEMU runs in a network of its own, so the host QEMU connects to has only
# what the harness puts there, and no internet. In a mount namespace of its
# own too, so the resolv.conf QEMU reads is the harness's and the
# developer's own stays as it is
if [ -z "${MISO_VMTEST_NETWORK:-}" ]; then
    MISO_VMTEST_NETWORK=1 exec unshare --user --map-root-user --net --mount bash "$0" "$@"
fi

ip link set lo up
# an address from the documentation range stands in for the internet, one
# from the unique local range for a host's own network
ip addr add 2001:db8::1/128 dev lo
ip addr add fdcc::1/128 dev lo
# and one from the NAT64 range, how an IPv6-only host reaches IPv4
ip addr add 64:ff9b::c000:201/128 dev lo
# and one from the range a network picks its own NAT64 prefix from
ip addr add 64:ff9b:1::c000:201/128 dev lo

kernel=${MISO_VMTEST_KERNEL:-/lib/modules/$(uname -r)/vmlinuz}
modules=${MISO_VMTEST_MODULES:-/lib/modules/$(uname -r)}
work=$(mktemp -d)
python3 hack/service.py 2001:db8::1 7 &
service=$!
python3 hack/service.py fdcc::1 7 &
local=$!
python3 hack/service.py 64:ff9b::c000:201 7 &
mapped=$!
python3 hack/service.py 64:ff9b:1::c000:201 7 &
localmapped=$!
python3 hack/resolver.py &
resolver=$!
exec {loopback}< <(python3 hack/loopback.py)
listener=$!
trap 'kill "$listener" "$service" "$local" "$mapped" "$localmapped" "$resolver"; rm -rf "$work"' EXIT
read -r port <&"$loopback"

# QEMU's built-in nameserver forwards to the resolvers of the host it runs
# on, and in this network that is the harness's own. Over the real file,
# because QEMU reads it by the name every resolver library uses
printf 'nameserver 127.0.0.1\nnameserver ::1\n' > "$work/resolv.conf"
resolvers=$(readlink -f /etc/resolv.conf)
if ! mount --bind "$work/resolv.conf" "$resolvers"; then
    echo "cannot put the harness's resolv.conf over $resolvers" >&2
    exit 1
fi

# the tests that must never reach it only look for its answer to be missing,
# so a listener that answers nothing would make them pass
answer=$(python3 -c 'import socket, sys; print(socket.create_connection(("127.0.0.1", int(sys.argv[1])), 5).recv(64).decode(), end="")' "$port")
if [ "$answer" != "loopback" ]; then
    echo "the loopback service answers $answer, not loopback" >&2
    exit 1
fi

# a resolver that answers nothing would look like a fence that drops DNS
resolved=$(python3 -c '
import socket, struct
query = struct.pack("!HHHHHH", 1, 0x0100, 1, 0, 0, 0) + b"\x04miso\x04test\x00" + struct.pack("!HH", 1, 1)
asking = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
asking.settimeout(5)
asking.sendto(query, ("127.0.0.1", 53))
print(socket.inet_ntoa(asking.recv(512)[-4:]), end="")
')
if [ "$resolved" != "192.0.2.53" ]; then
    echo "the resolver answers $resolved, not 192.0.2.53" >&2
    exit 1
fi

packages=${MISO_VMTEST_PACKAGES:-$( (grep -rlx --include='*_test.go' '//go:build vmtest' cmd internal || true) | xargs -r -n1 dirname | sort -u)}
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

# a small cache disk the tests copy into memory, to crash a copy of it
go run ./hack/cachedisk "$work/crash.img" $((32 << 20))
disks+=(-drive "file=$work/crash.img,format=raw,if=none,readonly=on,id=crash"
    -device virtio-blk-pci,drive=crash,serial=miso-test-crash)

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
# a decoy card first, so the tests pass only when the card is found by its
# MAC, never by its name or its place
nic=(-netdev hubport,id=decoy,hubid=0 -device virtio-net-pci,netdev=decoy,mac=52:54:00:00:00:01
    -netdev "user,id=net,net=10.0.2.0/24,host=10.0.2.2,dns=10.0.2.3,ipv6-net=fd6d:6973:6f00::/64,ipv6-host=fd6d:6973:6f00::2,ipv6-dns=fd6d:6973:6f00::3,guestfwd=tcp:10.0.2.100:7-cmd:echo miso,guestfwd=tcp:10.0.2.100:8-cmd:sh $PWD/hack/meet.sh $work/meet"
    -device "virtio-net-pci,netdev=net,mac=$mac")
environment+=("MISO_VMTEST_MAC=$mac" "MISO_VMTEST_ADDRESS=10.0.2.15/24" "MISO_VMTEST_GATEWAY=10.0.2.2")
environment+=("MISO_VMTEST_ADDRESS6=fd6d:6973:6f00::15/64" "MISO_VMTEST_GATEWAY6=fd6d:6973:6f00::2")
# a service beyond the builder that needs no internet, QEMU answers it, and
# a meeting point that answers met to two connections open at once
environment+=("MISO_VMTEST_SERVICE=10.0.2.100:7" "MISO_VMTEST_MEET=10.0.2.100:8")
# QEMU forwards to services in IPv4 only, so the IPv6 one is on the network
# QEMU runs in
environment+=("MISO_VMTEST_SERVICE6=[2001:db8::1]:7" "MISO_VMTEST_LOCAL6=[fdcc::1]:7")
environment+=("MISO_VMTEST_MAPPED6=[64:ff9b::c000:201]:7" "MISO_VMTEST_LOCALMAPPED6=[64:ff9b:1::c000:201]:7")
# the nameservers QEMU answers itself, forwarding to the harness's resolver,
# which answers every name with 192.0.2.53 and 2001:db8::53
environment+=("MISO_VMTEST_NAMESERVER=10.0.2.3" "MISO_VMTEST_NAMESERVER6=fd6d:6973:6f00::3")
# QEMU maps the gateway to the host's loopback
environment+=("MISO_VMTEST_HOST=10.0.2.2:$port")

# the modules the tests need that the kernel has not built in, each after
# what it depends on. Unpacked here, because kernels differ in how their
# modules are packed and not every kernel unpacks them itself. Numbered, so
# the VM loads them in order
mkdir -p "$work/root/modules"
loaded=()
for name in virtio_blk virtio_net macvlan btrfs overlay sch_ingress cls_flower act_gact loop vsock_loopback; do
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
    # a cache disk made the way a build makes it, and a disk with no file
    # system, which must never be taken for one. Each VM writes its own,
    # QEMU gives a writable image to one VM at a time. Big enough for an
    # imported base, and sparse, so it takes only what the tests write
    go run ./hack/cachedisk "$dir/cache.img" $((4 << 30))
    truncate -s 16M "$dir/blank.img"
    (cd "$dir/root" && find init modules | cpio --quiet -o -H newc) > "$dir/initrd"
done

# one package shows its VM as it runs, several would mix their lines
live=$([ "$(wc -w <<< "$packages")" -eq 1 ] && echo 1 || true)
vms=()
for pkg in $packages; do
    dir=$work/vm/$pkg
    # the test binary stops a hanging test, the outer timeout a hanging VM.
    # Unbuffered, so the log of a killed run shows how far it got
    timeout 3m qemu-system-x86_64 -enable-kvm -cpu host -m 4G -nographic -no-reboot \
        -kernel "$kernel" -initrd "$dir/initrd" "${disks[@]}" "${nic[@]}" \
        -drive "file=$dir/cache.img,format=raw,if=none,id=cache" -device virtio-blk-pci,drive=cache,serial=miso-cache \
        -drive "file=$dir/blank.img,format=raw,if=none,id=blank" -device virtio-blk-pci,drive=blank,serial=miso-test-blank \
        -append "console=ttyS0 panic=-1 quiet ${environment[*]} -- -test.v -test.timeout=2m $*" \
        < /dev/null 2>&1 | sed -u 's/\r$//' | if [ -n "$live" ]; then tee "$dir/log"; else cat > "$dir/log"; fi &
    vms+=($!)
done

# the VMs alone, the services never end
wait "${vms[@]}"

failed=0
for pkg in $packages; do
    echo "== $pkg on $kernel"
    if [ -z "$live" ]; then
        cat "$work/vm/$pkg/log"
    fi

    if ! grep -qx 'miso-vmtest: exit 0' "$work/vm/$pkg/log"; then
        failed=1
    fi
done

exit "$failed"
