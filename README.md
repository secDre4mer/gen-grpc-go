# Overview

gen-grpc-go is a library to generate GRPC protocol files and stubs from Golang interfaces.

This allows users to define Golang interfaces as the common API between multiple processes. This library
generates a client implementation and a server wrapper around the interface.

# Example

Start with generating the client and server sides for your interface:

```go
package main

import (
    gengrpc "github.com/secDre4mer/gen-grpc-go"
    "github.com/secDre4mer/gen-grpc-go/example/ifdecl"
)

func main() {
    var interfaceVar ifdecl.TestInterface
    gengrpc.GenerateGRPCForInterface(&interfaceVar, "./proto")
}
```

Then instantiate a client and a server:

```go
package main

import (
	"net"

	"github.com/secDre4mer/gen-grpc-go/example/ifdecl"
	"github.com/secDre4mer/gen-grpc-go/example/proto"
	"google.golang.org/grpc"
)

func main() {
	server := grpc.NewServer()
	listener, err := net.Listen("tcp", "127.0.0.1:12345")
	if err != nil {
		panic(err)
	}
	proto.RegisterTestInterface(server, TestInterfaceImpl{})
	go server.Serve(listener)
}
```

```go

package main

import (
	"github.com/secDre4mer/gen-grpc-go/example/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	grpcClient, err := grpc.Dial("127.0.0.1:12345", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	client := proto.MakeTestInterfaceClient(grpcClient)

	// Start calling methods on the client
}
```