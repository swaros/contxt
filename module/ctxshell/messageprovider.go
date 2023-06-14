// MIT License
//
// Copyright (c) 2020 Thomas Ziegler <thomas.zglr@googlemail.com>. All rights reserved.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the Software), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED AS IS, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// AINC-NOTE-0815

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
