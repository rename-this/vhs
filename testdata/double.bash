#!/bin/bash

while IFS="$\n" read -r line; do
    >&2 echo "recieved ${line:0:1}"
    echo "${line}" | jq --compact-output ".NumSpots += .NumSpots"
done
