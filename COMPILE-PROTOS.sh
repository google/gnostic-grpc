#!/bin/sh
go get github.com/golang/protobuf@v1.4.2
protoc --go_out=incompatibility/incompatibility-report/ incompatibility/incompatibility-report/incompatibility-report.proto