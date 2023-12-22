#!/bin/bash

INSTALL="false"
BUILD="false"
SERVER="false"
WEB="false"

_SERVER_DIR="src"
_WEB_DIR="src/web"
_PWD=$(pwd)
_OUTPUT_BIN="$_PWD/rat"

while getopts 'ibswh' OPTION; do
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
  "h")
    cat <<EOF
usage: $0 [flags...]
  -i build and install server binary to GOPATH
  -b build server binary to current directory
  -s build only server
  -w build only web
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

## default
if [ "$SERVER" = "false" ] && [ "$WEB" = "false" ]; then
  SERVER="true"
  WEB="true"
fi

if [ "$INSTALL" = "false" ] && [ "$BUILD" = "false" ]; then
  INSTALL="true"
fi

echo "INSTALL: $INSTALL"
echo "BUILD: $BUILD"
echo "SERVER: $SERVER"
echo "WEB: $WEB"

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
      -ldflags "-X main.version=$(git describe --tags)" \
      -o "$_OUTPUT_BIN" || exit 1
  fi

  if [ "$INSTALL" = "true" ]; then
    echo "installing server"
    go install \
      -v \
      -ldflags "-X main.version=$(git describe --tags)" || exit 1
  fi

  cd "$_PWD" || exit 1
fi
