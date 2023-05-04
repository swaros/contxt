package ctxout

import "fmt"

type FmtWarper struct{}

func NewFmtWrap() *FmtWarper {
	return &FmtWarper{}
}

func (f *FmtWarper) Filter(msg interface{}) interface{} {
	return msg
}

func (f *FmtWarper) Stream(msg ...interface{}) {
	fmt.Print(msg...)
}

func (f *FmtWarper) StreamLn(msg ...interface{}) {
	fmt.Println(msg...)
}
