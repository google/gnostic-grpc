// Copyright 2019 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/googleapis/gnostic-grpc/examples/end-to-end-grpc-gateway/bookstore"
	"google.golang.org/grpc"
	"log"
	"time"
)

var (
	serverAddr = flag.String("server_addr", "127.0.0.1:50051", "The server address in the format of host:port")
)

func main() {
	flag.Parse()

	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	conn, err := grpc.Dial(*serverAddr, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()

	client := bookstore.NewBookstoreClient(conn)
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	res, err := client.ListShelves(ctx, &empty.Empty{})
	if res != nil {
		fmt.Println("The themes of your shelves:")
		for _, shelf := range res.Shelves {
			fmt.Println(shelf.Theme)
		}
	}
}
