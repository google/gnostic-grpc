#!/bin/bash
gopath=$(go env GOPATH)
# ./COMPILE-PROTOS.sh
go build -o ${gopath}/bin/gnostic-grpc plugin