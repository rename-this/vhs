#!/bin/bash

trap exit INT

while IFS="$\n" read -r line; do
    >&2 echo "recieved ${line:0:1}"
    echo "${line:2}" | jq --compact-output ".body += \" [[hijacked ${line:0:1}]]\""
done
