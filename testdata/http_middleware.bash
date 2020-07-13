#!/bin/bash

while IFS="$\n" read -r line; do
    echo "${line:2}" | jq --compact-output ".body += \" [[hijacked ${line:0:1}]]\""
done
