#!/bin/bash
gopath=$(go env GOPATH)
./COMPILE-PROTOS.sh
go install .
