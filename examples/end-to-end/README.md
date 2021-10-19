# End-to-examples
This directory contains an end-to-end flow for generating a gRPC API with HTTP transcoding from an OpenAPI description.

## End-to-end flow with [envoy](https://www.envoyproxy.io/)

This example demonstrates an end-to-end flow for generating a gRPC API with HTTP transcoding from an
OpenAPI description.


#### What we will build:
![alt text](https://raw.githubusercontent.com/google/gnostic-grpc/master/examples/images/end-to-end-envoy.png "gRPC with Transcoding")

#### Prerequisite
Install [gnostic](https://github.com/google/gnostic), [gnostic-grpc](https://github.com/google/gnostic-grpc),
[go plugin for protoc](https://github.com/golang/protobuf/protoc-gen-go), and [gRPC](https://grpc.io/).

    go get -u github.com/google/gnostic
    go get -u github.com/google/gnostic-grpc
    go get -u github.com/golang/protobuf/protoc-gen-go
    go get -u google.golang.org/grpc
    
For simplicity lets create a temporary environment variable inside your terminal:
    
    export ANNOTATIONS="third-party/googleapis"
    
In order for this tutorial to work you should work inside this directory under `GOPATH`.

#### 1. Step: Generate a gRPC service (.proto) from an OpenAPI description

Use [gnostic](https://github.com/google/gnostic) to generate the Protocol buffer 
description (`bookstore.proto`) in the current directory:

    gnostic --grpc-out=. bookstore.yaml

#### 2. Step: Generate server-side support code for the gRPC service
Generate the gRPC stubs:
    
    protoc --proto_path=. --proto_path=${ANNOTATIONS} --go_out=plugins=grpc:bookstore bookstore.proto
    
 This generates `bookstore/bookstore.pb.go`.

#### 3. Step: Implement the server logic
We added an example implementation of the server using the generated gRPC stubs inside `bookstore/server.go`.

#### 4. Step: Generate the descriptor set for envoy
Given `bookstore.proto` generate the descriptor set.
    
    protoc --proto_path=${ANNOTATIONS} --proto_path=. --include_imports --include_source_info \
    --descriptor_set_out=envoy-proxy/proto.pb bookstore.proto
    
This generates `envoy-proxy/proto.pb`.

#### 5. Step: Set up an envoy proxy
The file `envoy-proxy/envoy.yaml` contains an envoy configuration with a gRPC-JSON [transcoder](https://www.envoyproxy.io/docs/envoy/latest/configuration/http_filters/grpc_json_transcoder_filter).
According to the configuration, port 51051 proxies gRPC requests to a gRPC server running on localhost:50051 and uses 
the gRPC-JSON transcoder filter to provide the RESTful JSON mapping. I.e.: you can either make gRPC or RESTful JSON 
requests to localhost:51051.
  
Get the envoy docker image:

    docker pull envoyproxy/envoy-dev:5d95032baa803f853e9120048b56c8be3dab4b0d  
  
The file `envoy-proxy/Dockerfile` uses the envoy image we just pulled as base image and copies `envoy.yaml`
and `proto.pb` to the filesystem of the docker container.  

Build a docker image:

    docker build -t envoy:v1 envoy-proxy
    
Run the docker container with the created image on port 51051:

    docker run -d --name envoy -p 9901:9901 -p 51051:51051 envoy:v1
    
#### 6. Step: Run the gRPC server
Run the gRPC server on port 50051:

    go run main.go

#### 7. Step: Test your API

##### cURL
Inside of a new terminal test your API:

Let's create a shelf first:

    curl -X POST \
      http://localhost:51051/shelves \
      -H 'Content-Type: application/json' \
      -d '{
        "name": "Books I need to read",
        "theme": "Non-fiction"
    }'
    
Get all existing shelves:

    curl -X GET http://localhost:51051/shelves
    
Create a book for the shelve with the id `1`:
    
    curl -X POST \
      http://localhost:51051/shelves/1/books \
      -H 'Content-Type: application/json' \
      -d '{
        "author": "Hans Rosling",
        "name": "Factfulness",
        "title": "Factfulness: Ten Reasons We'\''re wrong about the world - and Why Things Are Better Than You Think"
    }'
    
    
List all books for the shelve with the id `1`:

    curl -X GET http://localhost:51051/shelves/1/books
    
    
##### gRPC client

A sample gRPC client is provided inside `grpc-client/client.go` that lists all themes of your shelves:

    go run grpc-client/client.go
