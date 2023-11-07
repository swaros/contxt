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

type Topic string

const (
	TopicDefault Topic = "default"
	TopicInfo    Topic = "info"
	TopicError   Topic = "error"
)

type Msg struct {
	msg           string
	topic         Topic
	timeToDisplay time.Duration
	formatFunc    func(string) string
}

func NewPromptMessage(msg string, topic Topic, timeToDisplay time.Duration, formatFunc func(string) string) Msg {
	return Msg{
		msg:           msg,
		topic:         topic,
		timeToDisplay: timeToDisplay,
		formatFunc:    formatFunc,
	}
}

func DefaultPromptMessage(msg string, topic Topic, timeToDisplay time.Duration) Msg {
	return Msg{
		msg:           msg,
		timeToDisplay: timeToDisplay,
		topic:         topic,
		formatFunc:    func(s string) string { return s },
	}
}

func SimpleMsgOneSecond(msg string, topic Topic) *Msg {
	return &Msg{
		msg:           msg,
		timeToDisplay: time.Second,
		topic:         topic,
		formatFunc:    func(s string) string { return s },
	}
}

func (m *Msg) GetMsg() string {
	return m.formatFunc(m.msg)
}

func (m *Msg) GetTopic() Topic {
	return m.topic
}

func (m *Msg) GetTimeToDisplay() time.Duration {
	return m.timeToDisplay
}

func (m *Msg) SetFormatFunc() *Msg {
	m.formatFunc = func(s string) string { return s }
	return m
}
