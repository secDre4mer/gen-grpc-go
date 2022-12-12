package main

import (
	gengrpc "github.com/secDre4mer/gen-grpc-go"
	"github.com/secDre4mer/gen-grpc-go/example/ifdecl"
)

func main() {
	var interfaceVar ifdecl.TestInterface
	gengrpc.GenerateGRPCForInterface(&interfaceVar, "./proto")
}
