package testdata

import (
	"fmt"
)

type testStruct struct {
	A string
}

func (ts *testStruct) BBB() {

}

func (ts *testStruct) CCC() (int, int) {
	return 0, 2
}

func test() {
	var a = testStruct{}
	fmt.Println(a)
	b, c := a.CCC()
	fmt.Println(a.A, b, c)
	a.BBB()
	var kkk interface{} = a
	kkk.(*testStruct).BBB()
	test()
}
