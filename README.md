[![Build Status](https://travis-ci.com/googleapis/gnostic-grpc.svg?branch=master)](https://travis-ci.com/googleapis/gnostic-grpc)

# gnostic gRPC plugin
[GSoC 2019 project](https://summerofcode.withgoogle.com/projects/#5244822191865856)

This tool converts an OpenAPI v3.0 API description into a description of a gRPC
service that can be used to implement that API using [gRPC-JSON Transcoding](https://www.envoyproxy.io/docs/envoy/latest/configuration/http_filters/grpc_json_transcoder_filter). gRPC services are described using the [Protocol Buffers](https://developers.google.com/protocol-buffers/) language.

OpenAPI descriptions are read and processed with
[gnostic](https://github.com/googleapis/gnostic), and this tool runs as a
gnostic plugin.

## High level overview:
![alt text](https://drive.google.com/uc?export=view&id=1tqDvZLiXK40ISK_LgINQGsno9-MymRQP "High Level Overview")

Under the hood the plugin first creates a FileDescriptorSet (`bookststore.descr`) from the input
data. Then [protoreflect](https://github.com/jhump/protoreflect/) is used to print the output file. 

## How to use:    
Install gnostic and the plugin:
    
    go get -u github.com/googleapis/gnostic
    go get -u github.com/googleapis/gnostic-grpc

Run gnostic with the plugin:

    gnostic --grpc-out=examples/bookstore examples/bookstore/bookstore.yaml

This generates the gRPC service definition `examples/bookstore/bookstore.proto`.

## End-to-end example
This [directory](https://github.com/googleapis/gnostic-grpc/tree/master/examples/end-to-end) contains a tutorial on how to build a gRPC service that implements an OpenAPI specification.

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

