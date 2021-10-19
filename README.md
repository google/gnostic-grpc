[![Go Actions Status](https://github.com/google/gnostic-grpc/workflows/Go/badge.svg)](https://github.com/google/gnostic-grpc/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/google/gnostic-grpc)](https://goreportcard.com/report/github.com/google/gnostic-grpc)
[![Test Coverage](https://codecov.io/gh/google/gnostic-grpc/branch/master/graph/badge.svg)](https://codecov.io/gh/google)

# gnostic gRPC plugin

This plugin has two functionalities:
- Conversion of an OpenAPI v3.0 API description into a description of a gRPC
service that can be used to implement that API using [gRPC-JSON Transcoding](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/grpc_json_transcoder_filter). 
gRPC services are described using the [Protocol Buffers](https://developers.google.com/protocol-buffers/) language.
Continue reading this README for more information.
- Scanning OpenAPI v3 documents for equivalent gRPC service incompatibilities. For simple per file incompatibility scanning see the README file in the [directory](https://github.com/google/gnostic-grpc/tree/master/incompatibility). For analysis over a set of OpenAPI documents see the README file in the [directory](https://github.com/google/gnostic-grpc/tree/master/analysis).

OpenAPI descriptions are read and processed with
[gnostic](https://github.com/google/gnostic), and this tool runs as a
gnostic plugin.

## High level overview:
![High Level Overview](https://raw.githubusercontent.com/google/gnostic-grpc/master/examples/images/high-level-overview.png "High Level Overview")

Under the hood the plugin first creates a FileDescriptorSet (`bookststore.descr`) from the input
data. Then [protoreflect](https://github.com/jhump/protoreflect/) is used to print the output file. 

## How to use: 
Install gnostic and the plugin before Go 1.17:

    go get -u github.com/google/gnostic
    go get -u github.com/google/gnostic-grpc
    

with Go >= 1.17:

    go install github.com/google/gnostic@latest
    go install github.com/google/gnostic-grpc@latest

Run gnostic with the plugin:

    gnostic --grpc-out=examples/bookstore examples/bookstore/bookstore.yaml

This generates the gRPC service definition `examples/bookstore/bookstore.proto`.

## End-to-end example
This [directory](https://github.com/google/gnostic-grpc/tree/master/examples/end-to-end) contains a tutorial on how to build a gRPC service that implements an OpenAPI specification.

## What conversions are currently supported?

Given an [OpenAPI object](https://swagger.io/specification/#oasObject) following fields will be represented inside a
 .proto file:

| Object        | Fields        | Supported  |
| ------------- |:-------------:| -----:|
| OpenAPI object|               |       |
|               | openapi       |    No |
|               | info          |    No |
|               | servers       |    No |
|               | paths         |   Yes |
|               | components    |   Yes |
|               | security      |    No |
|               | tags          |    No |
|               | externalDocs  |    No |


## Disclaimer

This is prerelease software and work in progress. Feedback and
contributions are welcome, but we currently make no guarantees of
function or stability.

## Requirements

**gnostic-grpc** can be run in any environment that supports [Go](http://golang.org)
and the [Google Protocol Buffer Compiler](https://github.com/google/protobuf).

## Copyright

Copyright 2019, Google Inc.

## License

Released under the Apache 2.0 license.

