#!/bin/sh

TMP_GOBIN=$(mktemp -d -t gnostic-grpc.XXXXXXXXXX)

GOBIN=${TMP_GOBIN} go install ./search
GOBIN=${TMP_GOBIN} go install -mod=mod google.golang.org/protobuf/cmd/protoc-gen-go@v1.29.1
PATH="$TMP_GOBIN:$PATH" protoc --go_out=incompatibility/ ./incompatibility/incompatibility-report.proto

rm -rf "${TMP_GOBIN}"
