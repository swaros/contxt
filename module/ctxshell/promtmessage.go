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

func (m *Msg) GetTimeToDisplay() time.Duration {
	return m.timeToDisplay
}

func (m *Msg) SetFormatFunc() *Msg {
	m.formatFunc = func(s string) string { return s }
	return m
}
