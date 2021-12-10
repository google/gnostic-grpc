#!/bin/bash
gobin=$(go env GOBIN)
./COMPILE-PROTOS.sh
go build -o ${gobin}/gnostic-grpc