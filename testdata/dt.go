package testdata

import (
	"fmt"
)

type testStruct struct {
	A string
}

func (ts *testStruct) BBB() {

}

func test() {
	var a = testStruct{}
	fmt.Println(a)
	fmt.Println(a.A)
	a.BBB()
	var kkk interface{} = a
	kkk.(*testStruct).BBB()
	test()
}
