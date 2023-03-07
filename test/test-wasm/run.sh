#!/bin/bash

set -e

# run one/many
for arg in "$@"
do
  if [[ $arg == 1 ]]; then
    echo "build wasm..."
    GOOS=js GOARCH=wasm go build -o zim.wasm
    cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" .
  elif [[ $arg == 2 ]]; then
    echo "build wasm with tinygo..."
    tinygo build -o zim.wasm -target=wasm
    #cp $(tinygo env TINYGOROOT)/targets/wasm_exec.js .
  else
    echo unknown argument: $arg
  fi
done
