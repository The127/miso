#!/usr/bin/env bats
# MISO names the miso binary under test, see just test-cli.

setup() {
    # the cache of the test, which is not the user's
    export XDG_CACHE_HOME="$BATS_TEST_TMPDIR/cache"
    mkdir -p "$XDG_CACHE_HOME/miso/builder"
    DISK="$XDG_CACHE_HOME/miso/builder/layers.img"
}

# a cache disk of a size in MiB, as a build makes it
make_disk() {
    truncate -s "$1M" "$DISK"
    mkfs.ext4 -q "$DISK"
}

@test "a resize with no cache disk says there is none" {
    run "$MISO" resize

    [ "$status" -eq 1 ]
    [[ "$output" == *"no cache disk"* ]]
}

@test "a resize with a size it cannot read fails and names the flag" {
    make_disk 64

    run "$MISO" resize --size lots

    [ "$status" -eq 1 ]
    [[ "$output" == *"--size"* ]]
    [[ "$output" == *"lots"* ]]
}

@test "a resize to a size below the disk fails and leaves the disk as it was" {
    make_disk 64

    run "$MISO" resize --size 32MiB

    [ "$status" -eq 1 ]
    [[ "$output" == *"only grows"* ]]
    [ "$(stat -c %s "$DISK")" -eq $((64 * 1024 * 1024)) ]
}

@test "a resize makes the disk the size asked for and says so" {
    make_disk 64

    run "$MISO" resize --size 128MiB

    [ "$status" -eq 0 ]
    [[ "$output" == *"128 MiB"* ]]
    [ "$(stat -c %s "$DISK")" -eq $((128 * 1024 * 1024)) ]
}

@test "a resize with no size makes the disk 50 GiB" {
    make_disk 64

    run "$MISO" resize

    [ "$status" -eq 0 ]
    [[ "$output" == *"50 GiB"* ]]
    [ "$(stat -c %s "$DISK")" -eq $((50 * 1024 * 1024 * 1024)) ]
}

@test "a resize with no size leaves a disk that is large enough alone" {
    make_disk 64
    "$MISO" resize
    before="$(stat -c %s "$DISK")"

    run "$MISO" resize

    [ "$status" -eq 0 ]
    [[ "$output" == *"already"* ]]
    [ "$(stat -c %s "$DISK")" -eq "$before" ]
}
