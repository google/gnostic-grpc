#!/bin/bash
gopath=$(go env GOPATH)
go build -o ${gopath}/bin/gnostic-grpc-compatibility