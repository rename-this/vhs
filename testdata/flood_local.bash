#!/bin/bash

hping3 \
    -c 1000 \
    -S \
    -w 64 \
    -p 8111 \
    --flood \
    -e "test111" \
    127.0.0.1
