#!/bin/bash

/Users/rzvejs/go/bin/gow run . -c test-config.yaml &

cd web || exit

npm run dev
