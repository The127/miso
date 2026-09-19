#!/bin/sh
# A meeting point for the VM tests. QEMU runs it once for every connection,
# with the connection as its input and output. It answers met once another
# connection is open at the same time, and alone after ten seconds, so a
# test can prove two runs talk at once.
# Usage: meet.sh <directory shared by the connections>
dir=$1
mkdir -p "$dir"
me=$(mktemp "$dir/connection.XXXXXX")

waited=0
while [ "$(ls "$dir" | wc -l)" -lt 2 ]; do
    if [ "$waited" -ge 100 ]; then
        rm -f "$me"
        echo alone
        exit 0
    fi

    waited=$((waited + 1))
    sleep 0.1
done

echo met
# long enough for the other to see this one too
sleep 1
rm -f "$me"
