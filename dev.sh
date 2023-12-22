#!/bin/bash

_PWD=$(pwd)

function start_server() {
    cd "$_PWD" || exit 1
    cd src || exit 1
    /Users/rzvejs/go/bin/gow run . -c ../config-test.yaml
}

function start_web() {
    cd "$_PWD" || exit 1
    cd src/web || exit 1
    npm start
}

start_server &
start_web &

wait
