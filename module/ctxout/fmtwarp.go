package ctxout

import "fmt"

type FmtWarper struct{}

func (f *FmtWarper) Filter(msg interface{}) interface{} {
	return msg
}

func (f *FmtWarper) Stream(msg ...interface{}) {
	fmt.Println(msg...)
}

func (f *FmtWarper) StreamLn(msg ...interface{}) {
	fmt.Print(msg...)
}
