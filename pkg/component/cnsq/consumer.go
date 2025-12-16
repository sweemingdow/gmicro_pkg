package cnsq

import (
	"context"
	"github.com/nsqio/go-nsq"
	"github.com/sweemingdow/gmicro_pkg/pkg/mylog"
	"sync"
	"time"
)

type (
	ConsumerItem struct {
		Topic                 string
		Channel               string
		Concurrency           int
		MaxAttempts           uint16        // 消费返回error, 最大重试次数(重新入队)
		MsgTimeout            time.Duration // 消费超时时间
		RequeueDelayWhenRetry time.Duration // 消费失败(返回error), 重新入队延时

	}

	NsqCsConfig struct {
		NsqdDirectAddr    []string // [ip:port]
		NsqLookupdAddr    []string // [ip:port]
		HeartbeatInterval time.Duration
		Items             []ConsumerItem
	}

	NsqConsumer struct {
		cfg NsqCsConfig
		css []*nsq.Consumer
	}

	msgHandler struct {
		topic     string
		channel   string
		csFactory NsqMsgConsumeFactory
	}
)

func NewNsqConsumer(cfg NsqCsConfig, csFactory NsqMsgConsumeFactory) (*NsqConsumer, error) {
	if len(cfg.NsqdDirectAddr) == 0 && len(cfg.NsqLookupdAddr) == 0 {
		panic("addresses is required")
	}

	if len(cfg.Items) == 0 {
		panic("consumer items is required")
	}

	var newErr error
	css := make([]*nsq.Consumer, 0, len(cfg.Items))

	defer func() {
		if newErr != nil && len(css) > 0 {
			for _, cs := range css {
				cs.Stop()
			}
		}
	}()

	for _, item := range cfg.Items {
		csCfg := nsq.NewConfig()
		csCfg.MaxInFlight = item.Concurrency
		csCfg.MsgTimeout = item.MsgTimeout
		csCfg.MaxAttempts = item.MaxAttempts
		csCfg.HeartbeatInterval = cfg.HeartbeatInterval
		csCfg.DefaultRequeueDelay = item.RequeueDelayWhenRetry

		cs, err := nsq.NewConsumer(item.Topic, item.Channel, csCfg)
		if err != nil {
			newErr = err
			break
		}

		cs.AddHandler(msgHandler{
			topic:     item.Topic,
			channel:   item.Channel,
			csFactory: csFactory,
		})

		if len(cfg.NsqdDirectAddr) > 0 {
			if len(cfg.NsqdDirectAddr) > 1 {
				newErr = cs.ConnectToNSQDs(cfg.NsqdDirectAddr)
			} else {
				newErr = cs.ConnectToNSQD(cfg.NsqdDirectAddr[0])
			}
		} else {
			if len(cfg.NsqLookupdAddr) > 1 {
				newErr = cs.ConnectToNSQLookupds(cfg.NsqLookupdAddr)
			} else {
				newErr = cs.ConnectToNSQLookupd(cfg.NsqLookupdAddr[0])
			}
		}

		if newErr != nil {
			break
		}

		css = append(css, cs)
	}

	return &NsqConsumer{
		cfg: cfg,
		css: css,
	}, newErr
}

func (ncs *NsqConsumer) OnCreated(_ chan<- error) {
}

func (ncs *NsqConsumer) OnDispose(ctx context.Context) error {
	var wg sync.WaitGroup

	wg.Add(len(ncs.css))

	for _, cs := range ncs.css {
		cs := cs
		go func() {
			defer wg.Done()

			cs.Stop()
			<-cs.StopChan
		}()
	}

	allDone := make(chan struct{})
	go func() {
		wg.Wait()
		close(allDone)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-allDone:
		lg := mylog.AppLogger()
		lg.Info().Msg("nsq consumers stopped gracefully")
		return nil
	}
}

func (mh msgHandler) HandleMessage(m *nsq.Message) error {
	return mh.csFactory.MsgHandle(mh.topic, mh.channel, m)
}
