package ctxout

import "github.com/swaros/manout"

type MOWrap struct {
	basic     *BasicColors
	ctxBehave CtxOutBehavior
}

func NewMOWrap() *MOWrap {
	return &MOWrap{
		basic: NewBasicColors(),
	}
}

func (m *MOWrap) Filter(msg interface{}) interface{} {
	return manout.Message(msg)
}

func (m *MOWrap) Update(info CtxOutBehavior) {
	m.ctxBehave = info
	manout.ColorEnabled = !info.NoColored
}

func (m *MOWrap) GetInfo() CtxOutBehavior {
	return m.ctxBehave
}
