#!/usr/bin/env bats
# MISO names the miso binary under test, see just test-cli.

@test "a prune with no limit fails and names the flags" {
    run "$MISO" prune

    [ "$status" -eq 1 ]
    [[ "$output" == *"--older-than"* ]]
    [[ "$output" == *"--keep-storage"* ]]
}

@test "a prune with a size it cannot read fails and names the flag" {
    run "$MISO" prune --keep-storage lots

    [ "$status" -eq 1 ]
    [[ "$output" == *"--keep-storage"* ]]
    [[ "$output" == *"lots"* ]]
}

@test "a prune with a size no disk can hold fails and names the flag" {
    run "$MISO" prune --keep-storage 10EiB

    [ "$status" -eq 1 ]
    [[ "$output" == *"--keep-storage"* ]]
    [[ "$output" == *"too large"* ]]
}

@test "a prune with a negative age fails and names the flag" {
    run "$MISO" prune --older-than=-1h

    [ "$status" -eq 1 ]
    [[ "$output" == *"--older-than"* ]]
    [[ "$output" == *"above zero"* ]]
}

@test "a prune with a storage limit of zero fails and names the flag" {
    run "$MISO" prune --keep-storage 0

    [ "$status" -eq 1 ]
    [[ "$output" == *"--keep-storage"* ]]
    [[ "$output" == *"above zero"* ]]
}
