package ifdecl

import (
	"context"
	"errors"
)

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
	K(marshalable BinaryMarshalable) BinaryMarshalable
}

type TestStruct struct {
	A int
	B string
}

type BinaryMarshalable byte

func (b BinaryMarshalable) MarshalBinary() (data []byte, err error) {
	return []byte{byte(b)}, nil
}
func (b *BinaryMarshalable) UnmarshalBinary(data []byte) error {
	if len(data) != 1 {
		return errors.New("bad data length")
	}
	*b = BinaryMarshalable(data[0])
	return nil
}
