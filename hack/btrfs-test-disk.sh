#!/usr/bin/env bash
# Writes a raw disk with a GPT and one x86-64 root partition that holds a
# btrfs. Needs no root. The shape is one of
#   flat        etc/os-release saying ID=miso-test in the top level
#   subvolumes  Fedora's layout: subvolumes root, var and boot, and an fstab
#               in root that mounts them, os-release saying
#               ID=miso-test-subvolumes
#   default     openSUSE's layout: the default subvolume snapshot is the root
#               and its fstab mounts the subvolume var
#   twins       subvolumes one and two, each with an fstab that makes it the
#               root
#   missing     a root whose fstab mounts /var from a subvolume that is not
#               there
#   escape      like subvolumes, but var in root is a link to /escaped, out
#               of the image
set -euo pipefail

out=$1
shape=$2
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

tree=$work/tree
subvolumes=()
case $shape in
flat)
    mkdir -p "$tree/etc"
    echo "ID=miso-test" > "$tree/etc/os-release"
    ;;
subvolumes)
    mkdir -p "$tree/root/etc" "$tree/root/var" "$tree/root/boot" "$tree/var" "$tree/boot"
    echo "ID=miso-test-subvolumes" > "$tree/root/etc/os-release"
    cat > "$tree/root/etc/fstab" <<EOF
LABEL=test / btrfs compress=zstd:1,subvol=root 0 0
LABEL=test /var btrfs compress=zstd:1,subvol=var 0 0
LABEL=test /boot btrfs compress=zstd:1,subvol=boot 0 0
EOF
    echo var > "$tree/var/marker"
    echo boot > "$tree/boot/marker"
    subvolumes=(--subvol rw:root --subvol rw:var --subvol rw:boot)
    ;;
default)
    mkdir -p "$tree/snapshot/etc" "$tree/snapshot/var" "$tree/var"
    echo "ID=miso-test-default" > "$tree/snapshot/etc/os-release"
    cat > "$tree/snapshot/etc/fstab" <<EOF
LABEL=test / btrfs subvol=/snapshot 0 0
LABEL=test /var btrfs subvol=/var 0 0
EOF
    echo var > "$tree/var/marker"
    subvolumes=(--subvol default:snapshot --subvol rw:var)
    ;;
twins)
    for name in one two; do
        mkdir -p "$tree/$name/etc"
        echo "LABEL=test / btrfs subvol=$name 0 0" > "$tree/$name/etc/fstab"
    done
    subvolumes=(--subvol rw:one --subvol rw:two)
    ;;
missing)
    mkdir -p "$tree/root/etc" "$tree/root/var"
    cat > "$tree/root/etc/fstab" <<EOF
LABEL=test / btrfs subvol=root 0 0
LABEL=test /var btrfs subvol=nosuch 0 0
EOF
    subvolumes=(--subvol rw:root)
    ;;
escape)
    mkdir -p "$tree/root/etc" "$tree/var"
    ln -s /escaped "$tree/root/var"
    cat > "$tree/root/etc/fstab" <<EOF
LABEL=test / btrfs subvol=root 0 0
LABEL=test /var btrfs subvol=var 0 0
EOF
    echo var > "$tree/var/marker"
    subvolumes=(--subvol rw:root --subvol rw:var)
    ;;
*)
    echo "unknown shape $shape" >&2
    exit 1
    ;;
esac

# mkfs.btrfs cannot write at an offset, so the file system is made on its
# own and copied into the partition
truncate -s 128M "$work/fs.img"
mkfs.btrfs -q --rootdir "$tree" "${subvolumes[@]}" "$work/fs.img"

truncate -s 130M "$out"
sfdisk --quiet "$out" <<EOF
label: gpt
start=2048, size=262144, type=4F68BCE3-E8CD-4DB1-96E7-FBCAF984B709
EOF
dd if="$work/fs.img" of="$out" bs=1M seek=1 conv=notrunc,sparse status=none
