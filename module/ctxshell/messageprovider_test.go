package ctxshell_test

import (
	"testing"

	"github.com/swaros/contxt/module/ctxshell"
)

func TestMessageProvider(t *testing.T) {

	mprovider := ctxshell.NewCshellMsgScope(5)
	mprovider.Push("stdout", "hello")
	mprovider.Push("stdout", "world")
	mprovider.Push("stdout", "!")

	for i := 0; i < 3; i++ {
		msg := mprovider.Pop()
		if msg.MsgType != "stdout" {
			t.Errorf("expected stdout, got %s", msg.MsgType)
		}
		if msg.Msg != "hello" && msg.Msg != "world" && msg.Msg != "!" {
			t.Errorf("expected hello, world or !, got %s", msg.Msg)
		}
	}

}
