#!/usr/bin/env bash

set -e

# run one/many
for arg in "$@"
do
  if [[ $arg == 1 ]]; then
    echo "run zimapi.."
    cd ../../service
    swag init -d ./,../biz -g swag.go -p snakecase
    cd -
    go build -o ./zimapi zim.cn/service/apisvc/cmd
    ./zimapi -config config.toml
  elif [[ $arg == 2 ]]; then
    echo "run zimapi on linux.."
    git pull
    cd ../../service
    swag init -d ./,../biz -g swag.go -p snakecase
    cd -
    go build -o ./zimapi zim.cn/service/apisvc/cmd
    supervisorctl restart zimapi
    tail -f /var/log/zim/api.log
  elif [[ $arg == 3 ]]; then
    echo "run zimbroker.."
    go build -o ./zimbroker zim.cn/service/brokersvc/cmd
    ./zimbroker -config config.toml
  elif [[ $arg == 4 ]]; then
    echo "run zimbroker on linux.."
    git pull
    go build -o ./zimbroker zim.cn/service/brokersvc/cmd
    supervisorctl restart zimbroker
    tail -f /var/log/zim/broker.log
  elif [[ $arg == 5 ]]; then
    echo "run zimcron svc..."
    go build -o ./zimcron zim.cn/service/cronsvc/cmd
    ./zimcron -config config_86.toml -dump cronjob.dump
  elif [[ $arg == 6 ]]; then
    echo "run zimcron on linux"
    git pull
    go build -o ./zimcron zim.cn/service/cronsvc/cmd
    supervisorctl restart zimcron
    tail -f /var/log/zim/cron.log
  elif [[ $arg == api ]]; then
    echo "build api.md"
    go build -o ./tl zim.cn/tool/tl-schema
    ./tl md -o ../../doc/API.md ../../doc/API.txt
  else
    echo unknown argument: $arg
  fi
done
