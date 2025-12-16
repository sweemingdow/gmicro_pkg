package cnsq

import (
	"github.com/nsqio/go-nsq"
)

type HandlerRegister interface {
	Register(topic string, handler nsq.Handler)

	Clear()
}

type NsqMsgConsumeFactory interface {
	HandlerRegister

	MsgHandle(topic, channel string, m *nsq.Message) error
}

func NewStaticNsqMsgConsumeFactory() NsqMsgConsumeFactory {
	return &staticNsqMsgConsumeFactory{
		m: make(map[string]nsq.Handler),
	}
}

type staticNsqMsgConsumeFactory struct {
	m map[string]nsq.Handler
}

func (nf *staticNsqMsgConsumeFactory) Register(topic string, handler nsq.Handler) {
	nf.m[topic] = handler
}

func (nf *staticNsqMsgConsumeFactory) Clear() {
	clear(nf.m)
	nf.m = nil
}

func (nf *staticNsqMsgConsumeFactory) MsgHandle(topic, _ string, m *nsq.Message) error {
	if h, ok := nf.m[topic]; ok {
		return h.HandleMessage(m)
	} else {
		panic("no handler found with topic:" + topic)
	}
}
