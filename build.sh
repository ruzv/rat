#!/bin/bash

INSTALL="false"
BUILD="false"
SERVER="false"
WEB="false"
_DOCKER="false"

_SERVER_DIR="src"
_WEB_DIR="src/web"
_PWD=$(pwd)
_OUTPUT_BIN="$_PWD/rat"

while getopts 'ibswdh' OPTION; do
    case "$OPTION" in
    "i")
        INSTALL="true"
        ;;
    "b")
        BUILD="true"
        ;;
    "s")
        SERVER="true"
        ;;
    "w")
        WEB="true"
        ;;
    "d")
        _DOCKER="true"
        ;;
    "h")
        cat <<EOF
usage: $0 [flags...]
  -i - build and install server binary to GOPATH
  -b - build server binary to current directory
  -s - build only server
  -w - build only web
  -d - build docker image
  -h - show this help
examples:
  $0 -b -sw  - build web and then server binary
  $0 -i -sw  - build web and then server binary and install it
  $0 -d      - build docker image
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

_RAT_VERSION=$(git describe --tags)

echo "WEB:       $WEB"
echo "SERVER:    $SERVER"
echo "  INSTALL: $INSTALL"
echo "  BUILD:   $BUILD"
echo "DOCKER:    $_DOCKER"
echo
echo "RAT VERSION: $_RAT_VERSION"

if [ "$WEB" = "true" ]; then
    echo "building web"
    cd "$_WEB_DIR" || exit 1

    npm install || exit 1
    npm run build || exit 1

    cd "$_PWD" || exit 1
fi

if [ "$SERVER" = "true" ]; then
    cd "$_SERVER_DIR" || exit 1

    if [ "$BUILD" = "true" ]; then
        echo "building server"
        go build \
            -v \
            -ldflags "-X rat/buildinfo.version=$(git describe --tags)" \
            -o "$_OUTPUT_BIN" ||
            exit 1
    fi

    if [ "$INSTALL" = "true" ]; then
        echo "installing server"
        go install \
            -ldflags "-X rat/buildinfo.version=$(git describe --tags)" \
            -v ||
            exit 1
    fi

    cd "$_PWD" || exit 1
fi

if [ "$_DOCKER" = "true" ]; then
    echo "building docker image"

    RAT_VERSION="$_RAT_VERSION" \
        docker-compose build || exit 1
fi
