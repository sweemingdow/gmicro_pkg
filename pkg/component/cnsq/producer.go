package cnsq

import (
	"context"
	"github.com/nsqio/go-nsq"
	"github.com/sweemingdow/gmicro_pkg/pkg/mylog"
	"github.com/sweemingdow/gmicro_pkg/pkg/parser/json"
)

const ProducerLifetimeTag = "nsq_producer"

type (
	NsqPdConfig struct {
		NsqdAddr string // ip:port
	}

	NsqProducer struct {
		pd *nsq.Producer
	}
)

func NewNsqProducer(cfg NsqPdConfig) (*NsqProducer, error) {
	pdCfg := nsq.NewConfig()
	pd, err := nsq.NewProducer(cfg.NsqdAddr, pdCfg)
	if err != nil {
		return nil, err
	}

	pd.SetLogger(newProduceAdaptLogger(), nsq.LogLevelInfo)

	return &NsqProducer{
		pd: pd,
	}, nil
}

type PublishParam struct {
	Topic   string
	Payload []byte
}

func (npd *NsqProducer) Publish(pp PublishParam) error {
	if err := npd.pd.Publish(pp.Topic, pp.Payload); err != nil {
		return err
	}

	return nil
}

func (npd *NsqProducer) JsonPublish(topic string, val any) error {
	data, err := json.Fmt(val)
	if err != nil {
		return err
	}
	if err = npd.pd.Publish(topic, data); err != nil {
		return err
	}

	return nil
}

func (npd *NsqProducer) PublishAsync(pp PublishParam, doneChan chan *nsq.ProducerTransaction, args []any) error {
	if err := npd.pd.PublishAsync(pp.Topic, pp.Payload, doneChan, args...); err != nil {
		return err
	}

	return nil
}

func (npd *NsqProducer) OnCreated(_ chan<- error) {
}

func (npd *NsqProducer) OnDispose(ctx context.Context) error {
	stopped := make(chan struct{})

	go func() {
		defer close(stopped)
		npd.pd.Stop()
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-stopped:
		lg := mylog.GetStopMarkLogger()
		lg.Info().Msg("nsq producer stopped successfully")
		return nil
	}
}
