# End-to-examples
This directory contains an end-to-end flow for generating a gRPC API with HTTP transcoding from an OpenAPI description.

The tutorial uses the  [gRPC gateway plugin](https://github.com/grpc-ecosystem/grpc-gateway) for the proxy.

## End-to-end flow with gRPC gateway plugin

This example demonstrates an end-to-end flow for generating a gRPC API with HTTP transcoding from an
OpenAPI description.


#### What we will build:

![alt text](https://raw.githubusercontent.com/google/gnostic-grpc/master/examples/images/end-to-end-grpc-gateway.png "gRPC with Transcoding")

This tutorial has six steps:

1. Generate a gRPC service (.proto) from an OpenAPI description.
2. Generate server-side support code for the gRPC service.
3. Implement the server logic.
4. Set up a proxy that provides HTTP transcoding.
5. Run the proxy and the server.
6. Test your API with with curl and a gRPC client.

#### Prerequisite
Install [gnostic](https://github.com/google/gnostic), [gnostic-grpc](https://github.com/google/gnostic-grpc),
[go plugin for protoc](https://github.com/golang/protobuf/protoc-gen-go), [gRPC gateway plugin](https://github.com/grpc-ecosystem/grpc-gateway)
and [gRPC](https://grpc.io/)

    go get -u github.com/google/gnostic
    go get -u github.com/google/gnostic-grpc
    go get -u github.com/golang/protobuf/protoc-gen-go
    go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway
    go get -u google.golang.org/grpc
    
For simplicity lets create a temporary environment variable inside your terminal:
    
    export ANNOTATIONS="third-party/googleapis"
    
In order for this tutorial to work you should work inside this directory under `GOPATH`.

#### 1. Step

Use [gnostic](https://github.com/google/gnostic) to generate the Protocol buffer 
description (`bookstore.proto`) in the current directory:

    gnostic --grpc-out=. bookstore.yaml

#### 2. Step
Generate the gRPC stubs:
    
    protoc --proto_path=. --proto_path=${ANNOTATIONS} --go_out=plugins=grpc:bookstore bookstore.proto
    
 This generates `bookstore/bookstore.pb.go`.

#### 3. Step
We added an example implementation of the server using the generated gRPC stubs inside `bookstore/server.go`.
    
#### 4. Step
Generate the reverse proxy with the gRPC gateway plugin:

    protoc --proto_path=. --proto_path=${ANNOTATIONS} --grpc-gateway_out=bookstore bookstore.proto

This generates `bookstore/bookstore.pb.gw.go`.

We provided a sample implementation on how to use the proxy inside `bookstore/proxy.go`.

#### 5. Step
Start the proxy and the server:

    go run main.go
    
#### 6. Step

##### cURL
Inside of a new terminal test your API:

Let's create a shelf first:

    curl -X POST \
      http://localhost:8081/shelves \
      -H 'Content-Type: application/json' \
      -d '{
        "name": "Books I need to read",
        "theme": "Non-fiction"
    }'
    
Get all existing shelves:

    curl -X GET http://localhost:8081/shelves
    
Create a book for the shelve with the id `1`:
    
    curl -X POST \
      http://localhost:8081/shelves/1/books \
      -H 'Content-Type: application/json' \
      -d '{
        "author": "Hans Rosling",
        "name": "Factfulness",
        "title": "Factfulness: Ten Reasons We'\''re wrong about the world - and Why Things Are Better Than You Think"
    }'
    
    
List all books for the shelve with the id `1`:

    curl -X GET http://localhost:8081/shelves/1/books
    
    
##### gRPC client

A sample gRPC client is provided inside `grpc-client/client.go` that lists all themes of your shelves:

    go run grpc-client/client.go
