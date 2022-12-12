//go:generate go run ./generate/generate.go

package main

import (
	"context"
	"fmt"
	"net"

	"github.com/secDre4mer/gen-grpc-go/example/ifdecl"
	"github.com/secDre4mer/gen-grpc-go/example/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	server := grpc.NewServer()
	listener, err := net.Listen("tcp", "127.0.0.1:12345")
	if err != nil {
		panic(err)
	}
	proto.RegisterTestInterface(server, TestInterfaceImpl{})
	go server.Serve(listener)

	grpcClient, err := grpc.Dial("127.0.0.1:12345", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	client := proto.MakeTestInterfaceClient(grpcClient)

	fmt.Println(client.A(context.Background(), 42))
	fmt.Println(client.B("test", 42))
	fmt.Println(client.C(true))
	fmt.Println(client.D())
	fmt.Println(client.E(ifdecl.ArrayType{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}))
	fmt.Println(client.F([]byte{0, 1, 2, 3, 4, 5, 6}))
	fmt.Println(client.G([32]int{0, 1, 2, 3, 4, 5, 6}))
	fmt.Println(client.H([]int{0, 1, 2, 3, 4, 5, 6}))
	fmt.Println(client.I([]ifdecl.TestStruct{
		{42, "test"},
		{1, "test2"},
	}))
	fmt.Println(client.J([]ifdecl.TestStruct{
		{42, "test"},
		{1, "test2"},
	}))
}

type TestInterfaceImpl struct {
}

func (t TestInterfaceImpl) A(ctx context.Context, a int) int {
	return a
}

func (t TestInterfaceImpl) B(b string, c int) int {
	fmt.Println(b)
	return c
}
func (t TestInterfaceImpl) C(c bool) (bool, error) {
	return c, nil
}
func (t TestInterfaceImpl) D() (ifdecl.TestStruct, error) {
	return ifdecl.TestStruct{
		A: 42,
		B: "test",
	}, nil
}
func (t TestInterfaceImpl) E(a ifdecl.ArrayType) error {
	fmt.Println(a)
	return nil
}
func (t TestInterfaceImpl) F(bytes []byte) []byte {
	return bytes
}
func (t TestInterfaceImpl) G(bytes [32]int) [24]int {
	return *(*[24]int)(bytes[:24])
}
func (t TestInterfaceImpl) H(s ifdecl.SliceType) ifdecl.SliceType {
	return s
}
func (t TestInterfaceImpl) I(s ifdecl.StructSlice) ifdecl.StructSlice {
	return s
}
func (t TestInterfaceImpl) J(s []ifdecl.TestStruct) []ifdecl.TestStruct {
	return s
}
