#!/bin/bash
gopath=$(go env GOPATH)
./COMPILE-PROTOS.sh
cd plugin
go install
