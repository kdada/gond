package testdata

import (
	"fmt"
)

type testStruct struct {
	A string
}

func test() {
	var a = testStruct{}
	fmt.Println(a)
}
