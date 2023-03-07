#!/usr/bin/env bash

set -e

# INSTALL
# go install github.com/gogo/protobuf/protoc-gen-gofast@latest
# go install github.com/gogo/protobuf/protoc-gen-gogofaster

#protoc --go_out=. *.proto
#protoc --gofast_out=. *.proto
protoc --gogofaster_out=. *.proto