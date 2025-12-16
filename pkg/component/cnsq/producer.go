package cnsq

import (
	"context"
	"github.com/nsqio/go-nsq"
	"github.com/sweemingdow/gmicro_pkg/pkg/mylog"
)

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

	return &NsqProducer{
		pd: pd,
	}, nil
}

type PublishParam struct {
	Topic   string
	Payload []byte
}

func (npd *NsqProducer) Publish(pp PublishParam) error {
	err := npd.pd.Publish(pp.Topic, pp.Payload)
	if err != nil {
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
		lg := mylog.AppLogger()
		lg.Info().Msg("nsq producer stopped gracefully")
		return nil
	}
}
