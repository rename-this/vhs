#!/bin/bash

while IFS="$\n" read -r line; do
    echo "${line}" | jq --compact-output ".NumSpots += .NumSpots"
done
