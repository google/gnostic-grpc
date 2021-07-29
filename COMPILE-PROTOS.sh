#!/bin/sh
# go install github.com/golang/protobuf/protoc-gen-go
protoc --go_out=incompatibility/ incompatibility/incompatibility-report.proto
# go get github.com/golang/protobuf@v1.4.2
go get ./...
