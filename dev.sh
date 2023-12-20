#!/bin/bash

cd src || exit

/Users/rzvejs/go/bin/gow run . -c ../test-config.yaml

# cd web || exit
#
# npm run dev
