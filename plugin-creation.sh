#!/bin/bash
gobin=$(go env GOBIN)
./COMPILE-PROTOS.sh
go build
mv gnostic-grpc $gobin
