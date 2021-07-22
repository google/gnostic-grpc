#!/bin/sh
go get github.com/golang/protobuf@v1.4.2
go get
go install github.com/golang/protobuf/protoc-gen-go
protoc --go_out=incompatibility/incompatibility-report/ incompatibility/incompatibility-report/incompatibility-report.proto