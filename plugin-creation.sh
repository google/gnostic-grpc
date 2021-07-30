#!/bin/bash
gopath=$(go env GOPATH)
./COMPILE-PROTOS.sh
cd plugin
go build -o ${gopath}/bin/gnostic-grpc 