#!/bin/bash

_SERVER="false"
_WEB="false"
_PWD=$(pwd)

while getopts 'swh' OPTION; do
    case "$OPTION" in
    "s")
        _SERVER="true"
        ;;
    "w")
        _WEB="true"
        ;;
    "h")
        cat <<EOF
usage: $0 [flags...]
  -s start only server
  -w start only web
  -h show this help
EOF
        exit 0
        ;;
    *)
        echo "unknown option"
        exit 1
        ;;
    esac
done
shift "$(($OPTIND - 1))"

if [ "$_SERVER" = "false" ] && [ "$_WEB" = "false" ]; then
    _SERVER="true"
    _WEB="true"
fi

function start_server() {
    echo "start server"
    cd "$_PWD" || exit 1
    cd src || exit 1
    /Users/rzvejs/go/bin/gow run \
        -ldflags "-X rat/buildinfo.version=$(git describe --tags)" \
        . -c ../config-test.yaml
}

function start_web() {
    echo "start web"
    cd "$_PWD" || exit 1
    cd src/web || exit 1
    npm start
}

if [ "$_SERVER" = "true" ]; then
    start_server &
fi

if [ "$_WEB" = "true" ]; then
    start_web &
fi

wait
