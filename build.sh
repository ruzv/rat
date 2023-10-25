#!/bin/bash

INSTALL="false"
BUILD="false"
SERVER="false"
WEB="false"

while getopts 'ibsw' OPTION; do
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
  cd web || exit 1
  npm install || exit 1
  npm run build || exit 1
  cd .. || exit 1
fi

if [ "$SERVER" = "true" ]; then
  if [ "$BUILD" = "true" ]; then
    echo "building server"
    go build -ldflags "-X main.version=$(git describe --tags)" || exit 1
  fi

  if [ "$INSTALL" = "true" ]; then
    echo "installing server"
    go install -ldflags "-X main.version=$(git describe --tags)" || exit 1
  fi
fi
