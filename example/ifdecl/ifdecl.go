package ifdecl

import "context"

type ArrayType [32]byte

type SliceType []int

type StructSlice []TestStruct

type TestInterface interface {
	A(ctx context.Context, a int) int
	B(b string, c int) int
	C(c bool) (bool, error)
	D() (TestStruct, error)
	E(ArrayType) error
	F([]byte) []byte
	G([32]int) [24]int
	H(SliceType) SliceType
	I(StructSlice) StructSlice
	J([]TestStruct) []TestStruct
}

type TestStruct struct {
	A int
	B string
}
