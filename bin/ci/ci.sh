#!/usr/bin/env bash

set -e

# run one/many
for arg in "$@"
do
  if [[ $arg == 1 ]]; then
    echo "run user svc..."
    go build -o ./usersvc zim.cn/service/usersvc/cmd
    ./usersvc
  else
    echo unknown argument: $arg
  fi
done
