package main

/*
INSTALL protobuffer:
brew install protobuf
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

EDIT proto:
vim api.proto
syntax = "proto3";
package api;

import "google/protobuf/timestamp.proto";
option go_package = "zim.cn/pb";

message Person {
  string name = 1;
  int32 id = 2;  // Unique ID number for this person.
  string email = 3;
}

BUILD:
protoc --go_out=. *.proto
*/
