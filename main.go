package main

import "github.com/karim42benhammou/gnostic-grpc/plugin"

//go:generate sh COMPILE-PROTOS.sh

func main() {
	plugin.Main()
}
