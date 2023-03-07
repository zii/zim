#!/usr/bin/env bash

set -e

# 生成swagger文档
swag init -d ./,../biz -g swag.go -p snakecase
