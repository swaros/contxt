package ctxshell

import "time"

type CshellMessage struct {
	MsgType string // "stdout", "stderr", "info", "error"
	Msg     string
	Time    time.Time
}

type CshellMsgSCope struct {
	FIFO chan *CshellMessage
}

func NewCshellMessage(msgType, msg string) *CshellMessage {
	return &CshellMessage{
		MsgType: msgType,
		Msg:     msg,
		Time:    time.Now(),
	}
}

func NewCshellMsgScope(size int) *CshellMsgSCope {
	return &CshellMsgSCope{
		FIFO: make(chan *CshellMessage, size),
	}
}

func (t *CshellMsgSCope) Push(msgType, msg string) {
	t.FIFO <- NewCshellMessage(msgType, msg)
}

func (t *CshellMsgSCope) Pop() *CshellMessage {
	return <-t.FIFO
}

func (t *CshellMsgSCope) Size() int {
	return len(t.FIFO)
}

func (t *CshellMsgSCope) Close() {
	close(t.FIFO)
}

func (t *CshellMsgSCope) Flush() {
	for {
		select {
		case <-t.FIFO:
		default:
			return
		}
	}
}

func (t *CshellMsgSCope) FlushAndClose() {
	t.Flush()
	t.Close()
}

func (t *CshellMsgSCope) GetAllMessages() []*CshellMessage {
	var msgs []*CshellMessage
	for {
		select {
		case msg := <-t.FIFO:
			msgs = append(msgs, msg)
		default:
			return msgs
		}
	}
}
