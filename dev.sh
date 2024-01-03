#!/bin/bash

_SERVER="false"
_WEB="false"
_ROOT_DIR=$(pwd)

while getopts 'swhr:' OPTION; do
    case "$OPTION" in
    "s")
        _SERVER="true"
        ;;
    "w")
        _WEB="true"
        ;;
    "r")
        _ROOT_DIR="$OPTARG"
        ;;
    "h")
        cat <<EOF
usage: $0 [flags...]
  -s start only server
  -w start only web
  -r filepath to rat project root dir
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

echo "server:   $_SERVER"
echo "web:      $_WEB"
echo "root dir: $_ROOT_DIR"

function start_server() {
    echo "start server"
    cd "$_ROOT_DIR" || exit 1
    cd src || exit 1
    /Users/rzvejs/go/bin/gow run \
        -ldflags "-X rat/buildinfo.version=$(git describe --tags)" \
        . -c ../config-test.yaml
}

function start_web() {
    echo "start web"
    cd "$_ROOT_DIR" || exit 1
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
