#!/usr/bin/env bats
# MISO names the miso binary under test, see just test-cli.

@test "the agent refuses to run outside the builder VM" {
    run "$MISO" agent

    [ "$status" -eq 1 ]
    [[ "$output" == *"miso agent runs as the init of the builder VM"* ]]
}
