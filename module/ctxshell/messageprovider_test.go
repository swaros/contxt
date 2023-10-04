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

func TestMessageProviderFlush(t *testing.T) {

	mprovider := ctxshell.NewCshellMsgScope(5)
	mprovider.Push("stdout", "hello")
	mprovider.Push("stdout", "world")
	mprovider.Push("stdout", "!")

	mprovider.Flush()

	if mprovider.Size() != 0 {
		t.Errorf("expected size 0, got %d", mprovider.Size())
	}

}

func TestMessageProviderFlushAndClose(t *testing.T) {

	mprovider := ctxshell.NewCshellMsgScope(5)
	mprovider.Push("stdout", "hello")
	mprovider.Push("stdout", "world")
	mprovider.Push("stdout", "!")

	mprovider.FlushAndClose()

	if mprovider.Size() != 0 {
		t.Errorf("expected size 0, got %d", mprovider.Size())
	}

	if mprovider.Pop() != nil {
		t.Errorf("expected nil, got %v", mprovider.Pop())
	}

	mprovider.Push("stdout", "hello again")
	if mprovider.Size() != 0 {
		t.Errorf("expected size 0, got %d", mprovider.Size())
	}
}

func TestMessageProviderClose(t *testing.T) {

	mprovider := ctxshell.NewCshellMsgScope(5)
	mprovider.Push("stdout", "hello")
	mprovider.Push("stdout", "world")
	mprovider.Push("stdout", "!")

	mprovider.Close()

	if mprovider.Pop() != nil {
		t.Errorf("expected nil, got %v", mprovider.Pop())
	}
}

func TestMessageProviderSize(t *testing.T) {

	mprovider := ctxshell.NewCshellMsgScope(5)
	mprovider.Push("stdout", "hello")
	mprovider.Push("stdout", "world")
	mprovider.Push("stdout", "!")

	if mprovider.Size() != 3 {
		t.Errorf("expected size 3, got %d", mprovider.Size())
	}

}

func TestMessageGetAll(t *testing.T) {

	mprovider := ctxshell.NewCshellMsgScope(5)
	mprovider.Push("stdout", "hello")
	mprovider.Push("stdout", "world")
	mprovider.Push("stdout", "!")

	msgs := mprovider.GetAllMessages()

	if len(msgs) != 3 {
		t.Errorf("expected 3 messages, got %d", len(msgs))
	}

	if msgs[0].Msg != "hello" {
		t.Errorf("expected hello, got %s", msgs[0].Msg)
	}

	if msgs[1].Msg != "world" {
		t.Errorf("expected world, got %s", msgs[1].Msg)
	}

	if msgs[2].Msg != "!" {
		t.Errorf("expected !, got %s", msgs[2].Msg)
	}

}
