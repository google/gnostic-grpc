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

package bookstore

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/golang/glog"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var (
	bookstoreEndpoint = flag.String("bookstoreEndpoint", "localhost:50051", "endpoint of YourService")
)

func run() error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}
	err := RegisterBookstoreHandlerFromEndpoint(ctx, mux, *bookstoreEndpoint, opts)
	if err != nil {
		return err
	}

	fmt.Print("\nProxy listening on 8081\n")
	return http.ListenAndServe(":8081", mux)
}

func RunProxy() {
	flag.Parse()
	defer glog.Flush()
	if err := run(); err != nil {
		glog.Fatal(err)
	}
}
