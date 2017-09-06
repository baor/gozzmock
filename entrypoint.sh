#!/bin/sh

args="$@"

if [ -n "$EXPECTATIONS" ]; then
    args=$args" --expectations $EXPECTATIONS"

./gozzmock_bin $args