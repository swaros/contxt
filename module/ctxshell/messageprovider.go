package ctxshell

import "time"

type CshellMessage struct {
	MsgType string // "stdout", "stderr", "info", "error"
	Msg     string
	Time    time.Time
}

type CshellMsgFifo struct {
	fifo   chan *CshellMessage
	closed bool
}

func NewCshellMessage(msgType, msg string) *CshellMessage {
	return &CshellMessage{
		MsgType: msgType,
		Msg:     msg,
		Time:    time.Now(),
	}
}

func NewCshellMsgScope(size int) *CshellMsgFifo {
	return &CshellMsgFifo{
		fifo: make(chan *CshellMessage, size),
	}
}

func (t *CshellMsgFifo) Push(msgType, msg string) {
	if t.closed {
		return
	}
	t.fifo <- NewCshellMessage(msgType, msg)
}

func (t *CshellMsgFifo) Pop() *CshellMessage {
	if t.closed {
		return nil
	}
	return <-t.fifo
}

func (t *CshellMsgFifo) Size() int {
	return len(t.fifo)
}

func (t *CshellMsgFifo) Close() {
	t.closed = true
	close(t.fifo)
}

func (t *CshellMsgFifo) Flush() {
	if t.closed {
		return
	}
	for {
		select {
		case <-t.fifo:
		default:
			return
		}
	}
}

func (t *CshellMsgFifo) FlushAndClose() {
	t.Flush()
	t.Close()
}

func (t *CshellMsgFifo) GetAllMessages() []*CshellMessage {
	if t.closed {
		return nil
	}
	var msgs []*CshellMessage
	for {
		select {
		case msg := <-t.fifo:
			msgs = append(msgs, msg)
		default:
			return msgs
		}
	}
}
